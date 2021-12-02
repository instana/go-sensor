// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instasarama

import (
	"github.com/Shopify/sarama"
	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// ConsumerGroupHandler is a wrapper for sarama.ConsumerGroupHandler that creates an entry span for each
// incoming Kafka message, ensuring the extraction and continuation of the existing trace context if provided
type ConsumerGroupHandler struct {
	handler sarama.ConsumerGroupHandler
	sensor  *instana.Sensor
}

// WrapConsumerGroupHandler wraps the existing group handler and instruments its calls
func WrapConsumerGroupHandler(h sarama.ConsumerGroupHandler, sensor *instana.Sensor) *ConsumerGroupHandler {
	return &ConsumerGroupHandler{
		handler: h,
		sensor:  sensor,
	}
}

// Setup calls the underlying handler's Setup() method
func (h *ConsumerGroupHandler) Setup(sess sarama.ConsumerGroupSession) error {
	return h.handler.Setup(sess)
}

// Cleanup calls the underlying handler's Cleanup() method
func (h *ConsumerGroupHandler) Cleanup(sess sarama.ConsumerGroupSession) error {
	return h.handler.Cleanup(sess)
}

// ConsumeClaim injects the trace context into incoming message headers and delegates further processing to
// the underlying handler
func (h *ConsumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	wrappedSess := newConsumerGroupSession(sess, h.sensor)
	wrappedClaim := newConsumerGroupClaim(claim)

	go func() {
		for msg := range claim.Messages() {
			sp := wrappedSess.startSpan(msg)
			sp.Tracer().Inject(sp.Context(), ot.TextMap, ConsumerMessageCarrier{msg})
			wrappedClaim.messages <- msg
		}
		close(wrappedClaim.messages)
	}()

	return h.handler.ConsumeClaim(wrappedSess, wrappedClaim)
}

// consumerGroupSession is a wrapper for sarama.ConsumerGroupSession that keeps track of active spans associated
// with messages consumed during this session. The span is initiated by (*instasarama.ConsumerGroupHandler).ConsumeClaim()
// and finished when the message is marked as consumed by MarkMessage().
type consumerGroupSession struct {
	sarama.ConsumerGroupSession
	sensor      *instana.Sensor
	activeSpans *spanRegistry
}

func newConsumerGroupSession(sess sarama.ConsumerGroupSession, sensor *instana.Sensor) *consumerGroupSession {
	return &consumerGroupSession{
		ConsumerGroupSession: sess,
		sensor:               sensor,
		activeSpans:          newSpanRegistry(),
	}
}

func (sess *consumerGroupSession) startSpan(msg *sarama.ConsumerMessage) ot.Span {
	opts := []ot.StartSpanOption{
		ext.SpanKindConsumer,
		ot.Tags{
			"kafka.service": msg.Topic,
			"kafka.access":  "consume",
		},
	}
	if spanContext, ok := SpanContextFromConsumerMessage(msg, sess.sensor); ok {
		opts = append(opts, ot.ChildOf(spanContext))
	}

	sp := sess.sensor.Tracer().StartSpan("kafka", opts...)
	sess.activeSpans.Add(consumerSpanKey(msg), sp)

	return sp
}

func (sess *consumerGroupSession) MarkMessage(msg *sarama.ConsumerMessage, metadata string) {
	if sp, ok := sess.activeSpans.Remove(consumerSpanKey(msg)); ok {
		defer sp.Finish()
	}

	sess.ConsumerGroupSession.MarkMessage(msg, metadata)
}

// consumerGroupClaim is a wrapper for sarama.ConsumerGroupClaim that keeps messages after
// the trace header have been added until they are consumed by the original handler
type consumerGroupClaim struct {
	sarama.ConsumerGroupClaim
	messages chan *sarama.ConsumerMessage
}

func newConsumerGroupClaim(claim sarama.ConsumerGroupClaim) *consumerGroupClaim {
	return &consumerGroupClaim{
		ConsumerGroupClaim: claim,
		messages:           make(chan *sarama.ConsumerMessage),
	}
}

func (c *consumerGroupClaim) Messages() <-chan *sarama.ConsumerMessage {
	return c.messages
}
