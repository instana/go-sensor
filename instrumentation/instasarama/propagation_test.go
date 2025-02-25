// (c) Copyright IBM Corp. 2023

//go:build go1.17
// +build go1.17

package instasarama_test

import (
	"context"
	"errors"
	"testing"

	"github.com/IBM/sarama"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instasarama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProducerMessageWithSpan(t *testing.T) {

	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	sp := c.StartSpan("test-span")
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
	assert.ElementsMatch(t, expected, pm.Headers)

}

func TestProducerMessageWithSpanFromContext(t *testing.T) {

	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	sp := c.StartSpan("test-span")
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
	assert.ElementsMatch(t, expected, pm.Headers)

}

func TestProducerMessageCarrier_Set_FieldT(t *testing.T) {
	expected := []sarama.RecordHeader{}
	var msg sarama.ProducerMessage
	c := instasarama.ProducerMessageCarrier{&msg}
	c.Set(instana.FieldT, "0000000000000001deadbeefdeadbeef")

	expected = append(expected, []sarama.RecordHeader{
		{
			Key:   []byte(instasarama.FieldT),
			Value: []byte("0000000000000001deadbeefdeadbeef"),
		},
	}...)

	assert.Equal(t, expected, msg.Headers)
}

func TestProducerMessageCarrier_Update_FieldT(t *testing.T) {
	examples := map[string]struct {
		Value    string
		Headers  []sarama.RecordHeader
		Expected []sarama.RecordHeader
		EnvVar   string
	}{
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

	c.Set(instana.FieldS, "00000000deadbeef")

	var expected []sarama.RecordHeader

	expected = append(expected, []sarama.RecordHeader{
		{
			Key:   []byte(instasarama.FieldS),
			Value: []byte("00000000deadbeef"),
		},
	}...)

	assert.Equal(t, expected, msg.Headers)

}

func TestProducerMessageCarrier_Update_FieldS(t *testing.T) {
	examples := map[string]struct {
		Value    string
		Headers  []sarama.RecordHeader
		Expected []sarama.RecordHeader
	}{
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

	examples := []struct {
		Name     string
		Expected []sarama.RecordHeader
		Value    string
		EnvVar   string
	}{
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
	}

	for _, example := range examples {
		t.Run(example.Value, func(t *testing.T) {
			msg := sarama.ProducerMessage{Headers: example.Expected}
			c := instasarama.ProducerMessageCarrier{&msg}

			c.Set(instana.FieldL, example.Value)
			assert.Equal(t, example.Expected, msg.Headers)
		})
	}
}

func TestProducerMessageCarrier_Update_FieldL(t *testing.T) {
	var headerSuppressed []sarama.RecordHeader
	var headerNotSuppressed []sarama.RecordHeader

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

	msg := sarama.ProducerMessage{
		Headers: headerSuppressed,
	}
	c := instasarama.ProducerMessageCarrier{&msg}

	c.Set(instana.FieldL, "1")
	assert.ElementsMatch(t, headerNotSuppressed, msg.Headers)

}

func TestProducerMessageCarrier_RemoveAll(t *testing.T) {
	msg := sarama.ProducerMessage{
		Headers: []sarama.RecordHeader{
			{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
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
	headers := []sarama.RecordHeader{
		{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
		{Key: []byte("X_CUSTOM_2"), Value: []byte("value2")},
	}

	headers = append(headers, []sarama.RecordHeader{
		{Key: []byte("x_instana_t"), Value: []byte("000000000000000100000000abcdef12")},
		{Key: []byte("x_instana_s"), Value: []byte("00000000deadbeef")},
		{Key: []byte("x_instana_l_s"), Value: []byte("1")},
	}...)

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
			{Key: []byte("x_instana_t"), Value: []byte("000000000000000100000000abcdef12")},
			{Key: []byte("x_instana_s"), Value: []byte("00000000deadbeef")},
			{Key: []byte("x_instana_l_s"), Value: []byte("1")},
		},
	}
	c := instasarama.ProducerMessageCarrier{&msg}

	assert.Error(t, c.ForeachKey(func(k, v string) error {
		return errors.New("something went wrong")
	}))
}

func TestSpanContextFromConsumerMessage(t *testing.T) {

	c := instana.InitCollector(&instana.Options{
		AgentClient: alwaysReadyClient{},
		Recorder:    instana.NewTestRecorder(),
	})
	defer instana.ShutdownCollector()

	var headers []*sarama.RecordHeader

	headers = []*sarama.RecordHeader{
		{Key: []byte("x_instana_t"), Value: []byte("000000000000000100000000abcdef12")},
		{Key: []byte("x_instana_s"), Value: []byte("00000000deadbeef")},
		{Key: []byte("x_instana_l_s"), Value: []byte("1")},
	}

	msg := &sarama.ConsumerMessage{
		Headers: headers,
	}

	spanContext, ok := instasarama.SpanContextFromConsumerMessage(msg, c)
	require.True(t, ok)
	assert.Equal(t, instana.SpanContext{
		TraceIDHi: 0x00000001,
		TraceID:   0xabcdef12,
		SpanID:    0xdeadbeef,
		Baggage:   make(map[string]string),
	}, spanContext)

}

func TestSpanContextFromConsumerMessage_NoContext(t *testing.T) {
	examples := []struct {
		Name         string
		Headers      []*sarama.RecordHeader
		HeaderFormat string
	}{
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
	}

	for _, example := range examples {
		t.Run(example.Name, func(t *testing.T) {
			c := instana.InitCollector(&instana.Options{
				AgentClient: alwaysReadyClient{},
				Recorder:    instana.NewTestRecorder(),
			})
			defer instana.ShutdownCollector()

			msg := &sarama.ConsumerMessage{Headers: example.Headers}

			_, ok := instasarama.SpanContextFromConsumerMessage(msg, c)
			assert.False(t, ok)

		})
	}
}

func TestConsumerMessageCarrier_Set_FieldT(t *testing.T) {

	var msg sarama.ConsumerMessage
	c := instasarama.ConsumerMessageCarrier{&msg}

	c.Set(instana.FieldT, "0000000000000001deadbeefdeadbeef")

	var expected []*sarama.RecordHeader

	expected = append(expected, []*sarama.RecordHeader{
		{
			Key:   []byte(instasarama.FieldT),
			Value: []byte("0000000000000001deadbeefdeadbeef"),
		},
	}...)

	assert.Equal(t, expected, msg.Headers)

}

func TestConsumerMessageCarrier_Update_FieldT(t *testing.T) {
	examples := map[string]struct {
		Value    string
		Headers  []*sarama.RecordHeader
		Expected []*sarama.RecordHeader
		EnvVar   string
	}{

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
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {

			msg := sarama.ConsumerMessage{Headers: example.Headers}
			c := instasarama.ConsumerMessageCarrier{&msg}

			c.Set(instana.FieldT, example.Value)
			assert.ElementsMatch(t, example.Expected, msg.Headers)

		})
	}
}

func TestConsumerMessageCarrier_Set_FieldS(t *testing.T) {

	var msg sarama.ConsumerMessage
	var expected []*sarama.RecordHeader
	c := instasarama.ConsumerMessageCarrier{&msg}

	expected = append(expected, []*sarama.RecordHeader{
		{
			Key:   []byte(instasarama.FieldS),
			Value: []byte("00000000deadbeef"),
		},
	}...)

	c.Set(instana.FieldS, "00000000deadbeef")
	assert.Equal(t, expected, msg.Headers)

}

func TestConsumerMessageCarrier_Update_FieldS(t *testing.T) {
	examples := map[string]struct {
		Value    string
		Headers  []*sarama.RecordHeader
		Expected []*sarama.RecordHeader
		EnvVar   string
	}{

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
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {

			msg := sarama.ConsumerMessage{Headers: example.Headers}
			c := instasarama.ConsumerMessageCarrier{&msg}

			c.Set(instana.FieldS, example.Value)
			assert.ElementsMatch(t, example.Expected, msg.Headers)

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
	}

	for _, example := range examples {
		t.Run(example.Name, func(t *testing.T) {

			msg := sarama.ConsumerMessage{Headers: example.Expected}
			c := instasarama.ConsumerMessageCarrier{&msg}

			c.Set(instana.FieldL, example.Value)
			assert.Equal(t, example.Expected, msg.Headers)

		})
	}
}

func TestConsumerMessageCarrier_RemoveAll(t *testing.T) {
	msg := sarama.ConsumerMessage{
		Headers: []*sarama.RecordHeader{
			{Key: []byte("X_CUSTOM_1"), Value: []byte("value1")},
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
			{Key: []byte("x_instana_t"), Value: []byte("000000000000000100000000abcdef12")},
			{Key: []byte("x_instana_s"), Value: []byte("00000000deadbeef")},
			{Key: []byte("x_instana_l_s"), Value: []byte("1")},
		},
	}
	c := instasarama.ConsumerMessageCarrier{&msg}

	assert.Error(t, c.ForeachKey(func(k, v string) error {
		return errors.New("something went wrong")
	}))
}
