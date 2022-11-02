// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

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

func TestConsumer_ConsumePartition(t *testing.T) {
	headerFormats := []string{"binary", "string", "both"}

	for _, headerFormat := range headerFormats {
		recorder := instana.NewTestRecorder()
		sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

		messages := make(chan *sarama.ConsumerMessage, 1)
		c := &testConsumer{
			consumers: map[string]*testPartitionConsumer{
				"topic-1": {
					messages: messages,
				},
			},
		}

		if headerFormat == "binary" {
			messages <- &sarama.ConsumerMessage{
				Topic: "topic-1",
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
			}
		} else if headerFormat == "string" {
			messages <- &sarama.ConsumerMessage{
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
					{Key: []byte(instasarama.FieldLS), Value: []byte("1")},
				},
			}
		} else if headerFormat == "both" {
			messages <- &sarama.ConsumerMessage{
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
					{Key: []byte(instasarama.FieldLS), Value: []byte("1")},
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
			}
		} else {
			t.Fatalf("Unexpected header format: %s", headerFormat)
		}

		wrapped := instasarama.WrapConsumer(c, sensor)
		pc, err := wrapped.ConsumePartition("topic-1", 1, 2)
		require.NoError(t, err)

		_, ok := pc.(*instasarama.PartitionConsumer)
		require.True(t, ok)

		require.Empty(t, recorder.GetQueuedSpans())

	selectChannel:
		select {
		case <-pc.Messages():
			break selectChannel
		case <-time.After(1 * time.Second):
			t.Fatalf("partition consumer timed out after 1s")
		}

		assert.NotEmpty(t, recorder.GetQueuedSpans())
	}
}

func TestConsumer_ConsumePartition_Error(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	c := &testConsumer{
		Error: errors.New("something went wrong"),
		consumers: map[string]*testPartitionConsumer{
			"topic-1": {},
		},
	}

	wrapped := instasarama.WrapConsumer(c, sensor)
	_, err := wrapped.ConsumePartition("topic-1", 1, 2)
	assert.Error(t, err)
}

func TestConsumer_Topics(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	c := &testConsumer{
		topics: []string{"topic-1", "topic-2"},
	}

	wrapped := instasarama.WrapConsumer(c, sensor)

	topics, err := wrapped.Topics()
	require.NoError(t, err)

	assert.Equal(t, c.topics, topics)
}

func TestConsumer_Topics_Error(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	c := &testConsumer{
		Error:  errors.New("something went wrong"),
		topics: []string{"topic-1", "topic-2"},
	}

	wrapped := instasarama.WrapConsumer(c, sensor)
	_, err := wrapped.Topics()
	assert.Error(t, err)
}

func TestConsumer_Partitions(t *testing.T) {
	c := &testConsumer{
		partitions: map[string][]int32{
			"topic-1": {1, 2, 3},
		},
	}

	t.Run("existing", func(t *testing.T) {
		recorder := instana.NewTestRecorder()
		sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

		wrapped := instasarama.WrapConsumer(c, sensor)
		partitions, err := wrapped.Partitions("topic-1")
		require.NoError(t, err)

		assert.Equal(t, []int32{1, 2, 3}, partitions)

		assert.Empty(t, recorder.GetQueuedSpans())
	})

	t.Run("non-existing", func(t *testing.T) {
		recorder := instana.NewTestRecorder()
		sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

		wrapped := instasarama.WrapConsumer(c, sensor)
		partitions, err := wrapped.Partitions("topic-2")
		require.NoError(t, err)

		assert.Empty(t, partitions)

		assert.Empty(t, recorder.GetQueuedSpans())
	})
}

func TestConsumer_Partitions_Error(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	c := &testConsumer{
		Error: errors.New("something went wrong"),
		partitions: map[string][]int32{
			"topic-1": {1, 2, 3},
		},
	}

	wrapped := instasarama.WrapConsumer(c, sensor)
	_, err := wrapped.Partitions("topic-1")
	assert.Error(t, err)
}

func TestConsumer_HighWaterMarks(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	c := &testConsumer{
		offsets: map[string]map[int32]int64{
			"topic-1": {
				1: 42,
			},
		},
	}

	wrapped := instasarama.WrapConsumer(c, sensor)
	assert.Equal(t, c.offsets, wrapped.HighWaterMarks())

	assert.Empty(t, recorder.GetQueuedSpans())
}

func TestConsumer_Close(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	c := &testConsumer{}

	wrapped := instasarama.WrapConsumer(c, sensor)
	require.NoError(t, wrapped.Close())

	assert.True(t, c.Closed)
	assert.Empty(t, recorder.GetQueuedSpans())
}

func TestConsumer_Close_Error(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	c := &testConsumer{
		Error: errors.New("something went wrong"),
	}

	wrapped := instasarama.WrapConsumer(c, sensor)
	assert.Error(t, wrapped.Close())
}

type testConsumer struct {
	Closed bool
	Error  error

	topics     []string
	partitions map[string][]int32
	offsets    map[string]map[int32]int64
	consumers  map[string]*testPartitionConsumer
}

func (c *testConsumer) Topics() ([]string, error) {
	return c.topics, c.Error
}

func (c *testConsumer) Partitions(topic string) ([]int32, error) {
	return c.partitions[topic], c.Error
}

func (c *testConsumer) ConsumePartition(topic string, partition int32, offset int64) (sarama.PartitionConsumer, error) {
	return c.consumers[topic], c.Error
}

func (c *testConsumer) HighWaterMarks() map[string]map[int32]int64 {
	return c.offsets
}

func (c *testConsumer) Close() error {
	c.Closed = true
	return c.Error
}
