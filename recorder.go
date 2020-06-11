package instana

import (
	"sync"
	"time"
)

// A SpanRecorder handles all of the `RawSpan` data generated via an
// associated `Tracer` (see `NewStandardTracer`) instance. It also names
// the containing process and provides access to a straightforward tag map.
type SpanRecorder interface {
	// Implementations must determine whether and where to store `span`.
	RecordSpan(span *spanS)
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

	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for range ticker.C {
			if sensor.agent.Ready() {
				r.send()
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

// RecordSpan accepts spans to be recorded and and added to the span queue
// for eventual reporting to the host agent.
func (r *Recorder) RecordSpan(span *spanS) {
	// If we're not announced and not in test mode then just
	// return
	if !r.testMode && !sensor.agent.Ready() {
		return
	}

	r.Lock()
	defer r.Unlock()

	if len(r.spans) == sensor.options.MaxBufferedSpans {
		r.spans = r.spans[1:]
	}

	r.spans = append(r.spans, newSpan(span, sensor.agent.from))

	if r.testMode || !sensor.agent.Ready() {
		return
	}

	if len(r.spans) >= sensor.options.ForceTransmissionStartingAt {
		sensor.logger.Debug("Forcing spans to agent. Count:", len(r.spans))
		go r.send()
	}
}

// QueuedSpansCount returns the number of queued spans
//   Used only in tests currently.
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

// clearQueuedSpans brings the span queue to empty/0/nada
//   This function doesn't take the Lock so make sure to have
//   the write lock before calling.
//   This is meant to be called from GetQueuedSpans which handles
//   locking.
func (r *Recorder) clearQueuedSpans() {
	r.spans = r.spans[:0]
}

// Retrieve the queued spans and post them to the host agent asynchronously.
func (r *Recorder) send() {
	spansToSend := r.GetQueuedSpans()
	if len(spansToSend) == 0 {
		return
	}

	go sensor.agent.SendSpans(spansToSend)
}
