// (c) Copyright IBM Corp. 2022

package main

import (
	"github.com/Shopify/sarama"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instasarama"
	"github.com/opentracing/opentracing-go/ext"
)

// This example demonstrates how to instrument a sync Kafka producer using instasarama.
// Error handling is omitted for brevity.
func produce() {
	sensor := instana.NewSensor("my-service")
	brokers := []string{"localhost:9092"}

	config := sarama.NewConfig()
	// sarama requires Producer.Return.Successes to be set for sync producers
	config.Producer.Return.Successes = true
	// enable the use record headers added in kafka v0.11.0 and used to propagate
	// trace context
	config.Version = sarama.V0_11_0_0

	// create a new instrumented instance of sarama.SyncProducer
	producer, _ := instasarama.NewSyncProducer(brokers, config, sensor)

	// start a new entry span
	sp := sensor.Tracer().StartSpan("my-producing-method")
	ext.SpanKind.Set(sp, "entry")

	msg := &sarama.ProducerMessage{
		Topic:  "test-topic-1",
		Offset: sarama.OffsetNewest,
		Value:  sarama.StringEncoder("I am a message"),
		// ...
	}

	// inject the span before passing the message to producer
	msg = instasarama.ProducerMessageWithSpan(msg, sp)

	// pass it to the producer
	producer.SendMessage(msg)
}
