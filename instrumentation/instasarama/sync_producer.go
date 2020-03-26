package instasarama

import (
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
	sensor *instana.Sensor
}

// NewSyncProducer wraps sarama.SyncProducer instance and instruments its calls
func NewSyncProducer(sp sarama.SyncProducer, sensor *instana.Sensor) *SyncProducer {
	return &SyncProducer{
		SyncProducer: sp,
		sensor:       sensor,
	}
}

// SendMessage picks up the trace context previously added to the message with
// instasarama.ProducerMessageWithSpan(), starts a new child span and injects its
// context into the message headers before sending it to the underlying producer.
// The call will not be traced if there the message does not contain trace context.
func (p *SyncProducer) SendMessage(msg *sarama.ProducerMessage) (int32, int64, error) {
	var sp ot.Span

	// pick up the existing trace context if provided and start a new span
	switch sc, err := p.sensor.Tracer().Extract(ot.TextMap, ProducerMessageCarrier{msg}); err {
	case nil:
		sp = p.sensor.Tracer().StartSpan(
			"kafka",
			ext.SpanKindProducer,
			ot.ChildOf(sc),
			ot.Tags{
				"kafka.service": msg.Topic,
				"kafka.access":  "send",
			})
		defer sp.Finish()

		// forward the trace context, updating the span ids
		sp.Tracer().Inject(sp.Context(), ot.TextMap, ProducerMessageCarrier{msg})
	case ot.ErrSpanContextNotFound:
		p.sensor.Logger().Debug("no span context provided in message to %q, starting a new exit span", msg.Topic)
	case ot.ErrUnsupportedFormat:
		p.sensor.Logger().Info("unsupported span context format provided in message to %q", msg.Topic)
	default:
		p.sensor.Logger().Warn("failed to extract span context from producer message headers: ", err)
	}

	partition, offset, err := p.SyncProducer.SendMessage(msg)
	if err != nil && sp != nil {
		sp.SetTag("kafka.error", err)
		sp.LogFields(otlog.Error(err))
	}

	return partition, offset, err
}
