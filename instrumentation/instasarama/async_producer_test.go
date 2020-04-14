package instasarama_test

import (
	"errors"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instasarama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAsyncProducer_Input(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	msg := &sarama.ProducerMessage{
		Topic: "topic-1",
	}

	ap := newTestAsyncProducer(nil)
	defer ap.Teardown()

	wrapped := instasarama.NewAsyncProducer(ap, sarama.NewConfig(), sensor)
	wrapped.Input() <- msg

	select {
	case published := <-ap.input:
		assert.Equal(t, msg, published)
	case <-time.After(1 * time.Second):
		t.Fatalf("publishing via async producer timed out after 1s")
	}
}

func TestAsyncProducer_Successes(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	msg := &sarama.ProducerMessage{
		Topic: "topic-1",
	}

	ap := newTestAsyncProducer(nil)
	defer ap.Teardown()

	ap.successes <- msg

	wrapped := instasarama.NewAsyncProducer(ap, sarama.NewConfig(), sensor)

	select {
	case received := <-wrapped.Successes():
		assert.Equal(t, msg, received)
	case <-time.After(1 * time.Second):
		t.Fatalf("reading a success message from async producer timed out after 1s")
	}
}

func TestAsyncProducer_Errors(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	msg := &sarama.ProducerError{
		Err: errors.New("something went wrong"),
		Msg: &sarama.ProducerMessage{Topic: "topic-1"},
	}

	ap := newTestAsyncProducer(nil)
	defer ap.Teardown()

	ap.errors <- msg

	wrapped := instasarama.NewAsyncProducer(ap, sarama.NewConfig(), sensor)

	select {
	case received := <-wrapped.Errors():
		assert.Equal(t, msg, received)
	case <-time.After(1 * time.Second):
		t.Fatalf("reading an error message from async producer timed out after 1s")
	}
}

type testAsyncProducer struct {
	Error  error
	Closed bool
	Async  bool

	input     chan *sarama.ProducerMessage
	successes chan *sarama.ProducerMessage
	errors    chan *sarama.ProducerError
}

func newTestAsyncProducer(returnedErr error) *testAsyncProducer {
	return &testAsyncProducer{
		Error:     returnedErr,
		input:     make(chan *sarama.ProducerMessage, 1),
		successes: make(chan *sarama.ProducerMessage, 1),
		errors:    make(chan *sarama.ProducerError, 1),
	}
}

func (p *testAsyncProducer) AsyncClose() {
	p.Closed = true
	p.Async = true
}

func (p *testAsyncProducer) Close() error {
	p.Closed = true
	return p.Error
}

func (p *testAsyncProducer) Input() chan<- *sarama.ProducerMessage     { return p.input }
func (p *testAsyncProducer) Successes() <-chan *sarama.ProducerMessage { return p.successes }
func (p *testAsyncProducer) Errors() <-chan *sarama.ProducerError      { return p.errors }

func (p *testAsyncProducer) Teardown() {
	close(p.input)
	close(p.successes)
	close(p.errors)
}
