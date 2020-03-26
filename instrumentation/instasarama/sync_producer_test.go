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

func TestSyncProducer_SendMessage(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	parent := sensor.Tracer().StartSpan("test-span")

	p := &testSyncProducer{}
	wrapped := instasarama.NewSyncProducer(p, sensor)

	_, _, err := wrapped.SendMessage(
		instasarama.ProducerMessageWithSpan(&sarama.ProducerMessage{Topic: "test-topic"}, parent),
	)
	require.NoError(t, err)

	parent.Finish()

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 2)

	pSpan, err := extractAgentSpan(spans[1])
	require.NoError(t, err)

	cSpan, err := extractAgentSpan(spans[0])
	require.NoError(t, err)

	assert.Equal(t, 0, cSpan.Ec)
	assert.EqualValues(t, instana.ExitSpanKind, cSpan.Kind)

	assert.Equal(t, agentKafkaSpanData{
		Service: "test-topic",
		Access:  "send",
	}, cSpan.Data.Kafka)

	require.Len(t, p.Messages, 1)
	assert.Contains(t, p.Messages[0].Headers, sarama.RecordHeader{
		Key:   []byte("X_INSTANA_C"),
		Value: instasarama.PackTraceContextHeader(instana.FormatID(cSpan.TraceID), instana.FormatID(cSpan.SpanID)),
	})
	assert.Contains(t, p.Messages[0].Headers, sarama.RecordHeader{
		Key:   []byte("X_INSTANA_L"),
		Value: instasarama.PackTraceLevelHeader("1"),
	})

	assert.Equal(t, pSpan.TraceID, cSpan.TraceID)
	assert.Equal(t, pSpan.SpanID, cSpan.ParentID)
	assert.NotEqual(t, pSpan.SpanID, cSpan.SpanID)
}

func TestSyncProducer_SendMessage_NoTraceContext(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	p := &testSyncProducer{}
	wrapped := instasarama.NewSyncProducer(p, sensor)

	_, _, err := wrapped.SendMessage(&sarama.ProducerMessage{
		Topic: "test-topic",
	})
	require.NoError(t, err)

	spans := recorder.GetQueuedSpans()
	assert.Empty(t, spans)

	require.Len(t, p.Messages, 1)
	assert.Empty(t, p.Messages[0].Headers)
}

func TestSyncProducer_SendMessage_Error(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	p := &testSyncProducer{
		Error: errors.New("something went wrong"),
	}
	wrapped := instasarama.NewSyncProducer(p, sensor)

	parent := sensor.Tracer().StartSpan("test-span")
	_, _, err := wrapped.SendMessage(
		instasarama.ProducerMessageWithSpan(&sarama.ProducerMessage{Topic: "test-topic"}, parent),
	)
	parent.Finish()

	assert.Error(t, err)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 2)

	span, err := extractAgentSpan(spans[0])
	require.NoError(t, err)

	assert.Equal(t, 1, span.Ec)
	assert.EqualValues(t, instana.ExitSpanKind, span.Kind)

	assert.Equal(t, agentKafkaSpanData{
		Service: "test-topic",
		Access:  "send",
	}, span.Data.Kafka)
}

func TestSyncProducer_SendMessages(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	p := &testSyncProducer{}
	wrapped := instasarama.NewSyncProducer(p, sensor)

	_ = wrapped
}

func TestSyncProducer_Close(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	p := &testSyncProducer{}
	wrapped := instasarama.NewSyncProducer(p, sensor)
	wrapped.Close()

	assert.True(t, p.Closed)
}

type testSyncProducer struct {
	Error    error
	Messages []*sarama.ProducerMessage
	Closed   bool
}

func (p *testSyncProducer) SendMessage(msg *sarama.ProducerMessage) (partition int32, offset int64, err error) {
	p.Messages = append(p.Messages, msg)

	return 0, 0, p.Error
}

func (p *testSyncProducer) SendMessages(msgs []*sarama.ProducerMessage) error {
	p.Messages = append(p.Messages, msgs...)

	return p.Error
}

func (p *testSyncProducer) Close() error {
	p.Closed = true
	return p.Error
}
