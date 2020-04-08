package instasarama_test

import (
	"errors"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instasarama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPartitionConsumer_Messages(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	messages := []*sarama.ConsumerMessage{
		{Topic: "test-topic-1"},
		{Topic: "test-topic-2"},
	}

	pc := &testPartitionConsumer{
		messages: make(chan *sarama.ConsumerMessage, len(messages)),
	}
	for _, msg := range messages {
		pc.messages <- msg
	}
	close(pc.messages)

	wrapped := instasarama.NewPartitionConsumer(pc, sensor)

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

	assert.Equal(t, messages, collected)
}

func TestPartitionConsumer_AsyncClose(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	pc := &testPartitionConsumer{}

	wrapped := instasarama.NewPartitionConsumer(pc, sensor)
	wrapped.AsyncClose()

	assert.True(t, pc.Closed)
	assert.True(t, pc.Async)
}

func TestPartitionConsumer_Close(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	pc := &testPartitionConsumer{}

	wrapped := instasarama.NewPartitionConsumer(pc, sensor)
	require.NoError(t, wrapped.Close())

	assert.True(t, pc.Closed)
	assert.False(t, pc.Async)
}

func TestPartitionConsumer_Close_Error(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	pc := &testPartitionConsumer{
		Error: errors.New("something went wrong"),
	}

	wrapped := instasarama.NewPartitionConsumer(pc, sensor)
	assert.Error(t, wrapped.Close())
}

func TestPartitionConsumer_HighWaterMarkOffset(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	pc := &testPartitionConsumer{
		Offset: 42,
	}

	wrapped := instasarama.NewPartitionConsumer(pc, sensor)
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
