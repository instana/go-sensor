// (c) Copyright IBM Corp. 2022

package instana

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"
	"github.com/stretchr/testify/assert"
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

func TestPartiallyFlushDelayedSpans(t *testing.T) {
	defer resetDelayedSpans()

	//We need to simulate that the agent is not ready.
	ok, cleanupFunc := setupEnv()
	if ok {
		defer cleanupFunc()
	}

	recorder := NewTestRecorder()
	c := InitCollector(&Options{
		Service: "go-sensor-test",
		Tracer: TracerOptions{
			Secrets: DefaultSecretsMatcher(),
		},
		Recorder: recorder,
	})
	defer ShutdownCollector()

	generateSomeTraffic(c, maxDelayedSpans)

	assert.Len(t, delayed.spans, maxDelayedSpans)

	notReadyAfter := maxDelayedSpans / 10
	sensor.agent = &eventuallyNotReadyClient{
		notReadyAfter: uint64(notReadyAfter * 2),
	}

	delayed.flush()

	assert.Len(t, delayed.spans, maxDelayedSpans-notReadyAfter)
}

func TestFlushDelayedSpans(t *testing.T) {
	defer resetDelayedSpans()

	//We need to simulate that the agent is not ready.
	ok, cleanupFunc := setupEnv()
	if ok {
		defer cleanupFunc()
	}

	recorder := NewTestRecorder()
	c := InitCollector(&Options{
		Service: "go-sensor-test",
		Tracer: TracerOptions{
			Secrets: DefaultSecretsMatcher(),
		},
		Recorder: recorder,
	})
	defer ShutdownCollector()

	generateSomeTraffic(c, maxDelayedSpans)

	assert.Len(t, delayed.spans, maxDelayedSpans)

	sensor.agent = alwaysReadyClient{}

	delayed.flush()

	assert.Len(t, delayed.spans, 0)
}

func TestParallelFlushDelayedSpans(t *testing.T) {
	defer resetDelayedSpans()

	//We need to simulate that the agent is not ready.
	ok, cleanupFunc := setupEnv()
	if ok {
		defer cleanupFunc()
	}

	m, _ := NamedMatcher(ContainsIgnoreCaseMatcher, []string{"q", "secret"})

	recorder := NewTestRecorder()
	c := InitCollector(&Options{
		Service: "go-sensor-test",
		Tracer: TracerOptions{
			Secrets: m,
		},
		Recorder: recorder,
	})
	defer ShutdownCollector()

	generateSomeTraffic(c, maxDelayedSpans*2)

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

	recordedSpans := recorder.GetQueuedSpans()
	assert.Equal(t, maxDelayedSpans, len(recordedSpans))

	for _, v := range recordedSpans {
		if v, ok := v.Data.(HTTPSpanData); ok {
			assert.Equal(t, "q=%3Credacted%3E&secret=%3Credacted%3E", v.Tags.Params)
		} else {
			assert.Fail(t, "wrong span type")
		}
	}
}

func generateSomeTraffic(s TracerLogger, amount int) {
	h := TracingNamedHandlerFunc(s, "action", "/{action}", func(w http.ResponseWriter, req *http.Request) {
		_, _ = fmt.Fprintln(w, "Ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test?q=term&secret=mypassword", nil)

	rec := httptest.NewRecorder()

	for i := 0; i < amount; i++ {
		h.ServeHTTP(rec, req)
	}
}

func resetDelayedSpans() {
	delayed = &delayedSpans{
		spans: make(chan *spanS, maxDelayedSpans),
	}
}

func setupEnv() (bool, func()) {
	// The presence of INSTANA_ENDPOINT_URL will lead to the creation of serverless agent client.
	if url, ok := os.LookupEnv("INSTANA_ENDPOINT_URL"); ok {
		if err := os.Unsetenv("INSTANA_ENDPOINT_URL"); err != nil {
			fmt.Println("failed to unset INSTANA_ENDPOINT_URL")
			panic(err)
		}
		return ok, func() {
			if err := os.Setenv("INSTANA_ENDPOINT_URL", url); err != nil {
				fmt.Println("failed to set INSTANA_ENDPOINT_URL")
				panic(err)
			}
		}
	}
	return false, nil
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
