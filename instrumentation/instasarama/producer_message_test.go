package instasarama_test

import (
	"testing"

	"github.com/Shopify/sarama"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instasarama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProducerMessage_ProducerMessageWithSpan(t *testing.T) {
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
		{Key: []byte(instasarama.FieldL), Value: []byte{0x01}},
		{
			Key: []byte(instasarama.FieldC),
			Value: instasarama.PackTraceContextHeader(
				instana.FormatID(spans[0].TraceID),
				instana.FormatID(spans[0].SpanID),
			),
		},
	}

	assert.ElementsMatch(t, expected, pm.Headers)
}
