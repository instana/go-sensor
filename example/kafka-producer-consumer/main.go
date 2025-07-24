// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

//go:build go1.17
// +build go1.17

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/IBM/sarama"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instasarama"
	"github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"
)

var args struct {
	In, Out string
	Brokers []string
}

func main() {
	flag.StringVar(&args.In, "in", "", "Incoming event messages topic")
	flag.StringVar(&args.Out, "out", "", "Event rate messages topic")
	flag.Parse()

	if args.In == "" || args.Out == "" || flag.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] <broker-1>[ <broker-2> ...]\n\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(2)
	}
	args.Brokers = flag.Args()

	// First we create an instance of instana.Sensor, a container that will be used to inject
	// collector into all instrumented methods.
	collector := instana.InitCollector(&instana.Options{
		Service: "doubler",
		Tracer:  instana.DefaultTracerOptions(),
	})

	// First we set up and instrument producers and consumers. Instana uses the headers feature
	// introduced in Kafka v0.11 to propagate trace context. In order to use it, github.com/IBM/sarama
	// producers and consumers need to be explicitly configured with Version = sarama.V0_11_0_0 or above.
	conf := sarama.NewConfig()
	conf.Version = sarama.V0_11_0_0

	// Create and instrument consumer to read messages from incoming topic
	consumer, err := instasarama.NewConsumer(args.Brokers, conf, collector)
	if err != nil {
		log.Fatalln(err)
	}

	c, err := consumer.ConsumePartition(args.In, 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatalln(err)
	}
	defer c.Close()

	// Create and instrument an async producer to publish results
	producer, err := instasarama.NewAsyncProducer(args.Brokers, conf, collector)
	if err != nil {
		log.Fatalln(err)
	}
	defer producer.Close()

	processor := Doubler{
		Out:      args.Out,
		producer: producer,
		sensor:   collector,
	}

	for msg := range c.Messages() {
		if err := processor.ProcessMessage(msg); err != nil {
			log.Println(err)
		}
	}
}

// Doubler is a message processor, that multiplies the value of an incoming message by two
// and publishes it to the Out topic asynchronously
type Doubler struct {
	Out string

	producer sarama.AsyncProducer
	sensor   instana.TracerLogger
}

// ProcessMessage processes messages consumed from Kafka
func (d *Doubler) ProcessMessage(msg *sarama.ConsumerMessage) (err error) {
	var span opentracing.Span

	// Extract trace context injected into the message by Instana instrumentation for github.com/IBM/sarama
	if parentCtx, ok := instasarama.SpanContextFromConsumerMessage(msg, d.sensor); ok {
		span = d.sensor.Tracer().StartSpan("processMessage", opentracing.ChildOf(parentCtx))
		defer func() {
			// Catch any errors occurred while processing the message before finalizing the span.
			// In case there was any issue with (un-)marshaling the payload, the call will me marked
			// as erroneous in the dashboard.
			if err != nil {
				span.LogFields(otlog.Error(err))
			}

			span.Finish()
		}()
	}

	value, err := strconv.Atoi(string(msg.Value))
	if err != nil {
		return fmt.Errorf("malformed incoming value: %q", msg.Value)
	}

	res := &sarama.ProducerMessage{
		Topic: d.Out,
		Value: sarama.StringEncoder(strconv.Itoa(value * 2)),
	}

	// Check whether this call is traced and inject trace context into the outgoing message, so
	// an instrumented consumer on the other side could pick it up and continue
	if span != nil {
		res = instasarama.ProducerMessageWithSpan(res, span)
	}

	d.producer.Input() <- res

	return nil
}
