// (c) Copyright IBM Corp. 2023

package instasarama

import (
	"github.com/IBM/sarama"
	instana "github.com/instana/go-sensor"
)

// Consumer is a wrapper for sarama.Consumer that wraps and returns instrumented
// partition consumers
type Consumer struct {
	sarama.Consumer
	sensor instana.TracerLogger
}

// NewConsumer creates a new consumer using the given broker addresses and configuration, and
// instruments its calls
func NewConsumer(addrs []string, config *sarama.Config, sensor instana.TracerLogger) (sarama.Consumer, error) {
	c, err := sarama.NewConsumer(addrs, config)
	if err != nil {
		return c, err
	}

	return WrapConsumer(c, sensor), nil
}

// NewConsumerFromClient creates a new consumer using the given client and instruments its calls
func NewConsumerFromClient(client sarama.Client, sensor instana.TracerLogger) (sarama.Consumer, error) {
	c, err := sarama.NewConsumerFromClient(client)
	if err != nil {
		return c, err
	}

	return WrapConsumer(c, sensor), nil
}

// WrapConsumer wraps an existing sarama.Consumer instance and instruments its calls. To initialize
// a new instance of sarama.Consumer use instasarama.NewConsumer() and instasarama.NewConsumerFromclient()
// convenience methods instead
func WrapConsumer(c sarama.Consumer, sensor instana.TracerLogger) *Consumer {
	return &Consumer{
		Consumer: c,
		sensor:   sensor,
	}
}

// ConsumePartition instruments and returns the partition consumer returned by undelying sarama.Consumer
func (c *Consumer) ConsumePartition(topic string, partition int32, offset int64) (sarama.PartitionConsumer, error) {
	pc, err := c.Consumer.ConsumePartition(topic, partition, offset)
	if err != nil {
		return nil, err
	}

	return WrapPartitionConsumer(pc, c.sensor), nil
}
