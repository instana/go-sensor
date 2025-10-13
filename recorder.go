// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2016

package instana

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// A SpanRecorder handles all of the `RawSpan` data generated via an
// associated `Tracer` (see `NewStandardTracer`) instance. It also names
// the containing process and provides access to a straightforward tag map.
type SpanRecorder interface {
	// Implementations must determine whether and where to store `span`.
	RecordSpan(span *spanS)
	// Flush forces sending any buffered finished spans
	Flush(context.Context) error
}

// Recorder accepts spans, processes and queues them
// for delivery to the backend.
type Recorder struct {
	sync.RWMutex
	spans    []Span
	testMode bool
}

// NewRecorder initializes a new span recorder
func NewRecorder() *Recorder {
	r := &Recorder{}

	// Create a reference to r that will be captured by the goroutine
	recorder := r

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop() // Ensure ticker is stopped when goroutine exits

		for range ticker.C {

			if isAgentReady() {
				go func(*Recorder) {
					if err := r.Flush(context.Background()); err != nil {
						sensor.logger.Error("failed to flush the spans:  ", err.Error())
					}
				}(r)
			}
		}
	}()

	return r
}

// NewTestRecorder initializes a new span recorder that keeps all collected
// until they are requested. This recorder does not send spans to the agent (used for testing)
func NewTestRecorder() *Recorder {
	return &Recorder{
		testMode: true,
	}
}

// RecordSpan accepts spans to be recorded and added to the span queue
// for eventual reporting to the host agent.
func (r *Recorder) RecordSpan(span *spanS) {
	// Get all sensor-related values under a single lock to minimize contention
	muSensor.Lock()
	if sensor == nil {
		muSensor.Unlock()
		return
	}

	agentReady := sensor.Agent().Ready()
	maxBufferedSpans := sensor.options.MaxBufferedSpans
	forceTransmissionAt := sensor.options.ForceTransmissionStartingAt
	logger := sensor.logger
	muSensor.Unlock()

	// If we're not announced and not in test mode then just
	// return
	if !r.testMode && !agentReady {
		return
	}

	r.Lock()
	defer r.Unlock()

	if len(r.spans) == maxBufferedSpans {
		r.spans = r.spans[1:]
	}

	r.spans = append(r.spans, newSpan(span))

	if r.testMode || !agentReady {
		return
	}

	if len(r.spans) >= forceTransmissionAt {
		logger.Debug("forcing ", len(r.spans), "span(s) to the agent")
		// Create a reference to r for this goroutine to avoid race conditions
		rec := r
		go func(recorder *Recorder) {
			if err := recorder.Flush(context.Background()); err != nil {
				muSensor.Lock()
				if sensor != nil {
					sensor.logger.Error("failed to flush the spans: ", err.Error())
				}
				muSensor.Unlock()
			}
		}(rec)
	}
}

// QueuedSpansCount returns the number of queued spans
//
//	Used only in tests currently.
func (r *Recorder) QueuedSpansCount() int {
	r.RLock()
	defer r.RUnlock()
	return len(r.spans)
}

// GetQueuedSpans returns a copy of the queued spans and clears the queue.
func (r *Recorder) GetQueuedSpans() []Span {
	r.Lock()
	defer r.Unlock()

	// Copy queued spans
	queuedSpans := make([]Span, len(r.spans))
	copy(queuedSpans, r.spans)

	// and clear out the source
	r.clearQueuedSpans()
	return queuedSpans
}

// Flush sends queued spans to the agent
func (r *Recorder) Flush(ctx context.Context) error {
	spansToSend := r.GetQueuedSpans()
	if len(spansToSend) == 0 {
		return nil
	}

	muSensor.Lock()
	if sensor == nil {
		muSensor.Unlock()
		return nil
	}
	agent := sensor.Agent()
	muSensor.Unlock()

	if err := agent.SendSpans(spansToSend); err != nil {
		r.Lock()
		defer r.Unlock()

		// put failed spans in front of the queue to make sure they are evicted first
		// whenever the queue length exceeds options.MaxBufferedSpans
		r.spans = append(spansToSend, r.spans...)

		return fmt.Errorf("failed to send collected spans to the agent: %s", err)
	}

	return nil
}

// clearQueuedSpans brings the span queue to empty/0/nada
//
//	This function doesn't take the Lock so make sure to have
//	the write lock before calling.
//	This is meant to be called from GetQueuedSpans which handles
//	locking.
func (r *Recorder) clearQueuedSpans() {
	r.spans = r.spans[:0]
}
