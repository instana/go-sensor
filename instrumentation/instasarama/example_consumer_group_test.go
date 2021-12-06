// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instasarama_test

import (
	"context"

	"github.com/Shopify/sarama"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instasarama"
	"github.com/opentracing/opentracing-go"
)

// This example demonstrates how to instrument a Kafka consumer group using instasarama
// and extract the trace context to ensure continuation. Error handling is omitted for brevity.
func Example_consumerGroup() {
	sensor := instana.NewSensor("my-service")
	brokers := []string{"localhost:9092"}
	topics := []string{"records", "more-records"}

	conf := sarama.NewConfig()
	conf.Version = sarama.V0_11_0_0

	client, _ := sarama.NewConsumerGroup(brokers, "my-service-consumers", conf)
	defer client.Close()

	ctx := context.Background()
	consumer := instasarama.WrapConsumerGroupHandler(&Consumer{sensor: sensor}, sensor)

	// start consuming
	for {
		_ = client.Consume(ctx, topics, consumer)

		// ...
	}
}

type Consumer struct {
	sensor *instana.Sensor
}

func (*Consumer) Setup(sarama.ConsumerGroupSession) error {
	// setup consumer
	return nil
}

func (*Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	// cleanup consumer
	return nil
}

func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		c.processMessage(msg)
		session.MarkMessage(msg, "")
	}

	return nil
}

func (c *Consumer) processMessage(msg *sarama.ConsumerMessage) {
	// extract trace context and start a new span
	parentCtx, _ := instasarama.SpanContextFromConsumerMessage(msg, c.sensor)

	sp := c.sensor.Tracer().StartSpan("process-message", opentracing.ChildOf(parentCtx))
	defer sp.Finish()

	// process message
}
