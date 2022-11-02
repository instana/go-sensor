// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instasarama_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/Shopify/sarama"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instasarama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var headerFormats = []string{"" /* tests the default behavior */, "binary", "string", "both"}

func TestProducerMessageWithSpan(t *testing.T) {
	for _, headerFormat := range headerFormats {
		os.Setenv(instasarama.KafkaHeaderEnvVarKey, headerFormat)

		recorder := instana.NewTestRecorder()
		tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

		sp := tracer.StartSpan("test-span")
		pm := instasarama.ProducerMessageWithSpan(&sarama.ProducerMessage{
			Topic: "test-topic",
			Key:   sarama.StringEncoder("key1"),
			Value: sarama.StringEncoder("value1"),
			Headers: []sarama.RecordHeader{
				{Key: []byte("headerKey1"), Value: []byte("headerValue1")},
			},
		}, sp)
		sp.Finish()

		spans := recorder.GetQueuedSpans()
		require.Len(t, spans, 1)

		expected := []sarama.RecordHeader{
			{Key: []byte("headerKey1"), Value: []byte("headerValue1")},
		}

		if headerFormat == "both" || headerFormat == "binary" || headerFormat == "" /* -> default, currently both */ {
			expected = append(expected, []sarama.RecordHeader{
				{Key: []byte(instasarama.FieldL), Value: []byte{0x01}},
				{
					Key: []byte(instasarama.FieldC),
					Value: instasarama.PackTraceContextHeader(
						instana.FormatLongID(spans[0].TraceIDHi, spans[0].TraceID),
						instana.FormatID(spans[0].SpanID),
					),
				},
			}...)
		}

		if headerFormat == "both" || headerFormat == "string" || headerFormat == "" /* -> default, currently both */ {
			expected = append(expected, []sarama.RecordHeader{
				{Key: []byte(instasarama.FieldLS), Value: []byte("1")},
				{
					Key:   []byte(instasarama.FieldT),
					Value: []byte("0000000000000000" + instana.FormatID(spans[0].TraceID)),
				},
				{
					Key:   []byte(instasarama.FieldS),
					Value: []byte(instana.FormatID(spans[0].SpanID)),
				},
			}...)
		}
		assert.ElementsMatch(t, expected, pm.Headers)

		os.Unsetenv(instasarama.KafkaHeaderEnvVarKey)
	}
}

func TestProducerMessageWithSpanFromContext(t *testing.T) {
	for _, headerFormat := range headerFormats {
		os.Setenv(instasarama.KafkaHeaderEnvVarKey, headerFormat)

		recorder := instana.NewTestRecorder()
		tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

		sp := tracer.StartSpan("test-span")
		ctx := instana.ContextWithSpan(context.Background(), sp)

		pm := instasarama.ProducerMessageWithSpanFromContext(ctx, &sarama.ProducerMessage{
			Topic: "test-topic",
			Key:   sarama.StringEncoder("key1"),
			Value: sarama.StringEncoder("value1"),
			Headers: []sarama.RecordHeader{
				{Key: []byte("headerKey1"), Value: []byte("headerValue1")},
			},
		})
		sp.Finish()

		spans := recorder.GetQueuedSpans()
		require.Len(t, spans, 1)

		expected := []sarama.RecordHeader{
			{Key: []byte("headerKey1"), Value: []byte("headerValue1")},
		}

		if headerFormat == "both" || headerFormat == "binary" || headerFormat == "" /* -> default, currently both */ {
			expected = append(expected, []sarama.RecordHeader{
				{Key: []byte(instasarama.FieldL), Value: []byte{0x01}},
				{
					Key: []byte(instasarama.FieldC),
					Value: instasarama.PackTraceContextHeader(
						instana.FormatLongID(spans[0].TraceIDHi, spans[0].TraceID),
						instana.FormatID(spans[0].SpanID),
					),
				},
			}...)
		}

		if headerFormat == "both" || headerFormat == "string" || headerFormat == "" /* -> default, currently both */ {
			expected = append(expected, []sarama.RecordHeader{
				{Key: []byte(instasarama.FieldLS), Value: []byte("1")},
				{
					Key:   []byte(instasarama.FieldT),
					Value: []byte("0000000000000000" + instana.FormatID(spans[0].TraceID)),
				},
				{
					Key:   []byte(instasarama.FieldS),
					Value: []byte(instana.FormatID(spans[0].SpanID)),
				},
			}...)
		}
		assert.ElementsMatch(t, expected, pm.Headers)

		os.Unsetenv(instasarama.KafkaHeaderEnvVarKey)
	}
}

func TestProducerMessageCarrier_Set_FieldT(t *testing.T) {
	for _, headerFormat := range headerFormats {
		os.Setenv(instasarama.KafkaHeaderEnvVarKey, headerFormat)
		expected := []sarama.RecordHeader{}
		var msg sarama.ProducerMessage
		c := instasarama.ProducerMessageCarrier{&msg}
		c.Set(instana.FieldT, "0000000000000001deadbeefdeadbeef")

		if headerFormat == "both" || headerFormat == "binary" || headerFormat == "" /* -> default, currently both */ {
			expected = append(expected, []sarama.RecordHeader{
				{
					Key: []byte(instasarama.FieldC),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
						0xde, 0xad, 0xbe, 0xef, 0xde, 0xad, 0xbe, 0xef,
						// spanid
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					},
				},
			}...)
		}

		if headerFormat == "both" || headerFormat == "string" || headerFormat == "" /* -> default, currently both */ {
			expected = append(expected, []sarama.RecordHeader{
				{
					Key:   []byte(instasarama.FieldT),
					Value: []byte("0000000000000001deadbeefdeadbeef"),
				},
			}...)
		}

		assert.Equal(t, expected, msg.Headers)
		os.Unsetenv(instasarama.KafkaHeaderEnvVarKey)
	}
}

func TestProducerMessageCarrier_Update_FieldT(t *testing.T) {
	examples := map[string]struct {
		Value    string
		Headers  []sarama.RecordHeader
		Expected []sarama.RecordHeader
		EnvVar   string
	}{
		"existing has trace id only": {
			Value: "000000000000000000000000deadbeef",
			Headers: []sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
						0xab, 0xcd, 0xef, 0x12, 0xab, 0xcd, 0xef, 0x12,
						// span id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					},
				},
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			Expected: []sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
						0x00, 0x00, 0x00, 0x00, 0xde, 0xad, 0xbe, 0xef,
						// span id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					},
				},
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
		},
		"existing has span id only": {
			Value: "000000000000000000000000deadbeef",
			Headers: []sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
						// span id
						0x00, 0x00, 0x00, 0x00, 0xab, 0xcd, 0xef, 0x12,
					},
				},
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			Expected: []sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
						0x00, 0x00, 0x00, 0x00, 0xde, 0xad, 0xbe, 0xef,
						// span id
						0x00, 0x00, 0x00, 0x00, 0xab, 0xcd, 0xef, 0x12,
					},
				},
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
		},
		"existing has trace and span id": {
			Value: "000000000000000000000000deadbeef",
			Headers: []sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
						0xab, 0xcd, 0xef, 0x12, 0xab, 0xcd, 0xef, 0x12,
						// span id
						0x00, 0x00, 0x00, 0xab, 0xcd, 0xef, 0x12, 0x34,
					},
				},
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			Expected: []sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
						0x00, 0x00, 0x00, 0x00, 0xde, 0xad, 0xbe, 0xef,
						// span id
						0x00, 0x00, 0x00, 0xab, 0xcd, 0xef, 0x12, 0x34,
					},
				},
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
		},
		"128-bit trace id": {
			Value: "000000000000000200000000deadbeef",
			Headers: []sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
						0xab, 0xcd, 0xef, 0x12, 0xab, 0xcd, 0xef, 0x12,
						// span id
						0x00, 0x00, 0x00, 0xab, 0xcd, 0xef, 0x12, 0x34,
					},
				},
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			Expected: []sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02,
						0x00, 0x00, 0x00, 0x00, 0xde, 0xad, 0xbe, 0xef,
						// span id
						0x00, 0x00, 0x00, 0xab, 0xcd, 0xef, 0x12, 0x34,
					},
				},
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
		},
		"existing has trace id only: header is string": {
			Value: "000000000000000000000000deadbeef",
			Headers: []sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key:   []byte("x_instana_t"),
					Value: []byte("0000000000000000abcdef12abcdef12"),
				},
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			Expected: []sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key:   []byte("x_instana_t"),
					Value: []byte("000000000000000000000000deadbeef"),
				},
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			EnvVar: "string",
		},
		"existing has trace id only: header is both": {
			Value: "000000000000000000000000deadbeef",
			Headers: []sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key:   []byte("x_instana_t"),
					Value: []byte("0000000000000000abcdef12abcdef12"),
				},
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
						0xab, 0xcd, 0xef, 0x12, 0xab, 0xcd, 0xef, 0x12,
						// span id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					},
				},
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			Expected: []sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key:   []byte("x_instana_t"),
					Value: []byte("000000000000000000000000deadbeef"),
				},
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
						0x00, 0x00, 0x00, 0x00, 0xde, 0xad, 0xbe, 0xef,
						// span id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					},
				},
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			EnvVar: "both",
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			if example.EnvVar == "" {
				example.EnvVar = "binary"
			}
			os.Setenv(instasarama.KafkaHeaderEnvVarKey, example.EnvVar)

			msg := sarama.ProducerMessage{Headers: example.Headers}
			c := instasarama.ProducerMessageCarrier{&msg}

			c.Set(instana.FieldT, example.Value)
			assert.ElementsMatch(t, example.Expected, msg.Headers)

			os.Unsetenv(instasarama.KafkaHeaderEnvVarKey)
		})
	}
}

func TestProducerMessageCarrier_Set_FieldS(t *testing.T) {
	for _, headerFormat := range headerFormats {
		os.Setenv(instasarama.KafkaHeaderEnvVarKey, headerFormat)

		var msg sarama.ProducerMessage
		c := instasarama.ProducerMessageCarrier{&msg}

		c.Set(instana.FieldS, "00000000deadbeef")

		var expected []sarama.RecordHeader

		if headerFormat == "both" || headerFormat == "binary" || headerFormat == "" /* -> default, currently both */ {
			expected = append(expected, []sarama.RecordHeader{
				{
					Key: []byte(instasarama.FieldC),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
						// span id
						0x00, 0x00, 0x00, 0x00, 0xde, 0xad, 0xbe, 0xef,
					},
				},
			}...)
		}

		if headerFormat == "both" || headerFormat == "string" || headerFormat == "" /* -> default, currently both */ {
			expected = append(expected, []sarama.RecordHeader{
				{
					Key:   []byte(instasarama.FieldS),
					Value: []byte("00000000deadbeef"),
				},
			}...)
		}

		assert.Equal(t, expected, msg.Headers)

		os.Unsetenv(instasarama.KafkaHeaderEnvVarKey)
	}
}

func TestProducerMessageCarrier_Update_FieldS(t *testing.T) {
	examples := map[string]struct {
		Value    string
		Headers  []sarama.RecordHeader
		Expected []sarama.RecordHeader
		EnvVar   string
	}{
		"existing has trace id only": {
			Value: "00000000deadbeef",
			Headers: []sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
						0x00, 0x00, 0x00, 0x00, 0xab, 0xcd, 0xef, 0x12,
						// span id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					},
				},
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			Expected: []sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
						0x00, 0x00, 0x00, 0x00, 0xab, 0xcd, 0xef, 0x12,
						// span id
						0x00, 0x00, 0x00, 0x00, 0xde, 0xad, 0xbe, 0xef,
					},
				},
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
		},
		"existing has span id only": {
			Value: "00000000deadbeef",
			Headers: []sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
						// span id
						0x00, 0x00, 0x00, 0x00, 0xab, 0xcd, 0xef, 0x12,
					},
				},
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			Expected: []sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
						// span id
						0x00, 0x00, 0x00, 0x00, 0xde, 0xad, 0xbe, 0xef,
					},
				},
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
		},
		"existing has trace and span id": {
			Value: "00000000deadbeef",
			Headers: []sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
						0x00, 0x00, 0x00, 0x00, 0xab, 0xcd, 0xef, 0x12,
						// span id
						0x00, 0x00, 0x00, 0x00, 0xab, 0xcd, 0xef, 0x12,
					},
				},
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			Expected: []sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
						0x00, 0x00, 0x00, 0x00, 0xab, 0xcd, 0xef, 0x12,
						// span id
						0x00, 0x00, 0x00, 0x00, 0xde, 0xad, 0xbe, 0xef,
					},
				},
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
		},
		"existing has trace id only: header as string": {
			Value: "00000000deadbeef",
			Headers: []sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key:   []byte("x_instana_s"),
					Value: []byte("0000000000000000"),
				},
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			Expected: []sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key:   []byte("x_instana_s"),
					Value: []byte("00000000deadbeef"),
				},
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			EnvVar: "string",
		},
		"existing has trace id only: header as both": {
			Value: "00000000deadbeef",
			Headers: []sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key:   []byte("x_instana_s"),
					Value: []byte("0000000000000000"),
				},
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
						0x00, 0x00, 0x00, 0x00, 0xab, 0xcd, 0xef, 0x12,
						// span id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					},
				},
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			Expected: []sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key:   []byte("x_instana_s"),
					Value: []byte("00000000deadbeef"),
				},
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
						0x00, 0x00, 0x00, 0x00, 0xab, 0xcd, 0xef, 0x12,
						// span id
						0x00, 0x00, 0x00, 0x00, 0xde, 0xad, 0xbe, 0xef,
					},
				},
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			EnvVar: "both",
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			if example.EnvVar == "" {
				example.EnvVar = "binary"
			}

			os.Setenv(instasarama.KafkaHeaderEnvVarKey, example.EnvVar)

			msg := sarama.ProducerMessage{Headers: example.Headers}
			c := instasarama.ProducerMessageCarrier{&msg}

			c.Set(instana.FieldS, example.Value)
			assert.ElementsMatch(t, example.Expected, msg.Headers)

			os.Unsetenv(instasarama.KafkaHeaderEnvVarKey)
		})
	}
}

func TestProducerMessageCarrier_Set_FieldL(t *testing.T) {

	examples := []struct {
		Name     string
		Expected []sarama.RecordHeader
		Value    string
		EnvVar   string
	}{
		{
			Name: "Supressed as binary",
			Expected: []sarama.RecordHeader{
				{Key: []byte(instasarama.FieldL), Value: []byte{0x00}},
			},
			Value:  "0",
			EnvVar: "binary",
		},
		{
			Name: "Not supressed as binary",
			Expected: []sarama.RecordHeader{
				{Key: []byte(instasarama.FieldL), Value: []byte{0x01}},
			},
			Value:  "1",
			EnvVar: "binary",
		},
		{
			Name: "Supressed as string",
			Expected: []sarama.RecordHeader{
				{Key: []byte(instasarama.FieldLS), Value: []byte("0")},
			},
			Value:  "0",
			EnvVar: "string",
		},
		{
			Name: "Not supressed as string",
			Expected: []sarama.RecordHeader{
				{Key: []byte(instasarama.FieldLS), Value: []byte("1")},
			},
			Value:  "1",
			EnvVar: "string",
		},
		{
			Name: "Supressed as both",
			Expected: []sarama.RecordHeader{
				{Key: []byte(instasarama.FieldLS), Value: []byte("0")},
				{Key: []byte(instasarama.FieldL), Value: []byte{0x00}},
			},
			Value:  "0",
			EnvVar: "both",
		},
		{
			Name: "Not supressed as both",
			Expected: []sarama.RecordHeader{
				{Key: []byte(instasarama.FieldLS), Value: []byte("1")},
				{Key: []byte(instasarama.FieldL), Value: []byte{0x01}},
			},
			Value:  "1",
			EnvVar: "both",
		},
	}

	for _, example := range examples {
		t.Run(example.Value, func(t *testing.T) {
			os.Setenv(instasarama.KafkaHeaderEnvVarKey, example.EnvVar)
			msg := sarama.ProducerMessage{Headers: example.Expected}
			c := instasarama.ProducerMessageCarrier{&msg}

			c.Set(instana.FieldL, example.Value)
			assert.Equal(t, example.Expected, msg.Headers)
			os.Unsetenv(instasarama.KafkaHeaderEnvVarKey)
		})
	}
}

func TestProducerMessageCarrier_Update_FieldL(t *testing.T) {
	for _, headerFormat := range headerFormats {
		os.Setenv(instasarama.KafkaHeaderEnvVarKey, headerFormat)
		var headerSuppressed []sarama.RecordHeader
		var headerNotSuppressed []sarama.RecordHeader

		if headerFormat == "both" || headerFormat == "binary" || headerFormat == "" /* -> default, currently both */ {
			headerSuppressed = append(headerSuppressed, []sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{Key: []byte("x_instana_l"), Value: []byte{0x00}},
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			}...)

			headerNotSuppressed = append(headerNotSuppressed, []sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{Key: []byte("x_instana_l"), Value: []byte{0x01}},
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			}...)
		}

		if headerFormat == "both" || headerFormat == "string" || headerFormat == "" /* -> default, currently both */ {
			headerSuppressed = append(headerSuppressed, []sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{Key: []byte("x_instana_l_s"), Value: []byte("0")},
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			}...)

			headerNotSuppressed = append(headerNotSuppressed, []sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{Key: []byte("x_instana_l_s"), Value: []byte("1")},
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			}...)
		}

		msg := sarama.ProducerMessage{
			Headers: headerSuppressed,
		}
		c := instasarama.ProducerMessageCarrier{&msg}

		c.Set(instana.FieldL, "1")
		assert.ElementsMatch(t, headerNotSuppressed, msg.Headers)

		os.Unsetenv(instasarama.KafkaHeaderEnvVarKey)
	}
}

func TestProducerMessageCarrier_RemoveAll(t *testing.T) {
	msg := sarama.ProducerMessage{
		Headers: []sarama.RecordHeader{
			{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
			{
				Key: []byte("x_instana_c"),
				Value: []byte{
					// trace id
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
					0x00, 0x00, 0x00, 0x00, 0xab, 0xcd, 0xef, 0x12,
					// span id
					0x00, 0x00, 0x00, 0x00, 0xde, 0xad, 0xbe, 0xef,
				},
			},
			{
				Key:   []byte("x_instana_t"),
				Value: []byte("0000000000000000"),
			},
			{
				Key:   []byte("x_instana_s"),
				Value: []byte("0000000000000000"),
			},
			{
				Key:   []byte("x_instana_l_s"),
				Value: []byte("1"),
			},
			{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			{Key: []byte("x_INSTANA_L"), Value: []byte{0x01}},
		},
	}

	c := instasarama.ProducerMessageCarrier{&msg}
	c.RemoveAll()

	assert.ElementsMatch(t, []sarama.RecordHeader{
		{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
		{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
	}, msg.Headers)
}

func TestProducerMessageCarrier_ForeachKey(t *testing.T) {
	for _, headerFormat := range headerFormats {
		os.Setenv(instasarama.KafkaHeaderEnvVarKey, headerFormat)
		headers := []sarama.RecordHeader{
			{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
			{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
		}

		if headerFormat == "both" || headerFormat == "binary" || headerFormat == "" /* -> default, currently both */ {
			headers = append(headers, []sarama.RecordHeader{
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
						0x00, 0x00, 0x00, 0x00, 0xab, 0xcd, 0xef, 0x12,
						// span id
						0x00, 0x00, 0x00, 0x00, 0xde, 0xad, 0xbe, 0xef,
					},
				},
				{Key: []byte("x_INSTANA_L"), Value: []byte{0x01}},
			}...)
		}

		if headerFormat == "both" || headerFormat == "string" || headerFormat == "" /* -> default, currently both */ {
			headers = append(headers, []sarama.RecordHeader{
				{Key: []byte("x_instana_t"), Value: []byte("000000000000000100000000abcdef12")},
				{Key: []byte("x_instana_s"), Value: []byte("00000000deadbeef")},
				{Key: []byte("x_instana_l_s"), Value: []byte("1")},
			}...)
		}

		msg := sarama.ProducerMessage{
			Headers: headers,
		}
		c := instasarama.ProducerMessageCarrier{&msg}

		var collected []struct{ Key, Value string }
		require.NoError(t, c.ForeachKey(func(k, v string) error {
			collected = append(collected, struct{ Key, Value string }{k, v})
			return nil
		}))

		assert.ElementsMatch(t, []struct{ Key, Value string }{
			{Key: instana.FieldT, Value: "000000000000000100000000abcdef12"},
			{Key: instana.FieldS, Value: "00000000deadbeef"},
			{Key: instana.FieldL, Value: "1"},
		}, collected)

		os.Unsetenv(instasarama.KafkaHeaderEnvVarKey)
	}
}

func TestProducerMessageCarrier_ForeachKey_NoTracingHeaders(t *testing.T) {
	msg := sarama.ProducerMessage{
		Headers: []sarama.RecordHeader{
			{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
		},
	}
	c := instasarama.ProducerMessageCarrier{&msg}

	var collected []struct{ Key, Value string }
	require.NoError(t, c.ForeachKey(func(k, v string) error {
		collected = append(collected, struct{ Key, Value string }{k, v})
		return nil
	}))

	assert.Empty(t, collected)
}

func TestProducerMessageCarrier_ForeachKey_Error(t *testing.T) {
	msg := sarama.ProducerMessage{
		Headers: []sarama.RecordHeader{
			{
				Key: []byte("x_instana_c"),
				Value: []byte{
					// trace id
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
					0x00, 0x00, 0x00, 0x00, 0xab, 0xcd, 0xef, 0x12,
					// span id
					0x00, 0x00, 0x00, 0x00, 0xde, 0xad, 0xbe, 0xef,
				},
			},
			{Key: []byte("x_instana_t"), Value: []byte("000000000000000100000000abcdef12")},
			{Key: []byte("x_instana_s"), Value: []byte("00000000deadbeef")},
			{Key: []byte("x_instana_l_s"), Value: []byte("1")},
			{Key: []byte("x_INSTANA_L"), Value: []byte{0x01}},
		},
	}
	c := instasarama.ProducerMessageCarrier{&msg}

	assert.Error(t, c.ForeachKey(func(k, v string) error {
		return errors.New("something went wrong")
	}))
}

func TestSpanContextFromConsumerMessage(t *testing.T) {
	for _, headerFormat := range headerFormats {
		os.Setenv(instasarama.KafkaHeaderEnvVarKey, headerFormat)

		sensor := instana.NewSensorWithTracer(
			instana.NewTracerWithEverything(&instana.Options{}, instana.NewTestRecorder()),
		)

		var headers []*sarama.RecordHeader

		if headerFormat == "both" || headerFormat == "binary" || headerFormat == "" /* -> default, currently both */ {
			headers = []*sarama.RecordHeader{
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
						0x00, 0x00, 0x00, 0x00, 0xab, 0xcd, 0xef, 0x12,
						// span id
						0x00, 0x00, 0x00, 0x00, 0xde, 0xad, 0xbe, 0xef,
					},
				},
				{Key: []byte("x_instana_l"), Value: []byte{0x01}},
			}
		}

		if headerFormat == "both" || headerFormat == "string" || headerFormat == "" /* -> default, currently both */ {
			headers = []*sarama.RecordHeader{
				{Key: []byte("x_instana_t"), Value: []byte("000000000000000100000000abcdef12")},
				{Key: []byte("x_instana_s"), Value: []byte("00000000deadbeef")},
				{Key: []byte("x_instana_l_s"), Value: []byte("1")},
			}
		}

		msg := &sarama.ConsumerMessage{
			Headers: headers,
		}

		spanContext, ok := instasarama.SpanContextFromConsumerMessage(msg, sensor)
		require.True(t, ok)
		assert.Equal(t, instana.SpanContext{
			TraceIDHi: 0x00000001,
			TraceID:   0xabcdef12,
			SpanID:    0xdeadbeef,
			Baggage:   make(map[string]string),
		}, spanContext)

		os.Unsetenv(instasarama.KafkaHeaderEnvVarKey)
	}
}

func TestSpanContextFromConsumerMessage_NoContext(t *testing.T) {
	examples := []struct {
		Name         string
		Headers      []*sarama.RecordHeader
		HeaderFormat string
	}{
		{
			Name: "no tracing headers, header is binary",
			Headers: []*sarama.RecordHeader{
				{Key: []byte("key1"), Value: []byte("value1")},
				nil,
			},
			HeaderFormat: "binary",
		},
		{
			Name: "malformed tracing headers, header is binary",
			Headers: []*sarama.RecordHeader{
				{Key: []byte("x_instana_c"), Value: []byte("malformed")},
				{Key: []byte("x_instana_l"), Value: []byte{0x00}},
			},
			HeaderFormat: "binary",
		},
		{
			Name: "incomplete trace headers, header is binary",
			Headers: []*sarama.RecordHeader{
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
						0x00, 0x00, 0x00, 0x00, 0xab, 0xcd, 0xef, 0x12,
						// empty span id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					},
				},
				{Key: []byte("x_instana_l"), Value: []byte{0x01}},
			},
			HeaderFormat: "binary",
		},
		{
			Name: "no tracing headers, header is string",
			Headers: []*sarama.RecordHeader{
				{Key: []byte("key1"), Value: []byte("value1")},
				nil,
			},
			HeaderFormat: "string",
		},
		{
			Name: "malformed tracing headers, header is string",
			Headers: []*sarama.RecordHeader{
				{Key: []byte("x_instana_t"), Value: []byte("malformed")},
				{Key: []byte("x_instana_s"), Value: []byte("malformed")},
				{Key: []byte("x_instana_l_s"), Value: []byte("0")},
			},
			HeaderFormat: "string",
		},
		{
			Name: "incomplete trace headers, header is string",
			Headers: []*sarama.RecordHeader{
				{Key: []byte("x_instana_t"), Value: []byte("000000000000000100000000abcdef12")},
				{Key: []byte("x_instana_s"), Value: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
				{Key: []byte("x_instana_l_s"), Value: []byte("1")},
			},
			HeaderFormat: "string",
		},
		{
			Name: "malformed tracing headers, header is both",
			Headers: []*sarama.RecordHeader{
				{Key: []byte("x_instana_t"), Value: []byte("malformed")},
				{Key: []byte("x_instana_s"), Value: []byte("malformed")},
				{Key: []byte("x_instana_l_s"), Value: []byte("0")},
				{Key: []byte("x_instana_c"), Value: []byte("malformed")},
				{Key: []byte("x_instana_l"), Value: []byte{0x00}},
			},
			HeaderFormat: "both",
		},
		{
			Name: "incomplete trace headers, header is both",
			Headers: []*sarama.RecordHeader{
				{Key: []byte("x_instana_t"), Value: []byte("000000000000000100000000abcdef12")},
				{Key: []byte("x_instana_s"), Value: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
				{Key: []byte("x_instana_l_s"), Value: []byte("1")},
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
						0x00, 0x00, 0x00, 0x00, 0xab, 0xcd, 0xef, 0x12,
						// empty span id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					},
				},
				{Key: []byte("x_instana_l"), Value: []byte{0x01}},
			},
			HeaderFormat: "both",
		},
	}

	for _, example := range examples {
		t.Run(example.Name, func(t *testing.T) {
			os.Setenv(instasarama.KafkaHeaderEnvVarKey, example.HeaderFormat)
			sensor := instana.NewSensorWithTracer(
				instana.NewTracerWithEverything(&instana.Options{}, instana.NewTestRecorder()),
			)

			msg := &sarama.ConsumerMessage{Headers: example.Headers}

			_, ok := instasarama.SpanContextFromConsumerMessage(msg, sensor)
			assert.False(t, ok)

			os.Unsetenv(instasarama.KafkaHeaderEnvVarKey)
		})
	}
}

func TestConsumerMessageCarrier_Set_FieldT(t *testing.T) {
	headerFormats := []string{"binary", "string", "both"}
	for _, headerFormat := range headerFormats {
		os.Setenv(instasarama.KafkaHeaderEnvVarKey, headerFormat)

		var msg sarama.ConsumerMessage
		c := instasarama.ConsumerMessageCarrier{&msg}

		c.Set(instana.FieldT, "0000000000000001deadbeefdeadbeef")

		var expected []*sarama.RecordHeader

		if headerFormat == "both" || headerFormat == "binary" || headerFormat == "" /* -> default, currently both */ {
			expected = append(expected, []*sarama.RecordHeader{
				{
					Key: []byte(instasarama.FieldC),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
						0xde, 0xad, 0xbe, 0xef, 0xde, 0xad, 0xbe, 0xef,
						// spanid
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					},
				},
			}...)
		}

		if headerFormat == "both" || headerFormat == "string" || headerFormat == "" /* -> default, currently both */ {
			expected = append(expected, []*sarama.RecordHeader{
				{
					Key:   []byte(instasarama.FieldT),
					Value: []byte("0000000000000001deadbeefdeadbeef"),
				},
			}...)
		}

		assert.Equal(t, expected, msg.Headers)

		os.Unsetenv(instasarama.KafkaHeaderEnvVarKey)
	}
}

func TestConsumerMessageCarrier_Update_FieldT(t *testing.T) {
	examples := map[string]struct {
		Value    string
		Headers  []*sarama.RecordHeader
		Expected []*sarama.RecordHeader
		EnvVar   string
	}{
		"existing has trace id only, header is binary": {
			Value: "000000000000000100000000deadbeef",
			Headers: []*sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
						0xab, 0xcd, 0xef, 0x12, 0xab, 0xcd, 0xef, 0x12,
						// span id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					},
				},
				nil,
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			Expected: []*sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
						0x00, 0x00, 0x00, 0x00, 0xde, 0xad, 0xbe, 0xef,
						// span id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					},
				},
				nil,
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			EnvVar: "binary",
		},
		"existing has span id only, header is binary": {
			Value: "000000000000000100000000deadbeef",
			Headers: []*sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
						// span id
						0x00, 0x00, 0x00, 0x00, 0xab, 0xcd, 0xef, 0x12,
					},
				},
				nil,
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			Expected: []*sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
						0x00, 0x00, 0x00, 0x00, 0xde, 0xad, 0xbe, 0xef,
						// span id
						0x00, 0x00, 0x00, 0x00, 0xab, 0xcd, 0xef, 0x12,
					},
				},
				nil,
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			EnvVar: "binary",
		},
		"existing has trace and span id, header is binary": {
			Value: "000000000000000100000000deadbeef",
			Headers: []*sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
						0xab, 0xcd, 0xef, 0x12, 0xab, 0xcd, 0xef, 0x12,
						// span id
						0x00, 0x00, 0x00, 0xab, 0xcd, 0xef, 0x12, 0x34,
					},
				},
				nil,
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			Expected: []*sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
						0x00, 0x00, 0x00, 0x00, 0xde, 0xad, 0xbe, 0xef,
						// span id
						0x00, 0x00, 0x00, 0xab, 0xcd, 0xef, 0x12, 0x34,
					},
				},
				nil,
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			EnvVar: "binary",
		},
		"existing has trace id only, header is string": {
			Value: "000000000000000100000000deadbeef",
			Headers: []*sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key:   []byte("x_instana_t"),
					Value: []byte("0000000000000001abcdef12abcdef12"),
				},
				nil,
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			Expected: []*sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key:   []byte("x_instana_t"),
					Value: []byte("000000000000000100000000deadbeef"),
				},
				nil,
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			EnvVar: "string",
		},
		"existing has span id only, header is string": {
			Value: "000000000000000100000000deadbeef",
			Headers: []*sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key:   []byte("x_instana_s"),
					Value: []byte("00000000abcdef12"),
				},
				nil,
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			Expected: []*sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key:   []byte("x_instana_s"),
					Value: []byte("00000000abcdef12"),
				},
				{
					Key:   []byte(instasarama.FieldT),
					Value: []byte("000000000000000100000000deadbeef"),
				},
				nil,
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			EnvVar: "string",
		},
		"existing has trace and span id, header is string": {
			Value: "000000000000000100000000deadbeef",
			Headers: []*sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key:   []byte("x_instana_t"),
					Value: []byte("0000000000000001abcdef12abcdef12"),
				},
				{
					Key:   []byte("x_instana_s"),
					Value: []byte("000000abcdef1234"),
				},
				nil,
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			Expected: []*sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key:   []byte("x_instana_t"),
					Value: []byte("000000000000000100000000deadbeef"),
				},
				{
					Key:   []byte("x_instana_s"),
					Value: []byte("000000abcdef1234"),
				},
				nil,
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			EnvVar: "string",
		},
		"existing has trace id only, header is both": {
			Value: "000000000000000100000000deadbeef",
			Headers: []*sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
						0xab, 0xcd, 0xef, 0x12, 0xab, 0xcd, 0xef, 0x12,
						// span id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					},
				},
				{
					Key:   []byte("x_instana_t"),
					Value: []byte("0000000000000001abcdef12abcdef12"),
				},
				nil,
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			Expected: []*sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
						0x00, 0x00, 0x00, 0x00, 0xde, 0xad, 0xbe, 0xef,
						// span id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					},
				},
				{
					Key:   []byte("x_instana_t"),
					Value: []byte("000000000000000100000000deadbeef"),
				},
				nil,
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			EnvVar: "both",
		},
		"existing has span id only, header is both": {
			Value: "000000000000000100000000deadbeef",
			Headers: []*sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
						// span id
						0x00, 0x00, 0x00, 0x00, 0xab, 0xcd, 0xef, 0x12,
					},
				},
				{
					Key:   []byte("x_instana_s"),
					Value: []byte("00000000abcdef12"),
				},
				nil,
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			Expected: []*sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
						0x00, 0x00, 0x00, 0x00, 0xde, 0xad, 0xbe, 0xef,
						// span id
						0x00, 0x00, 0x00, 0x00, 0xab, 0xcd, 0xef, 0x12,
					},
				},
				{
					Key:   []byte("x_instana_s"),
					Value: []byte("00000000abcdef12"),
				},
				{
					Key:   []byte(instasarama.FieldT),
					Value: []byte("000000000000000100000000deadbeef"),
				},
				nil,
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			EnvVar: "both",
		},
		"existing has trace and span id, header is both": {
			Value: "000000000000000100000000deadbeef",
			Headers: []*sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
						0xab, 0xcd, 0xef, 0x12, 0xab, 0xcd, 0xef, 0x12,
						// span id
						0x00, 0x00, 0x00, 0xab, 0xcd, 0xef, 0x12, 0x34,
					},
				},
				{
					Key:   []byte("x_instana_t"),
					Value: []byte("0000000000000001abcdef12abcdef12"),
				},
				{
					Key:   []byte("x_instana_s"),
					Value: []byte("000000abcdef1234"),
				},
				nil,
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			Expected: []*sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
						0x00, 0x00, 0x00, 0x00, 0xde, 0xad, 0xbe, 0xef,
						// span id
						0x00, 0x00, 0x00, 0xab, 0xcd, 0xef, 0x12, 0x34,
					},
				},
				{
					Key:   []byte("x_instana_t"),
					Value: []byte("000000000000000100000000deadbeef"),
				},
				{
					Key:   []byte("x_instana_s"),
					Value: []byte("000000abcdef1234"),
				},
				nil,
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			EnvVar: "both",
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			os.Setenv(instasarama.KafkaHeaderEnvVarKey, example.EnvVar)

			msg := sarama.ConsumerMessage{Headers: example.Headers}
			c := instasarama.ConsumerMessageCarrier{&msg}

			c.Set(instana.FieldT, example.Value)
			assert.ElementsMatch(t, example.Expected, msg.Headers)

			os.Unsetenv(instasarama.KafkaHeaderEnvVarKey)
		})
	}
}

func TestConsumerMessageCarrier_Set_FieldS(t *testing.T) {
	for _, headerFormat := range headerFormats {
		os.Setenv(instasarama.KafkaHeaderEnvVarKey, headerFormat)

		var msg sarama.ConsumerMessage
		var expected []*sarama.RecordHeader
		c := instasarama.ConsumerMessageCarrier{&msg}

		if headerFormat == "both" || headerFormat == "binary" || headerFormat == "" /* -> default, currently both */ {
			expected = append(expected, []*sarama.RecordHeader{
				{
					Key: []byte(instasarama.FieldC),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
						// span id
						0x00, 0x00, 0x00, 0x00, 0xde, 0xad, 0xbe, 0xef,
					},
				},
			}...)
		}

		if headerFormat == "both" || headerFormat == "string" || headerFormat == "" /* -> default, currently both */ {
			expected = append(expected, []*sarama.RecordHeader{
				{
					Key:   []byte(instasarama.FieldS),
					Value: []byte("00000000deadbeef"),
				},
			}...)
		}

		c.Set(instana.FieldS, "00000000deadbeef")
		assert.Equal(t, expected, msg.Headers)

		os.Unsetenv(instasarama.KafkaHeaderEnvVarKey)
	}
}

func TestConsumerMessageCarrier_Update_FieldS(t *testing.T) {
	examples := map[string]struct {
		Value    string
		Headers  []*sarama.RecordHeader
		Expected []*sarama.RecordHeader
		EnvVar   string
	}{
		"existing has trace id only, header is binary": {
			Value: "00000000deadbeef",
			Headers: []*sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
						0x00, 0x00, 0x00, 0x00, 0xab, 0xcd, 0xef, 0x12,
						// span id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					},
				},
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			Expected: []*sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
						0x00, 0x00, 0x00, 0x00, 0xab, 0xcd, 0xef, 0x12,
						// span id
						0x00, 0x00, 0x00, 0x00, 0xde, 0xad, 0xbe, 0xef,
					},
				},
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			EnvVar: "binary",
		},
		"existing has span id only, header is binary": {
			Value: "00000000deadbeef",
			Headers: []*sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
						// span id
						0x00, 0x00, 0x00, 0x00, 0xab, 0xcd, 0xef, 0x12,
					},
				},
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			Expected: []*sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
						// span id
						0x00, 0x00, 0x00, 0x00, 0xde, 0xad, 0xbe, 0xef,
					},
				},
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			EnvVar: "binary",
		},
		"existing has trace and span id, header is binary": {
			Value: "00000000deadbeef",
			Headers: []*sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
						0x00, 0x00, 0x00, 0x00, 0xab, 0xcd, 0xef, 0x12,
						// span id
						0x00, 0x00, 0x00, 0x00, 0xab, 0xcd, 0xef, 0x12,
					},
				},
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			Expected: []*sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
						0x00, 0x00, 0x00, 0x00, 0xab, 0xcd, 0xef, 0x12,
						// span id
						0x00, 0x00, 0x00, 0x00, 0xde, 0xad, 0xbe, 0xef,
					},
				},
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			EnvVar: "binary",
		},
		"existing has trace id only, header is string": {
			Value: "00000000deadbeef",
			Headers: []*sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key:   []byte("x_instana_t"),
					Value: []byte("000000000000000100000000abcdef12"),
				},
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			Expected: []*sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key:   []byte("x_instana_t"),
					Value: []byte("000000000000000100000000abcdef12"),
				},
				{
					Key:   []byte("X_INSTANA_S"),
					Value: []byte("00000000deadbeef"),
				},
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			EnvVar: "string",
		},
		"existing has span id only, header is string": {
			Value: "00000000deadbeef",
			Headers: []*sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key:   []byte("x_instana_s"),
					Value: []byte("00000000abcdef12"),
				},
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			Expected: []*sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key:   []byte("x_instana_s"),
					Value: []byte("00000000deadbeef"),
				},
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			EnvVar: "string",
		},
		"existing has trace and span id, header is string": {
			Value: "00000000deadbeef",
			Headers: []*sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key:   []byte("x_instana_t"),
					Value: []byte("000000000000000100000000abcdef12"),
				},
				{
					Key:   []byte("x_instana_s"),
					Value: []byte("00000000abcdef12"),
				},
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			Expected: []*sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key:   []byte("x_instana_t"),
					Value: []byte("000000000000000100000000abcdef12"),
				},
				{
					Key:   []byte("x_instana_s"),
					Value: []byte("00000000deadbeef"),
				},
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			EnvVar: "string",
		},
		"existing has trace id only, header is both": {
			Value: "00000000deadbeef",
			Headers: []*sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
						0x00, 0x00, 0x00, 0x00, 0xab, 0xcd, 0xef, 0x12,
						// span id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					},
				},
				{
					Key:   []byte("x_instana_t"),
					Value: []byte("000000000000000100000000abcdef12"),
				},
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			Expected: []*sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
						0x00, 0x00, 0x00, 0x00, 0xab, 0xcd, 0xef, 0x12,
						// span id
						0x00, 0x00, 0x00, 0x00, 0xde, 0xad, 0xbe, 0xef,
					},
				},
				{
					Key:   []byte("x_instana_t"),
					Value: []byte("000000000000000100000000abcdef12"),
				},
				{
					Key:   []byte("X_INSTANA_S"),
					Value: []byte("00000000deadbeef"),
				},
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			EnvVar: "both",
		},
		"existing has span id only, header is both": {
			Value: "00000000deadbeef",
			Headers: []*sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
						// span id
						0x00, 0x00, 0x00, 0x00, 0xab, 0xcd, 0xef, 0x12,
					},
				},
				{
					Key:   []byte("x_instana_s"),
					Value: []byte("00000000abcdef12"),
				},
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			Expected: []*sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
						// span id
						0x00, 0x00, 0x00, 0x00, 0xde, 0xad, 0xbe, 0xef,
					},
				},
				{
					Key:   []byte("x_instana_s"),
					Value: []byte("00000000deadbeef"),
				},
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			EnvVar: "both",
		},
		"existing has trace and span id, header is both": {
			Value: "00000000deadbeef",
			Headers: []*sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
						0x00, 0x00, 0x00, 0x00, 0xab, 0xcd, 0xef, 0x12,
						// span id
						0x00, 0x00, 0x00, 0x00, 0xab, 0xcd, 0xef, 0x12,
					},
				},
				{
					Key:   []byte("x_instana_t"),
					Value: []byte("000000000000000100000000abcdef12"),
				},
				{
					Key:   []byte("x_instana_s"),
					Value: []byte("00000000abcdef12"),
				},
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			Expected: []*sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
						0x00, 0x00, 0x00, 0x00, 0xab, 0xcd, 0xef, 0x12,
						// span id
						0x00, 0x00, 0x00, 0x00, 0xde, 0xad, 0xbe, 0xef,
					},
				},
				{
					Key:   []byte("x_instana_t"),
					Value: []byte("000000000000000100000000abcdef12"),
				},
				{
					Key:   []byte("x_instana_s"),
					Value: []byte("00000000deadbeef"),
				},
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
			EnvVar: "both",
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			os.Setenv(instasarama.KafkaHeaderEnvVarKey, example.EnvVar)

			msg := sarama.ConsumerMessage{Headers: example.Headers}
			c := instasarama.ConsumerMessageCarrier{&msg}

			c.Set(instana.FieldS, example.Value)
			assert.ElementsMatch(t, example.Expected, msg.Headers)

			os.Unsetenv(instasarama.KafkaHeaderEnvVarKey)
		})
	}
}

func TestConsumerMessageCarrier_Set_FieldL(t *testing.T) {
	examples := []struct {
		Name     string
		Value    string
		Expected []*sarama.RecordHeader
		EnvVar   string
	}{
		{
			Name:  "Supressed, binary header",
			Value: "0",
			Expected: []*sarama.RecordHeader{
				{Key: []byte(instasarama.FieldL), Value: []byte{0x00}},
			},
			EnvVar: "binary",
		},
		{
			Name:  "Not supressed, binary header",
			Value: "1",
			Expected: []*sarama.RecordHeader{
				{Key: []byte(instasarama.FieldL), Value: []byte{0x01}},
			},
			EnvVar: "binary",
		},
		{
			Name:  "Supressed, string header",
			Value: "0",
			Expected: []*sarama.RecordHeader{
				{Key: []byte(instasarama.FieldLS), Value: []byte{0x00}},
			},
			EnvVar: "string",
		},
		{
			Name:  "Not supressed, string header",
			Value: "1",
			Expected: []*sarama.RecordHeader{
				{Key: []byte(instasarama.FieldLS), Value: []byte{0x01}},
			},
			EnvVar: "string",
		},
		{
			Name:  "Supressed, string and binary headers",
			Value: "0",
			Expected: []*sarama.RecordHeader{
				{Key: []byte(instasarama.FieldL), Value: []byte{0x00}},
				{Key: []byte(instasarama.FieldLS), Value: []byte{0x00}},
			},
			EnvVar: "both",
		},
		{
			Name:  "Not supressed, string and binary headers",
			Value: "1",
			Expected: []*sarama.RecordHeader{
				{Key: []byte(instasarama.FieldL), Value: []byte{0x01}},
				{Key: []byte(instasarama.FieldLS), Value: []byte{0x01}},
			},
			EnvVar: "both",
		},
	}

	for _, example := range examples {
		t.Run(example.Name, func(t *testing.T) {
			os.Setenv(instasarama.KafkaHeaderEnvVarKey, example.EnvVar)

			msg := sarama.ConsumerMessage{Headers: example.Expected}
			c := instasarama.ConsumerMessageCarrier{&msg}

			c.Set(instana.FieldL, example.Value)
			assert.Equal(t, example.Expected, msg.Headers)

			os.Unsetenv(instasarama.KafkaHeaderEnvVarKey)
		})
	}
}

func TestConsumerMessageCarrier_Update_FieldL(t *testing.T) {

	headerFormats := []string{"binary"}

	for _, headerFormat := range headerFormats {
		os.Setenv(instasarama.KafkaHeaderEnvVarKey, headerFormat)

		msg := sarama.ConsumerMessage{
			Headers: []*sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{Key: []byte("x_instana_l"), Value: []byte{0x00}},
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
		}
		c := instasarama.ConsumerMessageCarrier{&msg}

		c.Set(instana.FieldL, "1")
		assert.ElementsMatch(t, []*sarama.RecordHeader{
			{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
			{Key: []byte("x_instana_l"), Value: []byte{0x01}},
			{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
		}, msg.Headers)

		os.Unsetenv(instasarama.KafkaHeaderEnvVarKey)
	}
}

func TestConsumerMessageCarrier_RemoveAll(t *testing.T) {
	msg := sarama.ConsumerMessage{
		Headers: []*sarama.RecordHeader{
			{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
			{
				Key: []byte("x_instana_c"),
				Value: []byte{
					// trace id
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
					0x00, 0x00, 0x00, 0x00, 0xab, 0xcd, 0xef, 0x12,
					// span id
					0x00, 0x00, 0x00, 0x00, 0xde, 0xad, 0xbe, 0xef,
				},
			},
			{
				Key:   []byte("x_instana_t"),
				Value: []byte("000000000000000100000000abcdef12"),
			},
			{
				Key:   []byte("x_instana_s"),
				Value: []byte("00000000deadbeef"),
			},
			{
				Key:   []byte("x_instana_l_s"),
				Value: []byte("1"),
			},
			nil,
			{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			{Key: []byte("x_INSTANA_L"), Value: []byte{0x01}},
		},
	}

	c := instasarama.ConsumerMessageCarrier{&msg}
	c.RemoveAll()

	assert.ElementsMatch(t, []*sarama.RecordHeader{
		{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
		nil,
		{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
	}, msg.Headers)
}

func TestConsumerMessageCarrier_ForeachKey(t *testing.T) {
	headerFormats := []string{"both"}

	for _, headerFormat := range headerFormats {
		os.Setenv(instasarama.KafkaHeaderEnvVarKey, headerFormat)

		msg := sarama.ConsumerMessage{
			Headers: []*sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
		}

		if headerFormat == "both" || headerFormat == "binary" || headerFormat == "" /* -> default, currently both */ {
			msg.Headers = append(msg.Headers, []*sarama.RecordHeader{
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
						0x00, 0x00, 0x00, 0x00, 0xab, 0xcd, 0xef, 0x12,
						// span id
						0x00, 0x00, 0x00, 0x00, 0xde, 0xad, 0xbe, 0xef,
					},
				},
				nil,
				{Key: []byte("x_INSTANA_L"), Value: []byte{0x01}},
			}...)
		}

		if headerFormat == "both" || headerFormat == "string" || headerFormat == "" /* -> default, currently both */ {
			msg.Headers = append(msg.Headers, []*sarama.RecordHeader{
				{Key: []byte("x_instana_t"), Value: []byte("000000000000000100000000abcdef12")},
				{Key: []byte("x_instana_s"), Value: []byte("00000000deadbeef")},
				{Key: []byte("x_instana_l_s"), Value: []byte("1")},
				nil,
			}...)
		}

		c := instasarama.ConsumerMessageCarrier{&msg}

		var collected []struct{ Key, Value string }
		require.NoError(t, c.ForeachKey(func(k, v string) error {
			collected = append(collected, struct{ Key, Value string }{k, v})
			return nil
		}))

		assert.ElementsMatch(t, []struct{ Key, Value string }{
			{Key: instana.FieldT, Value: "000000000000000100000000abcdef12"},
			{Key: instana.FieldS, Value: "00000000deadbeef"},
			{Key: instana.FieldL, Value: "1"},
		}, collected)

		os.Unsetenv(instasarama.KafkaHeaderEnvVarKey)
	}
}

func TestConsumerMessageCarrier_ForeachKey_NoTracingHeaders(t *testing.T) {
	msg := sarama.ConsumerMessage{
		Headers: []*sarama.RecordHeader{
			{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
		},
	}
	c := instasarama.ConsumerMessageCarrier{&msg}

	var collected []struct{ Key, Value string }
	require.NoError(t, c.ForeachKey(func(k, v string) error {
		collected = append(collected, struct{ Key, Value string }{k, v})
		return nil
	}))

	assert.Empty(t, collected)
}

func TestConsumerMessageCarrier_ForeachKey_Error(t *testing.T) {
	msg := sarama.ConsumerMessage{
		Headers: []*sarama.RecordHeader{
			{
				Key: []byte("x_instana_c"),
				Value: []byte{
					// trace id
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
					0x00, 0x00, 0x00, 0x00, 0xab, 0xcd, 0xef, 0x12,
					// span id
					0x00, 0x00, 0x00, 0x00, 0xde, 0xad, 0xbe, 0xef,
				},
			},
			{Key: []byte("x_instana_t"), Value: []byte("000000000000000100000000abcdef12")},
			{Key: []byte("x_instana_s"), Value: []byte("00000000deadbeef")},
			{Key: []byte("x_instana_l_s"), Value: []byte("1")},
			{Key: []byte("x_INSTANA_L"), Value: []byte{0x01}},
		},
	}
	c := instasarama.ConsumerMessageCarrier{&msg}

	assert.Error(t, c.ForeachKey(func(k, v string) error {
		return errors.New("something went wrong")
	}))
}
