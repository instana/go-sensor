// (c) Copyright IBM Corp. 2023

//go:build go1.17
// +build go1.17

package main

import (
	"fmt"

	"github.com/IBM/sarama"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instasarama"
	"github.com/opentracing/opentracing-go"
)

// This example demonstrates how to instrument a Kafka consumer using instasarama
// and extract the trace context to ensure continuation. Error handling is omitted for brevity.
func consume(ch chan bool) {
	collector := instana.InitCollector(&instana.Options{
		Service: "my-service",
	})
	brokers := []string{"localhost:9092"}

	conf := sarama.NewConfig()
	conf.Version = sarama.V0_11_0_0

	// create a new instrumented instance of sarama.Consumer
	consumer, _ := instasarama.NewConsumer(brokers, conf, collector)

	c, _ := consumer.ConsumePartition("test-topic-1", 0, sarama.OffsetNewest)
	defer c.Close()

	for msg := range c.Messages() {
		fmt.Println("Got message: ", string(msg.Value))
		processMessage(msg, collector)
		ch <- true
	}
}

func processMessage(msg *sarama.ConsumerMessage, collector instana.TracerLogger) {
	// extract trace context and start a new span
	parentCtx, _ := instasarama.SpanContextFromConsumerMessage(msg, collector)

	sp := collector.Tracer().StartSpan("process-message", opentracing.ChildOf(parentCtx))
	sp.Finish()
}
