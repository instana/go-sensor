// (c) Copyright IBM Corp. 2023

//go:build go1.16
// +build go1.16

package instaamqp091_test

import (
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instaamqp091"
	"github.com/opentracing/opentracing-go/ext"
	amqp "github.com/rabbitmq/amqp091-go"
)

func Example_publisher() {
	exchangeName := "my-exchange"
	url := "amqp://guest:guest@localhost:5672/"

	// Create the Instana sensor
	sensor := instana.NewSensor("rabbitmq-client")

	c, err := amqp.Dial(url)
	failOnError(err, "Could not connect to the server")
	defer c.Close()

	ch, err := c.Channel()
	failOnError(err, "Could not acquire the channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(exchangeName, "fanout", true, false, false, false, nil)
	failOnError(err, "Could not declare the exchange")

	// There must be a new entry span per publish call.
	// In the most common cases, creating an entry span manually is not needed, as the entry span is originated from an
	// incoming HTTP client call.
	entrySpan := sensor.Tracer().StartSpan("my-publishing-method")
	ext.SpanKind.Set(entrySpan, ext.SpanKindRPCServerEnum)

	instaCh := instaamqp091.WrapChannel(sensor, ch, url)

	// Use the Instana `Publish` method with the same arguments as the original `Publish` method, with the additional
	// entrySpan argument.
	err = instaCh.Publish(entrySpan, exchangeName, "", false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte("My published message"),
	})

	failOnError(err, "Error publishing the message")
	entrySpan.Finish()
}
