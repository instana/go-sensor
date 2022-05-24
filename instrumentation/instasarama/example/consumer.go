// (c) Copyright IBM Corp. 2022

package main

import (
	"fmt"

	"github.com/Shopify/sarama"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instasarama"
	"github.com/opentracing/opentracing-go"
)

// This example demonstrates how to instrument a Kafka consumer using instasarama
// and extract the trace context to ensure continuation. Error handling is omitted for brevity.
func consume(ch chan bool) {
	sensor := instana.NewSensor("my-service")
	brokers := []string{"localhost:9092"}

	conf := sarama.NewConfig()
	conf.Version = sarama.V0_11_0_0

	// create a new instrumented instance of sarama.Consumer
	consumer, _ := instasarama.NewConsumer(brokers, conf, sensor)

	c, _ := consumer.ConsumePartition("test-topic-1", 0, sarama.OffsetNewest)
	defer c.Close()

	for msg := range c.Messages() {
		fmt.Println("Got message", msg)
		processMessage(msg, sensor)
		ch <- true
	}
}

func processMessage(msg *sarama.ConsumerMessage, sensor *instana.Sensor) {
	// extract trace context and start a new span
	parentCtx, _ := instasarama.SpanContextFromConsumerMessage(msg, sensor)

	sp := sensor.Tracer().StartSpan("process-message", opentracing.ChildOf(parentCtx))
	sp.Finish()
}
