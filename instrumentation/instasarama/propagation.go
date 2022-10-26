// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instasarama

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Shopify/sarama"
	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
)

const KafkaHeaderEnvVarKey = "INSTANA_KAFKA_HEADER_FORMAT"

const (
	// Legacy binary headers

	// FieldC is the trace context header key
	FieldC = "X_INSTANA_C"
	// FieldL is the trace level header key
	FieldL = "X_INSTANA_L"

	// New string headers

	// FieldT is the trace id
	FieldT = "X_INSTANA_T"
	// FieldS is the span id
	FieldS = "X_INSTANA_S"
	// FieldLS is the trace level
	FieldLS = "X_INSTANA_L_S"
)

const (
	binaryFormat = "binary"
	stringFormat = "string"
	bothFormat   = "both"
)

var (
	fieldCKey = []byte(FieldC)
	fieldLKey = []byte(FieldL)

	fieldTKey  = []byte(FieldT)
	fieldSKey  = []byte(FieldS)
	fieldLSKey = []byte(FieldLS)
)

// ProducerMessageWithSpan injects the tracing context into producer message headers to propagate
// them through the Kafka requests made with instasarama producers.
func ProducerMessageWithSpan(pm *sarama.ProducerMessage, sp ot.Span) *sarama.ProducerMessage {
	sp.Tracer().Inject(sp.Context(), ot.TextMap, ProducerMessageCarrier{Message: pm})
	return pm
}

// ProducerMessageWithSpanFromContext injects the tracing context into producer's
// message headers from the context if it is there.
func ProducerMessageWithSpanFromContext(ctx context.Context, pm *sarama.ProducerMessage) *sarama.ProducerMessage {
	sp, ok := instana.SpanFromContext(ctx)
	if !ok {
		return pm
	}

	return ProducerMessageWithSpan(pm, sp)
}

// ProducerMessageCarrier is a trace context carrier that propagates Instana OpenTracing
// headers throughout Kafka producer messages
type ProducerMessageCarrier struct {
	Message *sarama.ProducerMessage
}

// Set implements opentracing.TextMapWriter for ProducerMessageCarrier
func (c ProducerMessageCarrier) Set(key, val string) {
	kafkaHeaderFormat := getKafkaHeaderFormat()
	switch strings.ToLower(key) {
	case instana.FieldT:
		if kafkaHeaderFormat == bothFormat || kafkaHeaderFormat == binaryFormat {
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
		}

		if kafkaHeaderFormat == bothFormat || kafkaHeaderFormat == stringFormat {
			// There is no need to preserve any values, as thr trace context (aka X_INSTANA_C) is no longer present when the
			// header is of string format, wehre we have 2 separated header values: X_INSTANA_T and X_INSTANA_S
			existingT := val
			valLen := len(val)

			if valLen < 32 {
				existingT = strings.Repeat("0", 32-valLen) + val
			}

			if valLen > 32 {
				existingT = existingT[:32]
			}

			c.addOrReplaceHeader(fieldTKey, []byte(existingT))
		}
	case instana.FieldS:
		if kafkaHeaderFormat == bothFormat || kafkaHeaderFormat == binaryFormat {
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
		}

		if kafkaHeaderFormat == bothFormat || kafkaHeaderFormat == stringFormat {
			if len(val) > 16 {
				return // ignore hex-encoded span IDs longer than 64 bit
			}

			// There is no need to preserve any values, as thr trace context (aka X_INSTANA_C) is no longer present when the
			// header is of string format, wehre we have 2 separated header values: X_INSTANA_T and X_INSTANA_S
			c.addOrReplaceHeader(fieldSKey, []byte(val))
		}
	case instana.FieldL:
		if kafkaHeaderFormat == bothFormat || kafkaHeaderFormat == binaryFormat {
			c.addOrReplaceHeader(fieldLKey, PackTraceLevelHeader(val))
		}

		if kafkaHeaderFormat == bothFormat || kafkaHeaderFormat == stringFormat {
			c.addOrReplaceHeader(fieldLSKey, []byte(val))
		}
	}
}

// RemoveAll removes all tracing headers previously set by Set()
func (c ProducerMessageCarrier) RemoveAll() {
	var ln int
	for _, header := range c.Message.Headers {
		if bytes.EqualFold(header.Key, fieldCKey) ||
			bytes.EqualFold(header.Key, fieldLKey) ||
			bytes.EqualFold(header.Key, fieldTKey) ||
			bytes.EqualFold(header.Key, fieldSKey) ||
			bytes.EqualFold(header.Key, fieldLSKey) {
			continue
		}

		c.Message.Headers[ln] = header
		ln++
	}

	c.Message.Headers = c.Message.Headers[:ln]
}

// ForeachKey implements opentracing.TextMapReader for ProducerMessageCarrier
func (c ProducerMessageCarrier) ForeachKey(handler func(key, val string) error) error {
	kafkaHeaderFormat := getKafkaHeaderFormat()

	for _, header := range c.Message.Headers {
		switch {
		// If the customer sets kafka headers to be both binary and string, we don't want to duplicate the values for
		// X_INSTANA_T, X_INSTANA_S and X_INSTANA_L, so we bypass the binary one
		case bytes.EqualFold(header.Key, fieldCKey) && kafkaHeaderFormat != bothFormat:
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
		case bytes.EqualFold(header.Key, fieldLKey) && kafkaHeaderFormat != bothFormat:
			val, err := UnpackTraceLevelHeader(header.Value)
			if err != nil {
				return fmt.Errorf("malformed %q header: %s", header.Key, err)
			}

			if err := handler(instana.FieldL, val); err != nil {
				return err
			}
		case bytes.EqualFold(header.Key, fieldTKey):
			if err := handler(instana.FieldT, string(header.Value)); err != nil {
				return err
			}
		case bytes.EqualFold(header.Key, fieldSKey):
			if err := handler(instana.FieldS, string(header.Value)); err != nil {
				return err
			}
		case bytes.EqualFold(header.Key, fieldLSKey):
			if err := handler(instana.FieldL, string(header.Value)); err != nil {
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
	kafkaHeaderFormat := getKafkaHeaderFormat()

	switch strings.ToLower(key) {
	case instana.FieldT:
		if kafkaHeaderFormat == bothFormat || kafkaHeaderFormat == binaryFormat {
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
		}

		if kafkaHeaderFormat == bothFormat || kafkaHeaderFormat == stringFormat {
			// There is no need to preserve any values, as thr trace context (aka X_INSTANA_C) is no longer present when the
			// header is of string format, wehre we have 2 separated header values: X_INSTANA_T and X_INSTANA_S
			c.addOrReplaceHeader(fieldTKey, []byte(val))
		}
	case instana.FieldS:
		if kafkaHeaderFormat == bothFormat || kafkaHeaderFormat == binaryFormat {
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
		}

		if kafkaHeaderFormat == bothFormat || kafkaHeaderFormat == stringFormat {
			// There is no need to preserve any values, as thr trace context (aka X_INSTANA_C) is no longer present when the
			// header is of string format, wehre we have 2 separated header values: X_INSTANA_T and X_INSTANA_S
			c.addOrReplaceHeader(fieldSKey, []byte(val))
		}
	case instana.FieldL:
		if kafkaHeaderFormat == bothFormat || kafkaHeaderFormat == binaryFormat {
			c.addOrReplaceHeader(fieldLKey, PackTraceLevelHeader(val))
		}

		if kafkaHeaderFormat == bothFormat || kafkaHeaderFormat == stringFormat {
			c.addOrReplaceHeader(fieldLSKey, []byte(val))
		}
	}
}

// RemoveAll removes all tracing headers previously set by Set()
func (c ConsumerMessageCarrier) RemoveAll() {
	var ln int
	for _, header := range c.Message.Headers {
		if header != nil && (bytes.EqualFold(header.Key, fieldCKey) ||
			bytes.EqualFold(header.Key, fieldTKey) ||
			bytes.EqualFold(header.Key, fieldSKey) ||
			bytes.EqualFold(header.Key, fieldLSKey) ||
			bytes.EqualFold(header.Key, fieldLKey)) {
			continue
		}

		c.Message.Headers[ln] = header
		ln++
	}

	c.Message.Headers = c.Message.Headers[:ln]
}

// ForeachKey implements opentracing.TextMapReader for ConsumerMessageCarrier
func (c ConsumerMessageCarrier) ForeachKey(handler func(key, val string) error) error {
	kafkaHeaderFormat := getKafkaHeaderFormat()
	for _, header := range c.Message.Headers {
		if header == nil {
			continue
		}

		switch {
		case bytes.EqualFold(header.Key, fieldCKey) && kafkaHeaderFormat != bothFormat:
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
		case bytes.EqualFold(header.Key, fieldLKey) && kafkaHeaderFormat != bothFormat:
			val, err := UnpackTraceLevelHeader(header.Value)
			if err != nil {
				return fmt.Errorf("malformed %q header: %s", header.Key, err)
			}

			if err := handler(instana.FieldL, val); err != nil {
				return err
			}
		case bytes.EqualFold(header.Key, fieldTKey):
			if err := handler(instana.FieldT, string(header.Value)); err != nil {
				return err
			}
		case bytes.EqualFold(header.Key, fieldSKey):
			if err := handler(instana.FieldS, string(header.Value)); err != nil {
				return err
			}
		case bytes.EqualFold(header.Key, fieldLSKey):
			if err := handler(instana.FieldL, string(header.Value)); err != nil {
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

func getKafkaHeaderFormat() string {
	kafkaHeaderEnvVar, ok := os.LookupEnv(KafkaHeaderEnvVarKey)

	if ok && isValidFormat(kafkaHeaderEnvVar) {
		return kafkaHeaderEnvVar
	}

	return bothFormat
}

func isValidFormat(format string) bool {
	return format == binaryFormat || format == stringFormat || format == bothFormat
}
