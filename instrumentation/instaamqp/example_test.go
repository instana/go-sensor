// (c) Copyright IBM Corp. 2022

package instaamqp_test

import (
	"fmt"
	"log"
	"time"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instaamqp"

	"github.com/opentracing/opentracing-go/ext"
	"github.com/streadway/amqp"
)

const exchangeName = "the-exchange4"
const queueName = "the-queue1"

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func publish(numMessages int) {
	sensor := instana.NewSensor("rabbitmq-client")
	url := "amqp://guest:guest@localhost:5672/"

	conn, err := amqp.Dial(url)
	failOnError(err, "Could not connect to the server")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Could not acquire the channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(exchangeName, "fanout", true, false, false, false, nil)
	failOnError(err, "Could not declare the exchange")

	instaCh := instaamqp.WrapChannel(sensor, ch, url)

	for i := 0; i < numMessages; i++ {
		time.Sleep(time.Millisecond * 500)
		// There must be one span per publish call
		entrySpan := sensor.Tracer().StartSpan("testing")
		ext.SpanKind.Set(entrySpan, ext.SpanKindRPCServerEnum)

		err = instaCh.Publish(entrySpan, exchangeName, "", false, false, amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(fmt.Sprintf("Hello from the Example #%d", i+1)),
		})
		entrySpan.Finish()
		failOnError(err, fmt.Sprintf("Hello from the Example #%d", i+1))
		fmt.Printf("Sent message #%d\n", i+1)
	}

	time.Sleep(time.Second * 2)
	log.Println("Released after 5 seconds")
}

func consume() {
	url := "amqp://guest:guest@localhost:5672/"
	sensor := instana.NewSensor("rabbitmq-client")
	conn, err := amqp.Dial(url)

	failOnError(err, "Could not connect to the server")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Could not acquire the channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(exchangeName, "fanout", true, false, false, false, nil)
	failOnError(err, "Could not declare the exchange")

	q, err := ch.QueueDeclare(queueName, false, false, true, false, nil)
	failOnError(err, "Could not declare queue")

	err = ch.QueueBind(q.Name, "", exchangeName, false, nil)
	failOnError(err, "Could not bind the queue to the exchange")

	// Instana wrapper of amqp.Channel
	instaCh := instaamqp.WrapChannel(sensor, ch, url)

	msgs, err := instaCh.Consume(q.Name, "", true, false, false, false, nil)
	failOnError(err, "Could not consume messages")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			fmt.Println("message!", string(d.Body))
		}
	}()

	<-forever
}

func AppExample() {
	go consume()

	// give consume some time to be ready to receive messages
	time.Sleep(time.Second * 5)

	publish(1)
}
