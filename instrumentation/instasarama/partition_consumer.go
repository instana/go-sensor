// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instasarama

import (
	"github.com/Shopify/sarama"
	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// PartitionConsumer is a wrapper for sarama.PartitionConsumer that instruments its calls using
// provided instana.Sensor
type PartitionConsumer struct {
	sarama.PartitionConsumer
	sensor   *instana.Sensor
	messages chan *sarama.ConsumerMessage
}

// WrapPartitionConsumer wraps sarama.PartitionConsumer instance and instruments its calls
func WrapPartitionConsumer(c sarama.PartitionConsumer, sensor *instana.Sensor) *PartitionConsumer {
	pc := &PartitionConsumer{
		PartitionConsumer: c,
		sensor:            sensor,
		messages:          make(chan *sarama.ConsumerMessage),
	}

	go pc.consumeMessages()

	return pc
}

// Messages returns a channel of consumer messages of the underlying partition consumer
func (pc *PartitionConsumer) Messages() <-chan *sarama.ConsumerMessage {
	return pc.messages
}

func (pc *PartitionConsumer) consumeMessages() {
	for msg := range pc.PartitionConsumer.Messages() {
		pc.consumeMessage(msg)
	}
	close(pc.messages)
}

func (pc *PartitionConsumer) consumeMessage(msg *sarama.ConsumerMessage) {
	opts := []ot.StartSpanOption{
		ext.SpanKindConsumer,
		ot.Tags{
			"kafka.service": msg.Topic,
			"kafka.access":  "consume",
		},
	}
	if spanContext, ok := SpanContextFromConsumerMessage(msg, pc.sensor); ok {
		opts = append(opts, ot.ChildOf(spanContext))
	}

	sp := pc.sensor.Tracer().StartSpan("kafka", opts...)
	defer sp.Finish()

	// inject consumer span context, so that it becomes a parent for subcalls
	pc.sensor.Tracer().Inject(sp.Context(), ot.TextMap, ConsumerMessageCarrier{msg})

	pc.messages <- msg
}
