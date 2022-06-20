// (c) Copyright IBM Corp. 2022

package instasarama

import (
	"context"

	"github.com/Shopify/sarama"
	instana "github.com/instana/go-sensor"
)

// NewConsumerGroup creates an instrumented sarama.ConsumerGroup
func NewConsumerGroup(addrs []string, groupID string, config *sarama.Config, sensor *instana.Sensor) (sarama.ConsumerGroup, error) {
	c, err := sarama.NewConsumerGroup(addrs, groupID, config)
	if err != nil {
		return nil, err
	}

	return consumerGroup{c, sensor}, nil
}

// NewConsumerGroupFromClient creates an instrumented sarama.ConsumerGroup from sarama.Client
func NewConsumerGroupFromClient(groupID string, client sarama.Client, sensor *instana.Sensor) (sarama.ConsumerGroup, error) {
	c, err := sarama.NewConsumerGroupFromClient(groupID, client)
	if err != nil {
		return nil, err
	}

	return consumerGroup{c, sensor}, nil
}

type consumerGroup struct {
	sarama.ConsumerGroup
	sensor *instana.Sensor
}

func (c consumerGroup) Errors() <-chan error {
	return c.ConsumerGroup.Errors()
}

func (c consumerGroup) Close() error {
	return c.ConsumerGroup.Close()
}

func (c consumerGroup) Consume(ctx context.Context, topics []string, handler sarama.ConsumerGroupHandler) error {
	if _, ok := handler.(*ConsumerGroupHandler); ok {
		return c.ConsumerGroup.Consume(ctx, topics, handler)
	}

	return c.ConsumerGroup.Consume(ctx, topics, WrapConsumerGroupHandler(handler, c.sensor))
}
