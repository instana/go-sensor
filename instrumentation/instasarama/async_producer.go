package instasarama

import (
	"github.com/Shopify/sarama"
	instana "github.com/instana/go-sensor"
)

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
}

const (
	apSuccessesChanReady = uint8(1) << iota
	apErrorsChanReady

	apAllChansReady = apSuccessesChanReady | apErrorsChanReady
)

// WrapAsyncProducer wraps an existing sarama.AsyncProducer and instruments its calls. It requires the same
// config that was used to create this consumer to detect whether the producer is supposed to return
// successes/errors.
func NewAsyncProducer(p sarama.AsyncProducer, conf *sarama.Config, sensor *instana.Sensor) *AsyncProducer {
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
	}

	go ap.consume()

	return ap
}

// Input is the input channel for the user to write messages to that they
// wish to send
func (p *AsyncProducer) Input() chan<- *sarama.ProducerMessage { return p.input }

// Successes is the success output channel back to the user
func (p *AsyncProducer) Successes() <-chan *sarama.ProducerMessage { return p.successes }

// Errors is the error output channel back to the user
func (p *AsyncProducer) Errors() <-chan *sarama.ProducerError { return p.errors }

func (p *AsyncProducer) consume() {
	for p.channelStates&apAllChansReady != 0 {
		select {
		case msg := <-p.input:
			p.AsyncProducer.Input() <- msg
		case msg, ok := <-p.AsyncProducer.Successes():
			if !ok {
				p.channelStates &= ^apSuccessesChanReady
				continue
			}
			p.successes <- msg
		case msg, ok := <-p.AsyncProducer.Errors():
			if !ok {
				p.channelStates &= ^apErrorsChanReady
				continue
			}
			p.errors <- msg
		}
	}
}
