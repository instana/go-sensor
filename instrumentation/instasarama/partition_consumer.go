package instasarama

import (
	"github.com/Shopify/sarama"
	instana "github.com/instana/go-sensor"
)

// PartitionConsumer is a wrapper for sarama.PartitionConsumer that instruments its calls using
// provided instana.Sensor
type PartitionConsumer struct {
	consumer sarama.PartitionConsumer
	sensor   *instana.Sensor
	messages chan *sarama.ConsumerMessage
}

// NewPartitionConsumer wraps sarama.PartitionConsumer instance and instruments its calls
func NewPartitionConsumer(c sarama.PartitionConsumer, sensor *instana.Sensor) *PartitionConsumer {
	pc := &PartitionConsumer{
		consumer: c,
		sensor:   sensor,
		messages: make(chan *sarama.ConsumerMessage, cap(c.Messages())),
	}

	go pc.consumeMessages()

	return pc
}

// AsyncClose closes the underlying partition consumer asynchronously
func (pc *PartitionConsumer) AsyncClose() {
	pc.consumer.AsyncClose()
}

// Close closes the underlying partition consumer
func (pc *PartitionConsumer) Close() error {
	return pc.consumer.Close()
}

// Messages returns a channel of consumer messages of the underlying partition consumer
func (pc *PartitionConsumer) Messages() <-chan *sarama.ConsumerMessage {
	return pc.messages
}

// Errors returns a channel of consumer errors of the underlying partition consumer
func (pc *PartitionConsumer) Errors() <-chan *sarama.ConsumerError {
	return pc.consumer.Errors()
}

// HighWaterMarkOffset returns the high water mark offset of the underlying partition consumer
func (pc *PartitionConsumer) HighWaterMarkOffset() int64 {
	return pc.consumer.HighWaterMarkOffset()
}

func (pc *PartitionConsumer) consumeMessages() {
	for msg := range pc.consumer.Messages() {
		pc.messages <- msg
	}
	close(pc.messages)
}
