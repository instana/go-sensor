// (c) Copyright IBM Corp. 2023

//go:build go1.16
// +build go1.16

package instaamqp091

import (
	"net/url"

	instana "github.com/instana/go-sensor"
	amqp "github.com/rabbitmq/amqp091-go"

	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
)

const (
	consume   = "consume"
	publish   = "publish"
	operation = "rabbitmq"
)

// PubCons contains all methods that we want to instrument from the amqp library
type PubCons interface {
	Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
	Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error)
}

// AmqpChannel is a wrapper around the amqp.Channel object and contains all the relevant information to be tracked
type AmqpChannel struct {
	url    string
	pc     PubCons
	sensor *instana.Sensor
}

// Publish replaces the original amqp.Channel.Publish method in order to collect the relevant data to be tracked
func (c AmqpChannel) Publish(entrySpan ot.Span, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	opts := []ot.StartSpanOption{
		ext.SpanKindProducer,
		ot.ChildOf(entrySpan.Context()),
		ot.Tags{
			"rabbitmq.exchange": exchange,
			"rabbitmq.key":      key,
			"rabbitmq.sort":     publish,
			"rabbitmq.address":  c.url,
		},
	}

	logger := c.sensor.Logger()
	tracer := c.sensor.Tracer()
	sp := tracer.StartSpan(operation, opts...)

	if msg.Headers == nil {
		msg.Headers = amqp.Table{}
	}

	err := tracer.Inject(sp.Context(), ot.TextMap, &messageCarrier{msg.Headers, logger})

	if err != nil {
		logger.Debug(err)
	}

	res := c.pc.Publish(exchange, key, mandatory, immediate, msg)
	resCopy := res

	if resCopy != nil {
		errorText := resCopy.Error()
		sp.SetTag("rabbitmq.error", errorText)
		sp.LogFields(otlog.Object("error", errorText))
	}
	sp.Finish()

	return res
}

// Consume replaces the original amqp.Channel.Consume method in order to collect the relevant data to be tracked
func (c AmqpChannel) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	deliveryChan, err := c.pc.Consume(queue, consumer, autoAck, exclusive, noLocal, noWait, args)

	if err == nil {
		// Creates a pipe channel that receives the read data from the original channel
		pipeCh := make(chan amqp.Delivery, cap(deliveryChan))

		go func() {
			for {
				deliveryData, more := <-deliveryChan

				if !more {
					close(pipeCh)
					return
				}

				c.consumeMessage(pipeCh, deliveryData, queue)
			}
		}()

		return pipeCh, err
	}

	return deliveryChan, err
}

func (c AmqpChannel) consumeMessage(pipeCh chan amqp.Delivery, deliveryData amqp.Delivery, queue string) {
	opts := []ot.StartSpanOption{
		ext.SpanKindConsumer,
		ot.Tags{
			"rabbitmq.exchange": deliveryData.Exchange,
			"rabbitmq.key":      queue,
			"rabbitmq.sort":     consume,
			"rabbitmq.address":  c.url,
		},
	}

	logger := c.sensor.Logger()
	tracer := c.sensor.Tracer()
	sc, err := tracer.Extract(ot.TextMap, &messageCarrier{deliveryData.Headers, logger})

	if err != nil {
		logger.Debug(err)
	}

	opts = append(opts, ot.ChildOf(sc))
	sp := tracer.StartSpan(operation, opts...)

	err = tracer.Inject(sp.Context(), ot.TextMap, &messageCarrier{deliveryData.Headers, logger})

	if err != nil {
		logger.Debug(err)
	}

	sp.Finish()
	pipeCh <- deliveryData
}

// WrapChannel returns the AmqpChannel, which is Instana's wrapper around amqp.Channel
func WrapChannel(sensor *instana.Sensor, ch PubCons, serverUrl string) *AmqpChannel {
	sUrl := ""
	urlObj, err := url.Parse(serverUrl)

	if err == nil {
		sUrl = urlObj.Scheme + "://" + urlObj.Host
	}

	return &AmqpChannel{sUrl, ch, sensor}
}
