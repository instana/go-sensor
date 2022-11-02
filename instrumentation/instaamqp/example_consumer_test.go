// (c) Copyright IBM Corp. 2022

package instaamqp_test

import (
	"fmt"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instaamqp"

	"github.com/streadway/amqp"
)

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

	for d := range msgs {
		fmt.Println("Got a message:", string(d.Body))
	}
}
