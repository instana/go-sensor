package instasarama

import (
	"github.com/Shopify/sarama"
	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"
)

type spanKey struct {
	Topic     string
	Partition int32
	Offset    int64
}

func activeSpanKey(msg *sarama.ProducerMessage) spanKey {
	return spanKey{
		Topic:     msg.Topic,
		Partition: msg.Partition,
		Offset:    msg.Offset,
	}
}

// AsyncProducer is a wrapper for sarama.AsyncProducer that instruments its calls using
// provided instana.Sensor
type AsyncProducer struct {
	sarama.AsyncProducer
	sensor *instana.Sensor

	awaitResult bool

	input     chan *sarama.ProducerMessage
	successes chan *sarama.ProducerMessage
	errors    chan *sarama.ProducerError

	channelStates uint8 // bit fields describing the open/closed state of the response channels
	activeSpans   map[spanKey]ot.Span
}

const (
	apSuccessesChanReady = uint8(1) << iota
	apErrorsChanReady

	apAllChansReady = apSuccessesChanReady | apErrorsChanReady
)

// NewAsyncProducer creates a new sarama.AsyncProducer using the given broker addresses and configuration, and
// instruments its calls
func NewAsyncProducer(addrs []string, conf *sarama.Config, sensor *instana.Sensor) (sarama.AsyncProducer, error) {
	ap, err := sarama.NewAsyncProducer(addrs, conf)
	if err != nil {
		return ap, err
	}

	return WrapAsyncProducer(ap, conf, sensor), nil
}

// NewAsyncProducerFromClient creates a new sarama.AsyncProducer using the given client, and
// instruments its calls
func NewAsyncProducerFromClient(client sarama.Client, sensor *instana.Sensor) (sarama.AsyncProducer, error) {
	ap, err := sarama.NewAsyncProducerFromClient(client)
	if err != nil {
		return ap, err
	}

	return WrapAsyncProducer(ap, client.Config(), sensor), nil
}

// WrapAsyncProducer wraps an existing sarama.AsyncProducer and instruments its calls. It requires the same
// config that was used to create this consumer to detect whether the producer is supposed to return
// successes/errors. To initialize a new  sync producer instance use instasarama.NewAsyncProducer() and
// instasarama.NewAsyncProducerFromClient() convenience methods instead
func WrapAsyncProducer(p sarama.AsyncProducer, conf *sarama.Config, sensor *instana.Sensor) *AsyncProducer {
	ap := &AsyncProducer{
		AsyncProducer: p,
		sensor:        sensor,
		input:         make(chan *sarama.ProducerMessage),
		successes:     make(chan *sarama.ProducerMessage),
		errors:        make(chan *sarama.ProducerError),
		channelStates: apAllChansReady,
	}

	if conf != nil {
		ap.awaitResult = conf.Producer.Return.Successes && conf.Producer.Return.Errors
		ap.activeSpans = make(map[spanKey]ot.Span)
	}

	go ap.consume()

	return ap
}

// Input is the input channel for the user to write messages to that they wish to send. The async producer
// will than create a new exit span for each message that has trace context added with instasarama.ProducerMessageWithSpan()
func (p *AsyncProducer) Input() chan<- *sarama.ProducerMessage { return p.input }

// Successes is the success output channel back to the user
func (p *AsyncProducer) Successes() <-chan *sarama.ProducerMessage { return p.successes }

// Errors is the error output channel back to the user
func (p *AsyncProducer) Errors() <-chan *sarama.ProducerError { return p.errors }

func (p *AsyncProducer) consume() {
	for p.channelStates&apAllChansReady != 0 {
		select {
		case msg := <-p.input:
			sp := startProducerSpan(p.sensor, msg)
			if sp != nil {
				if p.awaitResult { // postpone span finish until the result is received
					p.activeSpans[activeSpanKey(msg)] = sp
				} else {
					sp.Finish()
				}

				p.sensor.Tracer().Inject(sp.Context(), ot.TextMap, ProducerMessageCarrier{msg})
			}

			p.AsyncProducer.Input() <- msg
		case msg, ok := <-p.AsyncProducer.Successes():
			if !ok {
				p.channelStates &= ^apSuccessesChanReady
				continue
			}
			p.successes <- msg

			key := activeSpanKey(msg)
			if sp, ok := p.activeSpans[key]; ok {
				delete(p.activeSpans, key)

				sp.Finish()
			}
		case msg, ok := <-p.AsyncProducer.Errors():
			if !ok {
				p.channelStates &= ^apErrorsChanReady
				continue
			}
			p.errors <- msg

			key := activeSpanKey(msg.Msg)
			if sp, ok := p.activeSpans[key]; ok {
				delete(p.activeSpans, key)

				sp.SetTag("kafka.error", msg.Err)
				sp.LogFields(otlog.Error(msg.Err))
				sp.Finish()
			}
		}
	}
}
