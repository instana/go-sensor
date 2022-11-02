// (c) Copyright IBM Corp. 2022

package instaamqp

import (
	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
	"github.com/streadway/amqp"
)

// messageCarrier holds the data injected into and extracted from a span's context to assure span correlation
type messageCarrier struct {
	headers amqp.Table
	logger  instana.LeveledLogger
}

// ForeachKey implements opentracing.TextMapReader for MessageCarrier
func (c *messageCarrier) ForeachKey(handler func(key, value string) error) error {
	headers := c.headers

	if headers == nil {
		c.logger.Info("amqp: no headers provided")
		return nil
	}

	for key, val := range headers {
		valAsString, ok := val.(string)

		if !ok {
			continue
		}

		if err := handler(key, valAsString); err != nil {
			return err
		}
	}

	return nil
}

// Set implements opentracing.TextMapWriter for MessageCarrier
func (c *messageCarrier) Set(key, value string) {
	headers := c.headers
	if c.headers == nil {
		c.logger.Info("amqp: no headers provided")

		return
	}

	switch key {
	case instana.FieldT, instana.FieldL, instana.FieldS:
		headers[key] = value
	}
}

// SpanContextFromConsumerMessage extracts the tracing context from amqp.Delivery#Headers
func SpanContextFromConsumerMessage(d amqp.Delivery, sensor *instana.Sensor) (ot.SpanContext, bool) {
	sc, err := sensor.Tracer().Extract(ot.TextMap, &messageCarrier{d.Headers, sensor.Logger()})
	if err != nil {
		return nil, false
	}

	return sc, true
}
