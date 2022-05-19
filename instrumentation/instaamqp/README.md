Instana instrumentation for streadway-amqp
==========================================

This module contains instrumentation code for RabbitMQ clients written with [`streadway-amqp`](https://pkg.go.dev/github.com/streadway/amqp@v1.0.0).

[![GoDoc](https://pkg.go.dev/badge/github.com/instana/go-sensor/instrumentation/instaamqp)][godoc]

Installation
------------

To add the module to your `go.mod` file run the following command in your project directory:

```bash
$ go get github.com/instana/go-sensor/instrumentation/instaamqp
```

Usage
-----

`instaamqp` offers a function wrapper around [`amqp.Channel`][amqp.Channel] that returns an `instaamqp.AmqpChannel` instance.
This Instana object provides instrumentation for the `amqp.Channel.Publish` and `amqp.Channel.Consume` methods, that are
responsible for tracing data from messages sent and received.

For any other `amqp.Channel` methods, the original `amqp.Channel` instance can be normally used.

A publisher example:

```go
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

	// There must be an entry span per publish call.
	// In most common cases, creating an entry span manually is not needed, as the entry span is originated from an
	// incoming HTTP client call.
	entrySpan := sensor.Tracer().StartSpan("my-publishing-method")
	ext.SpanKind.Set(entrySpan, ext.SpanKindRPCServerEnum)

	// We wrap the original amqp.Channel.Publish and amqp.Channel.Consume methods into an Instana object.
	instaCh := instaamqp.WrapChannel(sensor, ch, url)

	// Use the Instana `Publish` method with the same arguments as the original `Publish` method, with the additional
	// `entrySpan` argument. That's it!
	err = instaCh.Publish(entrySpan, exchangeName, "", false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(fmt.Sprintf("My published message")),
	})

	failOnError(err, "Error publishing the message")
	entrySpan.Finish()
}
```

A consumer example:

```go
func Example_consumer() {
	exchangeName := "my-exchange"
	queueName := "my-queue"
	url := "amqp://guest:guest@localhost:5672/"

	sensor := instana.NewSensor("rabbitmq-client")

	c, err := amqp.Dial(url)
	failOnError(err, "Could not connect to the server")
	defer c.Close()

	ch, err := c.Channel()
	failOnError(err, "Could not acquire the channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(exchangeName, "fanout", true, false, false, false, nil)
	failOnError(err, "Could not declare the exchange")

	q, err := ch.QueueDeclare(queueName, false, false, true, false, nil)
	failOnError(err, "Could not declare queue")

	err = ch.QueueBind(q.Name, "", exchangeName, false, nil)
	failOnError(err, "Could not bind the queue to the exchange")

	instaCh := instaamqp.WrapChannel(sensor, ch, url)

	// Use the Instana `Consume` method with the same arguments as the original `Consume` method.
	msgs, err := instaCh.Consume(q.Name, "", true, false, false, false, nil)
	failOnError(err, "Could not consume messages")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			fmt.Println("Got a message:", string(d.Body))
		}
	}()

	<-forever
}
```


See the [`instaamqp` package documentation][godoc] for detailed examples.


[godoc]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instaamqp
[instaamqp.WrapChannel]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instaamqp#WrapChannel
[amqp.Channel]: https://pkg.go.dev/github.com/streadway/amqp@v1.0.0#Channel
