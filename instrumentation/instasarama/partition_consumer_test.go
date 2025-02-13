// (c) Copyright IBM Corp. 2023

//go:build go1.17
// +build go1.17

package instasarama_test

import (
	"errors"
	"testing"
	"time"

	"github.com/IBM/sarama"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instasarama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPartitionConsumer_Messages(t *testing.T) {
	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	messages := []*sarama.ConsumerMessage{
		{
			Topic: "instrumented-producer",
			Headers: []*sarama.RecordHeader{
				{
					Key:   []byte("x_instana_t"),
					Value: []byte("0000000000000000000000000abcde12"),
				},
				{
					Key:   []byte("x_instana_s"),
					Value: []byte("00000000deadbeef"),
				},
				{
					Key:   []byte("x_instana_l_s"),
					Value: []byte("1"),
				},
			},
		},
		{Topic: "not-instrumented-producer"},
	}

	pc := &testPartitionConsumer{
		messages: make(chan *sarama.ConsumerMessage, len(messages)),
	}
	for _, msg := range messages {
		pc.messages <- msg
	}
	close(pc.messages)

	wrapped := instasarama.WrapPartitionConsumer(pc, c)

	var collected []*sarama.ConsumerMessage
	timeout := time.After(1 * time.Second)

CONSUMER_LOOP:
	for {
		select {
		case msg, ok := <-wrapped.Messages():
			if !ok {
				break CONSUMER_LOOP
			}
			collected = append(collected, msg)
		case <-timeout:
			t.Fatalf("consuming (*instasarama.PartitionConsumer).Messages() timed out")
		}
	}

	_, open := <-wrapped.Messages()
	assert.False(t, open)
	require.Len(t, collected, len(messages))

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, len(collected))

	t.Run("message with trace context", func(t *testing.T) {
		msg := collected[0]
		assert.Equal(t, "instrumented-producer", msg.Topic)

		span, err := extractAgentSpan(spans[0])
		require.NoError(t, err)

		assert.EqualValues(t, "000000000abcde12", span.TraceID)
		assert.EqualValues(t, "00000000deadbeef", span.ParentID)

		assert.Contains(t, msg.Headers, &sarama.RecordHeader{
			Key:   []byte("x_instana_t"),
			Value: []byte(span.TraceID),
		})
		assert.Contains(t, msg.Headers, &sarama.RecordHeader{
			Key:   []byte("x_instana_s"),
			Value: []byte(span.SpanID),
		})
		assert.Contains(t, msg.Headers, &sarama.RecordHeader{
			Key:   []byte("x_instana_l_s"),
			Value: []byte("1"),
		})
	})

	t.Run("message without trace context", func(t *testing.T) {
		msg := collected[1]
		assert.Equal(t, "not-instrumented-producer", msg.Topic)

		span, err := extractAgentSpan(spans[1])
		require.NoError(t, err)

		assert.NotEmpty(t, span.TraceID)
		assert.Empty(t, span.ParentID)
		assert.EqualValues(t, span.TraceID, span.SpanID)

		assert.ElementsMatch(t, msg.Headers, []*sarama.RecordHeader{
			{
				Key:   []byte("X_INSTANA_T"),
				Value: []byte(span.TraceID),
			},
			{
				Key:   []byte("X_INSTANA_S"),
				Value: []byte(span.SpanID),
			},
			{
				Key:   []byte("X_INSTANA_L_S"),
				Value: []byte("1"),
			},
		})
	})
}

func TestPartitionConsumer_AsyncClose(t *testing.T) {
	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	pc := &testPartitionConsumer{}

	wrapped := instasarama.WrapPartitionConsumer(pc, c)
	wrapped.AsyncClose()

	assert.True(t, pc.Closed)
	assert.True(t, pc.Async)
}

func TestPartitionConsumer_Close(t *testing.T) {
	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	pc := &testPartitionConsumer{}

	wrapped := instasarama.WrapPartitionConsumer(pc, c)
	require.NoError(t, wrapped.Close())

	assert.True(t, pc.Closed)
	assert.False(t, pc.Async)
}

func TestPartitionConsumer_Close_Error(t *testing.T) {
	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	pc := &testPartitionConsumer{
		Error: errors.New("something went wrong"),
	}

	wrapped := instasarama.WrapPartitionConsumer(pc, c)
	assert.Error(t, wrapped.Close())
}

func TestPartitionConsumer_HighWaterMarkOffset(t *testing.T) {
	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	pc := &testPartitionConsumer{
		Offset: 42,
	}

	wrapped := instasarama.WrapPartitionConsumer(pc, c)
	assert.Equal(t, pc.Offset, wrapped.HighWaterMarkOffset())
}

type testPartitionConsumer struct {
	messages chan *sarama.ConsumerMessage
	errors   chan *sarama.ConsumerError

	Offset int64
	Error  error
	Closed bool
	Async  bool
}

func (pc *testPartitionConsumer) Pause() {
	panic("needs to be implemented if testcase is failing")
}

func (pc *testPartitionConsumer) Resume() {
	panic("needs to be implemented if testcase is failing")
}

func (pc *testPartitionConsumer) IsPaused() bool {
	panic("needs to be implemented if testcase is failing")
}

// AsyncClose closes the underlying partition consumer asynchronously
func (pc *testPartitionConsumer) AsyncClose() {
	pc.Closed = true
	pc.Async = true
}

// Close closes the underlying partition consumer
func (pc *testPartitionConsumer) Close() error {
	pc.Closed = true
	pc.Async = false

	return pc.Error
}

// Messages returns a channel of consumer messages of the underlying partition consumer
func (pc *testPartitionConsumer) Messages() <-chan *sarama.ConsumerMessage {
	return pc.messages
}

// Errors returns a channel of consumer errors of the underlying partition consumer
func (pc *testPartitionConsumer) Errors() <-chan *sarama.ConsumerError {
	return pc.errors
}

// HighWaterMarkOffset returns the high water mark offset of the underlying partition consumer
func (pc *testPartitionConsumer) HighWaterMarkOffset() int64 {
	return pc.Offset
}
