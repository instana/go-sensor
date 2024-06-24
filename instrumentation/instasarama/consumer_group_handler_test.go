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

func TestConsumerGroupHandler_ConsumeClaim(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	messages := []*sarama.ConsumerMessage{
		{
			Topic: "topic-1",
			Headers: []*sarama.RecordHeader{
				{
					Key:   []byte("x_instana_t"),
					Value: []byte("0000000000000000000000000abcde12"),
				},
				{
					Key:   []byte("x_instana_s"),
					Value: []byte("00000000deadbeef"),
				},
				{Key: []byte("x_instana_l_s"), Value: []byte("1")},
			},
		},
		{Topic: "topic-2"},
		{
			Topic: "topic-3",
			Headers: []*sarama.RecordHeader{
				{
					Key: []byte("x_instana_c"),
					Value: []byte{
						// trace id
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
						0x00, 0x00, 0x00, 0x00, 0x0a, 0xbc, 0xde, 0x12,
						// span id
						0x00, 0x00, 0x00, 0x00, 0xde, 0xad, 0xbe, 0xef,
					},
				},
				{Key: []byte("x_instana_l"), Value: []byte{0x01}},
			},
		},
	}

	claim := &testConsumerGroupClaim{
		messages: make(chan *sarama.ConsumerMessage, len(messages)),
	}
	for _, msg := range messages {
		claim.messages <- msg
	}
	close(claim.messages)

	sess := &testConsumerGroupSession{}

	h := &testConsumerGroupHandler{}
	wrapped := instasarama.WrapConsumerGroupHandler(h, sensor)

	require.NoError(t, wrapped.ConsumeClaim(sess, claim))

	assert.Equal(t, messages, h.Messages)      // all messages were processed
	assert.Equal(t, h.Messages, sess.Messages) // all messages are marked

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 3)

	t.Run("span for message with trace headers", func(t *testing.T) {
		span, err := extractAgentSpan(spans[0])
		require.NoError(t, err)

		assert.EqualValues(t, "000000000abcde12", span.TraceID)
		assert.EqualValues(t, "00000000deadbeef", span.ParentID)
		assert.NotEqual(t, span.ParentID, span.SpanID)

		assert.Equal(t, span.Name, "kafka")
		assert.EqualValues(t, span.Kind, instana.EntrySpanKind)

		assert.Equal(t, agentKafkaSpanData{
			Service: "topic-1",
			Access:  "consume",
		}, span.Data.Kafka)

		assert.Contains(t, h.Messages[0].Headers, &sarama.RecordHeader{
			Key:   []byte("x_instana_t"),
			Value: []byte(span.TraceID),
		})
		assert.Contains(t, h.Messages[0].Headers, &sarama.RecordHeader{
			Key:   []byte("x_instana_s"),
			Value: []byte(span.SpanID),
		})
		assert.Contains(t, h.Messages[0].Headers, &sarama.RecordHeader{
			Key:   []byte("x_instana_l_s"),
			Value: []byte("1"),
		})
	})

	t.Run("span for message without trace headers", func(t *testing.T) {
		span, err := extractAgentSpan(spans[1])
		require.NoError(t, err)

		assert.NotEmpty(t, span.TraceID)
		assert.Empty(t, span.ParentID)
		assert.EqualValues(t, span.TraceID, span.SpanID)

		assert.Equal(t, span.Name, "kafka")
		assert.EqualValues(t, span.Kind, instana.EntrySpanKind)

		assert.Equal(t, agentKafkaSpanData{
			Service: "topic-2",
			Access:  "consume",
		}, span.Data.Kafka)

		assert.Contains(t, h.Messages[1].Headers, &sarama.RecordHeader{
			Key:   []byte("X_INSTANA_T"),
			Value: []byte(span.TraceID),
		})
		assert.Contains(t, h.Messages[1].Headers, &sarama.RecordHeader{
			Key:   []byte("X_INSTANA_S"),
			Value: []byte(span.SpanID),
		})
		assert.Contains(t, h.Messages[1].Headers, &sarama.RecordHeader{
			Key:   []byte("X_INSTANA_L_S"),
			Value: []byte("1"),
		})
	})

	t.Run("span for message with binary trace headers", func(t *testing.T) {
		// Binary headers are no longer supported.
		// The expected behavior, in the absence of string headers, should match the behavior observed when there are no trace headers at all.
		span, err := extractAgentSpan(spans[2])
		require.NoError(t, err)

		// Verify that the processed trace and parent IDs do not match the values passed in the headers (since they were binary headers).
		assert.NotEqualValues(t, "000000000abcde12", span.TraceID)
		assert.NotEqualValues(t, "00000000deadbeef", span.ParentID)
		assert.EqualValues(t, span.TraceID, span.SpanID)

		assert.Equal(t, span.Name, "kafka")
		assert.EqualValues(t, span.Kind, instana.EntrySpanKind)

		assert.Equal(t, agentKafkaSpanData{
			Service: "topic-3",
			Access:  "consume",
		}, span.Data.Kafka)

		assert.Contains(t, h.Messages[2].Headers, &sarama.RecordHeader{
			Key:   []byte("X_INSTANA_T"),
			Value: []byte(span.TraceID),
		})
		assert.Contains(t, h.Messages[2].Headers, &sarama.RecordHeader{
			Key:   []byte("X_INSTANA_S"),
			Value: []byte(span.SpanID),
		})
		assert.Contains(t, h.Messages[2].Headers, &sarama.RecordHeader{
			Key:   []byte("X_INSTANA_L_S"),
			Value: []byte("1"),
		})
	})
}

func TestConsumerGroupHandler_Setup(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	h := &testConsumerGroupHandler{}
	wrapped := instasarama.WrapConsumerGroupHandler(h, sensor)

	require.NoError(t, wrapped.Setup(&testConsumerGroupSession{}))
	assert.True(t, h.SetupCalled)

	assert.Empty(t, recorder.GetQueuedSpans())
}

func TestConsumerGroupHandler_Setup_Error(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	h := &testConsumerGroupHandler{
		Error: errors.New("something went wrong"),
	}
	wrapped := instasarama.WrapConsumerGroupHandler(h, sensor)

	assert.Error(t, wrapped.Setup(&testConsumerGroupSession{}))

	assert.Empty(t, recorder.GetQueuedSpans())
}

func TestConsumerGroupHandler_Cleanup(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	h := &testConsumerGroupHandler{}
	wrapped := instasarama.WrapConsumerGroupHandler(h, sensor)

	require.NoError(t, wrapped.Cleanup(&testConsumerGroupSession{}))
	assert.True(t, h.CleanupCalled)

	assert.Empty(t, recorder.GetQueuedSpans())
}

func TestConsumerGroupHandler_Cleanup_Error(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	h := &testConsumerGroupHandler{
		Error: errors.New("something went wrong"),
	}
	wrapped := instasarama.WrapConsumerGroupHandler(h, sensor)

	assert.Error(t, wrapped.Cleanup(&testConsumerGroupSession{}))

	assert.Empty(t, recorder.GetQueuedSpans())
}

type testConsumerGroupHandler struct {
	Error error

	SetupCalled, CleanupCalled bool
	Messages                   []*sarama.ConsumerMessage
}

func (h *testConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	h.SetupCalled = true
	return h.Error
}

func (h *testConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	h.CleanupCalled = true
	return h.Error
}

func (h *testConsumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		h.Messages = append(h.Messages, msg)
		sess.MarkMessage(msg, "")
	}

	return h.Error
}

type testConsumerGroupClaim struct {
	messages chan *sarama.ConsumerMessage
}

func (c *testConsumerGroupClaim) Topic() string                            { return "test-topic" }
func (c *testConsumerGroupClaim) Partition() int32                         { return 0 }
func (c *testConsumerGroupClaim) InitialOffset() int64                     { return 0 }
func (c *testConsumerGroupClaim) HighWaterMarkOffset() int64               { return 100 }
func (c *testConsumerGroupClaim) Messages() <-chan *sarama.ConsumerMessage { return c.messages }

type testConsumerGroupSession struct {
	Messages []*sarama.ConsumerMessage
}

func (s *testConsumerGroupSession) Commit()                    {}
func (s *testConsumerGroupSession) Claims() map[string][]int32 { return nil }
func (s *testConsumerGroupSession) MemberID() string           { return "" }
func (s *testConsumerGroupSession) GenerationID() int32        { return 0 }
func (s *testConsumerGroupSession) MarkOffset(topic string, partition int32, offset int64, metadata string) {
}
func (s *testConsumerGroupSession) ResetOffset(topic string, partition int32, offset int64, metadata string) {
}
func (s *testConsumerGroupSession) MarkMessage(msg *sarama.ConsumerMessage, metadata string) {
	s.Messages = append(s.Messages, msg)
}
func (s *testConsumerGroupSession) Context() context.Context { return context.Background() }
