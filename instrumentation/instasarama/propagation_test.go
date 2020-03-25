package instasarama_test

import (
	"errors"
	"testing"

	"github.com/Shopify/sarama"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instasarama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProducerMessageCarrier_Set_FieldT(t *testing.T) {
	var msg sarama.ProducerMessage
	c := instasarama.ProducerMessageCarrier{&msg}

	c.Set(instana.FieldT, "deadbeefdeadbeef")
	assert.Equal(t, []sarama.RecordHeader{
		{
			Key: []byte(instasarama.FieldC),
			Value: []byte{
				// trace id
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0xde, 0xad, 0xbe, 0xef, 0xde, 0xad, 0xbe, 0xef,
				// spanid
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			},
		},
	}, msg.Headers)
}

func TestProducerMessageCarrier_Update_FieldT(t *testing.T) {
	examples := map[string]struct {
		Value    string
		Headers  []sarama.RecordHeader
		Expected []sarama.RecordHeader
	}{
		"existing has trace id only": {
			Value: "deadbeef",
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
			Value: "deadbeef",
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
			Value: "deadbeef",
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
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			msg := sarama.ProducerMessage{Headers: example.Headers}
			c := instasarama.ProducerMessageCarrier{&msg}

			c.Set(instana.FieldT, example.Value)
			assert.ElementsMatch(t, example.Expected, msg.Headers)
		})
	}
}

func TestProducerMessageCarrier_Set_FieldS(t *testing.T) {
	var msg sarama.ProducerMessage
	c := instasarama.ProducerMessageCarrier{&msg}

	c.Set(instana.FieldS, "deadbeef")
	assert.Equal(t, []sarama.RecordHeader{
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
	}, msg.Headers)
}

func TestProducerMessageCarrier_Update_FieldS(t *testing.T) {
	examples := map[string]struct {
		Value    string
		Headers  []sarama.RecordHeader
		Expected []sarama.RecordHeader
	}{
		"existing has trace id only": {
			Value: "deadbeef",
			Headers: []sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
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
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
						0x00, 0x00, 0x00, 0x00, 0xab, 0xcd, 0xef, 0x12,
						// span id
						0x00, 0x00, 0x00, 0x00, 0xde, 0xad, 0xbe, 0xef,
					},
				},
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
		},
		"existing has span id only": {
			Value: "deadbeef",
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
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
						// span id
						0x00, 0x00, 0x00, 0x00, 0xde, 0xad, 0xbe, 0xef,
					},
				},
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
		},
		"existing has trace and span id": {
			Value: "deadbeef",
			Headers: []sarama.RecordHeader{
				{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
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
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
						0x00, 0x00, 0x00, 0x00, 0xab, 0xcd, 0xef, 0x12,
						// span id
						0x00, 0x00, 0x00, 0x00, 0xde, 0xad, 0xbe, 0xef,
					},
				},
				{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			},
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			msg := sarama.ProducerMessage{Headers: example.Headers}
			c := instasarama.ProducerMessageCarrier{&msg}

			c.Set(instana.FieldS, example.Value)
			assert.ElementsMatch(t, example.Expected, msg.Headers)
		})
	}
}

func TestProducerMessageCarrier_Set_FieldL(t *testing.T) {
	examples := map[string][]sarama.RecordHeader{
		"0": []sarama.RecordHeader{
			{Key: []byte(instasarama.FieldL), Value: []byte{0x00}},
		},
		"1": []sarama.RecordHeader{
			{Key: []byte(instasarama.FieldL), Value: []byte{0x01}},
		},
	}

	for value, expected := range examples {
		t.Run(value, func(t *testing.T) {
			msg := sarama.ProducerMessage{Headers: expected}
			c := instasarama.ProducerMessageCarrier{&msg}

			c.Set(instana.FieldL, value)
			assert.Equal(t, expected, msg.Headers)
		})
	}
}

func TestProducerMessageCarrier_Update_FieldL(t *testing.T) {
	msg := sarama.ProducerMessage{
		Headers: []sarama.RecordHeader{
			{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
			{Key: []byte("x_instana_l"), Value: []byte{0x00}},
			{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
		},
	}
	c := instasarama.ProducerMessageCarrier{&msg}

	c.Set(instana.FieldL, "1")
	assert.ElementsMatch(t, []sarama.RecordHeader{
		{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
		{Key: []byte("x_instana_l"), Value: []byte{0x01}},
		{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
	}, msg.Headers)
}

func TestProducerMessageCarrier_ForeachKey(t *testing.T) {
	msg := sarama.ProducerMessage{
		Headers: []sarama.RecordHeader{
			{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
			{
				Key: []byte("x_instana_c"),
				Value: []byte{
					// trace id
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0xab, 0xcd, 0xef, 0x12,
					// span id
					0x00, 0x00, 0x00, 0x00, 0xde, 0xad, 0xbe, 0xef,
				},
			},
			{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
			{Key: []byte("x_INSTANA_L"), Value: []byte{0x01}},
		},
	}
	c := instasarama.ProducerMessageCarrier{&msg}

	var collected []struct{ Key, Value string }
	require.NoError(t, c.ForeachKey(func(k, v string) error {
		collected = append(collected, struct{ Key, Value string }{k, v})
		return nil
	}))

	assert.ElementsMatch(t, []struct{ Key, Value string }{
		{Key: instana.FieldT, Value: "abcdef12"},
		{Key: instana.FieldS, Value: "deadbeef"},
		{Key: instana.FieldL, Value: "1"},
	}, collected)
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
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0xab, 0xcd, 0xef, 0x12,
					// span id
					0x00, 0x00, 0x00, 0x00, 0xde, 0xad, 0xbe, 0xef,
				},
			},
			{Key: []byte("x_INSTANA_L"), Value: []byte{0x01}},
		},
	}
	c := instasarama.ProducerMessageCarrier{&msg}

	assert.Error(t, c.ForeachKey(func(k, v string) error {
		return errors.New("something went wrong")
	}))
}
