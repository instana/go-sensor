// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instasarama

import (
	"bytes"

	"github.com/Shopify/sarama"
	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
)

// SyncProducer is a wrapper for sarama.SyncProducer that instruments its calls using
// provided instana.Sensor
type SyncProducer struct {
	sarama.SyncProducer
	sensor         *instana.Sensor
	propageContext bool
}

// NewSyncProducer creates a new SyncProducer using the given broker addresses and configuration, and
// instruments its calls
func NewSyncProducer(addrs []string, config *sarama.Config, sensor *instana.Sensor) (sarama.SyncProducer, error) {
	sp, err := sarama.NewSyncProducer(addrs, config)
	if err != nil {
		return sp, err
	}

	return WrapSyncProducer(sp, config, sensor), nil
}

// NewSyncProducerFromClient creates a new SyncProducer using the given client, and instruments its calls
func NewSyncProducerFromClient(client sarama.Client, sensor *instana.Sensor) (sarama.SyncProducer, error) {
	sp, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		return sp, err
	}

	return WrapSyncProducer(sp, client.Config(), sensor), nil
}

// WrapSyncProducer wraps an existing sarama.SyncProducer instance and instruments its calls. It requires the same
// config that was used to create this producer to detect the Kafka version and whether it's supposed to return
// successes/errors. To initialize a new sync producer instance use instasarama.NewSyncProducer() and
// instasarama.NewSyncProducerFromClient() convenience methods instead
func WrapSyncProducer(sp sarama.SyncProducer, config *sarama.Config, sensor *instana.Sensor) *SyncProducer {
	return &SyncProducer{
		SyncProducer:   sp,
		sensor:         sensor,
		propageContext: contextPropagationSupported(config.Version),
	}
}

// SendMessage picks up the trace context previously added to the message with
// instasarama.ProducerMessageWithSpan(), starts a new child span and injects its
// context into the message headers before sending it to the underlying producer.
// The call will not be traced if there the message does not contain trace context
func (p *SyncProducer) SendMessage(msg *sarama.ProducerMessage) (int32, int64, error) {
	sp := startProducerSpan(p.sensor, msg)
	if sp != nil {
		defer sp.Finish()

		carrier := ProducerMessageCarrier{msg}
		if p.propageContext {
			// forward the trace context, updating the span ids
			sp.Tracer().Inject(sp.Context(), ot.TextMap, carrier)
		} else {
			// remove previously added trace headers
			carrier.RemoveAll()
		}
	}

	partition, offset, err := p.SyncProducer.SendMessage(msg)
	if err != nil && sp != nil {
		sp.SetTag("kafka.error", err)
		sp.LogFields(otlog.Error(err))
	}

	return partition, offset, err
}

// SendMessages starts a new span and injects its context into messages headers before
// sending them with the underlying producer.
//
// This method attempts to use the existing trace context found in message headers.
// There will be NO SPAN CREATED for this call if messages originate from different trace contexts.
// A possible use case that result in such behavior would be if messages resulted from different
// HTTP requests are buffered and later being sent in one batch asynchronously.
// In case you want your batch publish operation to be a part of a specific trace, make sure that
// you inject the parent span of this trace explicitly before calling `SendMessages()`, i.e.
//
//	type MessageCollector struct {
//		CollectedMessages []*sarama.ProducerMessage
//		producer *instasarama.SyncProducer
//		// ...
//	}
//
//	func (c MessageCollector) Flush(ctx context.Context) error {
//		// extract the parent span from context and use it to continue the trace
//		if parentSpan, ok := instana.SpanFromContext(ctx); ok {
//			// start a new span for the batch send job
//			sp := parentSpan.Tracer().StartSpan("batch-send", ot.ChilfOf(parentSpan.Context()))
//			defer sp.Finish()
//
//			// inject the trace context into every collected message, overriding the existing one
//			for i, msg := range c.CollectedMessages {
//				c.CollectedMessages = instasarama.ProducerMessageWithSpan(msg, sp)
//			}
//		}
//
//		return c.producer.SendMessages(c.CollectedMessages)
//	}
func (p *SyncProducer) SendMessages(msgs []*sarama.ProducerMessage) error {
	if len(msgs) == 0 {
		return nil
	}

	var sp ot.Span
	if producerMessagesFromSameContext(msgs) {
		sp = startProducerSpan(p.sensor, msgs[0])
	}

	if sp != nil {
		defer sp.Finish()

		instana.BatchSize(len(msgs)).Set(sp)

		// collect unique topics from the rest of messages and inject trace context in one go
		topics := make(map[string]struct{})
		for _, msg := range msgs {
			if _, ok := topics[msg.Topic]; !ok {
				topics[msg.Topic] = struct{}{}
			}

			// forward the trace context, updating the span id
			sp.Tracer().Inject(sp.Context(), ot.TextMap, ProducerMessageCarrier{msg})
		}

		// send topics as a comma-separated string
		buf := bytes.NewBuffer(nil)
		for topic := range topics {
			buf.WriteString(topic)
			buf.WriteByte(',')
		}
		buf.Truncate(buf.Len() - 1) // truncate trailing comma
		sp.SetTag("kafka.service", buf.String())
	}

	err := p.SyncProducer.SendMessages(msgs)
	if err != nil && sp != nil {
		sp.SetTag("kafka.error", err)
		sp.LogFields(otlog.Error(err))
	}

	return err
}

// startSpan picks up the existing trace context provided in the message and returns a new child
// span. It returns nil if there is no valid context provided in the message
func startProducerSpan(sensor *instana.Sensor, msg *sarama.ProducerMessage) ot.Span {
	switch sc, err := sensor.Tracer().Extract(ot.TextMap, ProducerMessageCarrier{msg}); err {
	case nil:
		return sensor.Tracer().StartSpan(
			"kafka",
			ext.SpanKindProducer,
			ot.ChildOf(sc),
			ot.Tags{
				"kafka.service": msg.Topic,
				"kafka.access":  "send",
			},
		)
	case ot.ErrSpanContextNotFound:
		sensor.Logger().Debug("no span context provided in message to ", msg.Topic, ", skipping the call", msg.Topic)
	case ot.ErrUnsupportedFormat:
		sensor.Logger().Info("unsupported span context format provided in message to ", msg.Topic, ", skipping the call")
	default:
		sensor.Logger().Warn("failed to extract span context from producer message headers: ", err)
	}

	return nil
}

func producerMessagesFromSameContext(msgs []*sarama.ProducerMessage) bool {
	if len(msgs) == 0 {
		return true
	}

	firstTraceID, firstSpanID, err := extractTraceSpanID(msgs[0])
	if err != nil {
		return false
	}

	for _, msg := range msgs[1:] {
		traceID, spanID, err := extractTraceSpanID(msg)
		if err != nil {
			return false
		}

		if traceID != firstTraceID || spanID != firstSpanID {
			return false
		}
	}

	return true
}
