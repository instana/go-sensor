// +build go1.9

package instasarama

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/Shopify/sarama"
	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
)

const (
	// FieldC is the trace context header key
	FieldC = "X_INSTANA_C"
	// FieldL is the trace level header key
	FieldL = "X_INSTANA_L"
)

var (
	fieldCKey = []byte(FieldC)
	fieldLKey = []byte(FieldL)
)

// ProducerMessageWithSpan injects the tracing context into producer message headers to propagate
// them through the Kafka requests made with instasarama producers.
func ProducerMessageWithSpan(pm *sarama.ProducerMessage, sp ot.Span) *sarama.ProducerMessage {
	sp.Tracer().Inject(sp.Context(), ot.TextMap, ProducerMessageCarrier{Message: pm})
	return pm
}

// ProducerMessageCarrier is a trace context carrier that propagates Instana OpenTracing
// headers throughout Kafka producer messages
type ProducerMessageCarrier struct {
	Message *sarama.ProducerMessage
}

// Set implements opentracing.TextMapWriter for ProducerMessageCarrier
func (c ProducerMessageCarrier) Set(key, val string) {
	switch strings.ToLower(key) {
	case instana.FieldT:
		if len(val) > 32 {
			return // ignore hex-encoded trace IDs longer than 128 bit
		}

		traceContext := PackTraceContextHeader(val, "")
		if i, ok := c.indexOf(fieldCKey); ok {
			// preserve the trace ID if the trace context header already present
			existingC := c.Message.Headers[i].Value
			if len(existingC) >= 16 {
				copy(traceContext[16:], existingC[16:])
			}
		}

		c.addOrReplaceHeader(fieldCKey, traceContext)
	case instana.FieldS:
		if len(val) > 16 {
			return // ignore hex-encoded span IDs longer than 64 bit
		}

		traceContext := PackTraceContextHeader("", val)
		if i, ok := c.indexOf(fieldCKey); ok {
			// preserve the span ID if the trace context header already present
			existingC := c.Message.Headers[i].Value
			if len(existingC) >= 16 {
				copy(traceContext[:16], existingC[:16])
			}
		}

		c.addOrReplaceHeader(fieldCKey, traceContext)
	case instana.FieldL:
		c.addOrReplaceHeader(fieldLKey, PackTraceLevelHeader(val))
	}
}

// RemoveAll removes all tracing headers previously set by Set()
func (c ProducerMessageCarrier) RemoveAll() {
	var ln int
	for _, header := range c.Message.Headers {
		if bytes.EqualFold(header.Key, fieldCKey) || bytes.EqualFold(header.Key, fieldLKey) {
			continue
		}

		c.Message.Headers[ln] = header
		ln++
	}

	c.Message.Headers = c.Message.Headers[:ln]
}

// ForeachKey implements opentracing.TextMapReader for ProducerMessageCarrier
func (c ProducerMessageCarrier) ForeachKey(handler func(key, val string) error) error {
	for _, header := range c.Message.Headers {
		switch {
		case bytes.EqualFold(header.Key, fieldCKey):
			traceID, spanID, err := UnpackTraceContextHeader(header.Value)
			if err != nil {
				return fmt.Errorf("malformed %q header: %s", header.Key, err)
			}

			if err := handler(instana.FieldT, string(traceID)); err != nil {
				return err
			}

			if err := handler(instana.FieldS, string(spanID)); err != nil {
				return err
			}
		case bytes.EqualFold(header.Key, fieldLKey):
			val, err := UnpackTraceLevelHeader(header.Value)
			if err != nil {
				return fmt.Errorf("malformed %q header: %s", header.Key, err)
			}

			if err := handler(instana.FieldL, val); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c ProducerMessageCarrier) addOrReplaceHeader(key, val []byte) {
	if i, ok := c.indexOf(key); ok {
		c.Message.Headers[i].Value = val
		return
	}

	c.Message.Headers = append(c.Message.Headers, sarama.RecordHeader{Key: key, Value: val})
}

func (c ProducerMessageCarrier) indexOf(key []byte) (int, bool) {
	for i, header := range c.Message.Headers {
		if bytes.EqualFold(key, header.Key) {
			return i, true
		}
	}

	return -1, false
}

// SpanContextFromConsumerMessage extracts the tracing context from consumer message
func SpanContextFromConsumerMessage(cm *sarama.ConsumerMessage, sensor *instana.Sensor) (ot.SpanContext, bool) {
	spanContext, err := sensor.Tracer().Extract(ot.TextMap, ConsumerMessageCarrier{Message: cm})
	if err != nil {
		return nil, false
	}

	return spanContext, true
}

// ConsumerMessageCarrier is a trace context carrier that extracts Instana OpenTracing
// headers from Kafka consumer messages
type ConsumerMessageCarrier struct {
	Message *sarama.ConsumerMessage
}

// Set implements opentracing.TextMapWriter for ConsumerMessageCarrier
func (c ConsumerMessageCarrier) Set(key, val string) {
	switch strings.ToLower(key) {
	case instana.FieldT:
		if len(val) > 32 {
			return // ignore hex-encoded trace IDs longer than 128 bit
		}

		traceContext := PackTraceContextHeader(val, "")
		if i, ok := c.indexOf(fieldCKey); ok {
			// preserve the trace ID if the trace context header already present
			existingC := c.Message.Headers[i].Value
			if len(existingC) >= 16 {
				copy(traceContext[16:], existingC[16:])
			}
		}

		c.addOrReplaceHeader(fieldCKey, traceContext)
	case instana.FieldS:
		if len(val) > 16 {
			return // ignore hex-encoded span IDs longer than 64 bit
		}

		traceContext := PackTraceContextHeader("", val)
		if i, ok := c.indexOf(fieldCKey); ok {
			// preserve the span ID if the trace context header already present
			existingC := c.Message.Headers[i].Value
			if len(existingC) >= 16 {
				copy(traceContext[:16], existingC[:16])
			}
		}

		c.addOrReplaceHeader(fieldCKey, traceContext)
	case instana.FieldL:
		c.addOrReplaceHeader(fieldLKey, PackTraceLevelHeader(val))
	}
}

// RemoveAll removes all tracing headers previously set by Set()
func (c ConsumerMessageCarrier) RemoveAll() {
	var ln int
	for _, header := range c.Message.Headers {
		if header != nil && (bytes.EqualFold(header.Key, fieldCKey) || bytes.EqualFold(header.Key, fieldLKey)) {
			continue
		}

		c.Message.Headers[ln] = header
		ln++
	}

	c.Message.Headers = c.Message.Headers[:ln]
}

// ForeachKey implements opentracing.TextMapReader for ConsumerMessageCarrier
func (c ConsumerMessageCarrier) ForeachKey(handler func(key, val string) error) error {
	for _, header := range c.Message.Headers {
		if header == nil {
			continue
		}

		switch {
		case bytes.EqualFold(header.Key, fieldCKey):
			traceID, spanID, err := UnpackTraceContextHeader(header.Value)
			if err != nil {
				return fmt.Errorf("malformed %q header: %s", header.Key, err)
			}

			if err := handler(instana.FieldT, string(traceID)); err != nil {
				return err
			}

			if err := handler(instana.FieldS, string(spanID)); err != nil {
				return err
			}
		case bytes.EqualFold(header.Key, fieldLKey):
			val, err := UnpackTraceLevelHeader(header.Value)
			if err != nil {
				return fmt.Errorf("malformed %q header: %s", header.Key, err)
			}

			if err := handler(instana.FieldL, val); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c ConsumerMessageCarrier) addOrReplaceHeader(key, val []byte) {
	if i, ok := c.indexOf(key); ok {
		c.Message.Headers[i].Value = val
		return
	}

	c.Message.Headers = append(c.Message.Headers, &sarama.RecordHeader{Key: key, Value: val})
}

func (c ConsumerMessageCarrier) indexOf(key []byte) (int, bool) {
	for i, header := range c.Message.Headers {
		if header == nil {
			continue
		}

		if bytes.EqualFold(key, header.Key) {
			return i, true
		}
	}

	return -1, false
}

func contextPropagationSupported(ver sarama.KafkaVersion) bool {
	return ver.IsAtLeast(sarama.V0_11_0_0)
}

func extractTraceSpanID(msg *sarama.ProducerMessage) (string, string, error) {
	var traceID, spanID string
	err := ProducerMessageCarrier{msg}.ForeachKey(func(k, v string) error {
		switch k {
		case instana.FieldT:
			traceID = v
		case instana.FieldS:
			spanID = v
		}

		return nil
	})

	return traceID, spanID, err
}
