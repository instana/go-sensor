package instana

import (
	"context"
	"fmt"
	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
)

func TestAppendALotDelayedSpans(t *testing.T) {
	ds := &delayedSpans{
		spans: make(chan *spanS, maxDelayedSpans),
	}

	i := 0
	for i <= 2*maxDelayedSpans {
		ds.append(&spanS{})
		i++
	}

	assert.Len(t, ds.spans, maxDelayedSpans)
}

func resetDelayedSpans() {
	delayed = &delayedSpans{
		spans: make(chan *spanS, maxDelayedSpans),
	}
}

func TestProcessDelayedSpans(t *testing.T) {
	defer resetDelayedSpans()

	recorder := NewTestRecorder()
	s := NewSensorWithTracer(NewTracerWithEverything(&Options{
		Service: "go-sensor-test",
		Tracer: TracerOptions{
			Secrets: DefaultSecretsMatcher(),
		},
	}, recorder))
	defer ShutdownSensor()

	generateSomeTraffic(s, 2*maxDelayedSpans)

	assert.Len(t, delayed.spans, maxDelayedSpans)

	stop := make(chan struct{})
	defer close(stop)

	c := delayed.process(stop)

	var spansProcessed []*spanS
	for s := range c {
		spansProcessed = append(spansProcessed, s)
	}

	assert.Len(t, spansProcessed, maxDelayedSpans)

	for _, s := range spansProcessed {
		assert.Equal(t, "q=term&secret=%3Credacted%3E", s.Tags["http.params"])
	}
}

func TestParallelProcessDelayedSpans(t *testing.T) {
	defer resetDelayedSpans()

	recorder := NewTestRecorder()
	s := NewSensorWithTracer(NewTracerWithEverything(&Options{
		Service: "go-sensor-test",
		Tracer: TracerOptions{
			Secrets: DefaultSecretsMatcher(),
		},
	}, recorder))
	defer ShutdownSensor()

	generateSomeTraffic(s, maxDelayedSpans*100)

	assert.Len(t, delayed.spans, maxDelayedSpans)

	workers := 15
	worker := 0
	wg := sync.WaitGroup{}
	wg.Add(workers)

	var ops uint64

	for worker < workers {
		go func() {
			stop := make(chan struct{})
			c := delayed.process(stop)

			for range c {
				atomic.AddUint64(&ops, 1)
			}

			wg.Done()
		}()
		worker++
	}

	wg.Wait()

	assert.Equal(t, uint64(maxDelayedSpans), ops)
}

func TestPartiallyFlushDelayedSpans(t *testing.T) {
	defer resetDelayedSpans()

	recorder := NewTestRecorder()
	s := NewSensorWithTracer(NewTracerWithEverything(&Options{
		Service: "go-sensor-test",
		Tracer: TracerOptions{
			Secrets: DefaultSecretsMatcher(),
		},
	}, recorder))
	defer ShutdownSensor()

	generateSomeTraffic(s, maxDelayedSpans)

	assert.Len(t, delayed.spans, maxDelayedSpans)

	notReadyAfter := maxDelayedSpans / 10
	sensor.agent = &eventuallyNotReadyClient{
		notReadyAfter: uint64(notReadyAfter),
	}

	delayed.flush()

	assert.Len(t, delayed.spans, maxDelayedSpans-notReadyAfter)
}

func TestFlushDelayedSpans(t *testing.T) {
	defer resetDelayedSpans()

	recorder := NewTestRecorder()
	s := NewSensorWithTracer(NewTracerWithEverything(&Options{
		Service: "go-sensor-test",
		Tracer: TracerOptions{
			Secrets: DefaultSecretsMatcher(),
		},
	}, recorder))
	defer ShutdownSensor()

	generateSomeTraffic(s, maxDelayedSpans)

	assert.Len(t, delayed.spans, maxDelayedSpans)

	sensor.agent = alwaysReadyClient{}

	delayed.flush()

	assert.Len(t, delayed.spans, 0)
}

func TestParallelFlushDelayedSpans(t *testing.T) {
	defer resetDelayedSpans()

	recorder := NewTestRecorder()
	s := NewSensorWithTracer(NewTracerWithEverything(&Options{
		Service: "go-sensor-test",
		Tracer: TracerOptions{
			Secrets: DefaultSecretsMatcher(),
		},
	}, recorder))
	defer ShutdownSensor()

	generateSomeTraffic(s, maxDelayedSpans*100)

	assert.Len(t, delayed.spans, maxDelayedSpans)

	workers := 15
	worker := 0
	wg := sync.WaitGroup{}
	wg.Add(workers)

	sensor.agent = alwaysReadyClient{}

	for worker < workers {
		go func() {
			delayed.flush()
			wg.Done()
		}()
		worker++
	}

	wg.Wait()

	assert.Equal(t, maxDelayedSpans, len(recorder.GetQueuedSpans()))
}

type eventuallyNotReadyClient struct {
	notReadyAfter uint64
	ops           uint64
}

func (e *eventuallyNotReadyClient) Ready() bool {
	n := atomic.AddUint64(&e.ops, 1)
	return n <= e.notReadyAfter
}

func (*eventuallyNotReadyClient) SendMetrics(data acceptor.Metrics) error           { return nil }
func (*eventuallyNotReadyClient) SendEvent(event *EventData) error                  { return nil }
func (*eventuallyNotReadyClient) SendSpans(spans []Span) error                      { return nil }
func (*eventuallyNotReadyClient) SendProfiles(profiles []autoprofile.Profile) error { return nil }
func (*eventuallyNotReadyClient) Flush(context.Context) error                       { return nil }

func generateSomeTraffic(s *Sensor, amount int) {
	h := TracingNamedHandlerFunc(s, "action", "/{action}", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(w, "Ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test?q=term&secret=mypassword", nil)

	rec := httptest.NewRecorder()

	for i := 0; i < amount; i++ {
		h.ServeHTTP(rec, req)
	}
}
