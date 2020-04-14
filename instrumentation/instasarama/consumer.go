package instasarama

import (
	"github.com/Shopify/sarama"
	instana "github.com/instana/go-sensor"
)

// Consumer is a wrapper for sarama.Consumer that wraps and returns instrumented
// partition consumers
type Consumer struct {
	sarama.Consumer
	sensor *instana.Sensor
}

// NewConsumer wraps sarama.Consumer instance and instruments its calls
func NewConsumer(c sarama.Consumer, sensor *instana.Sensor) *Consumer {
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

	return NewPartitionConsumer(pc, c.sensor), nil
}
