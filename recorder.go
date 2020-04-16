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

// NewRecorder Establish a Recorder span recorder
func NewRecorder() *Recorder {
	r := new(Recorder)
	r.init()
	return r
}

// NewTestRecorder Establish a new span recorder used for testing
func NewTestRecorder() *Recorder {
	r := new(Recorder)
	r.testMode = true
	r.init()
	return r
}

func (r *Recorder) init() {
	r.clearQueuedSpans()

	if r.testMode {
		return
	}

	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for range ticker.C {
			if sensor.agent.canSend() {
				r.send()
			}
		}
	}()
}

// RecordSpan accepts spans to be recorded and and added to the span queue
// for eventual reporting to the host agent.
func (r *Recorder) RecordSpan(span *spanS) {
	// If we're not announced and not in test mode then just
	// return
	if !r.testMode && !sensor.agent.canSend() {
		return
	}

	r.Lock()
	defer r.Unlock()

	if len(r.spans) == sensor.options.MaxBufferedSpans {
		r.spans = r.spans[1:]
	}

	r.spans = append(r.spans, newSpan(span, sensor.agent.from))

	if r.testMode || !sensor.agent.canSend() {
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
	var mbs int

	if len(r.spans) > 0 {
		if sensor != nil {
			mbs = sensor.options.MaxBufferedSpans
		} else {
			mbs = DefaultMaxBufferedSpans
		}
		r.spans = make([]Span, 0, mbs)
	}
}

// Retrieve the queued spans and post them to the host agent asynchronously.
func (r *Recorder) send() {
	spansToSend := r.GetQueuedSpans()
	if len(spansToSend) > 0 {
		go func() {
			_, err := sensor.agent.request(sensor.agent.makeURL(agentTracesURL), "POST", spansToSend)
			if err != nil {
				sensor.logger.Debug("Posting traces failed in send(): ", err)
				sensor.agent.reset()
			}
		}()
	}
}
