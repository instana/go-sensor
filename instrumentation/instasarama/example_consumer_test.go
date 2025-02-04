// (c) Copyright IBM Corp. 2023

//go:build go1.17
// +build go1.17

package instasarama_test

import (
	"fmt"

	"github.com/IBM/sarama"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instasarama"
	"github.com/opentracing/opentracing-go"
)

// This example demonstrates how to instrument a Kafka consumer using instasarama
// and extract the trace context to ensure continuation. Error handling is omitted for brevity.
func Example_consumer() {
	c := instana.InitCollector(&instana.Options{
		Service: "my-service",
	})
	defer instana.ShutdownCollector()

	brokers := []string{"localhost:9092"}

	conf := sarama.NewConfig()
	conf.Version = sarama.V0_11_0_0

	// create a new instrumented instance of sarama.Consumer
	consumer, _ := instasarama.NewConsumer(brokers, conf, c)

	cp, _ := consumer.ConsumePartition("test-topic-1", 0, sarama.OffsetNewest)
	defer cp.Close()

	for msg := range cp.Messages() {
		fmt.Println("Got messagge", msg)
		processMessage(msg, c)
	}
}

func processMessage(msg *sarama.ConsumerMessage, sensor instana.TracerLogger) {
	// extract trace context and start a new span
	parentCtx, _ := instasarama.SpanContextFromConsumerMessage(msg, sensor)

	sp := sensor.Tracer().StartSpan("process-message", opentracing.ChildOf(parentCtx))
	defer sp.Finish()

	// process message
}
