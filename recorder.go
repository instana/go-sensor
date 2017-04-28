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
	spans    []jsonSpan
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

// GetSpans returns a copy of the array of spans accumulated so far.
func (r *Recorder) GetSpans() []jsonSpan {
	r.RLock()
	defer r.RUnlock()
	spans := make([]jsonSpan, len(r.spans))
	copy(spans, r.spans)
	return spans
}

func (r *Recorder) init() {
	r.Reset()

	if r.testMode {
		return
	}

	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for range ticker.C {
			// Only attempt to send spans if we're announced and if the buffer is not empty
			if sensor.agent.canSend() && len(r.spans) > 0 {
				log.debug("Sending spans to agent", len(r.spans))
				r.send()
			}
		}
	}()
}

// Reset Drops all queued spans to sent to the backend
func (r *Recorder) Reset() {
	r.Lock()
	defer r.Unlock()
	r.spans = make([]jsonSpan, 0, sensor.options.MaxBufferedSpans)
}

// RecordSpan accepts spans to be recorded and sent to the backend
func (r *Recorder) RecordSpan(span *spanS) {
	// If we're not announced and not in test mode then just
	// return
	if !r.testMode && !sensor.agent.canSend() {
		return
	}

	var data = &jsonData{}
	kind := span.getSpanKind()

	data.SDK = &jsonSDKData{
		Name:   span.Operation,
		Type:   kind,
		Custom: &jsonCustomData{Tags: span.Tags, Logs: span.collectLogs()}}

	baggage := make(map[string]string)
	span.context.ForeachBaggageItem(func(k string, v string) bool {
		baggage[k] = v

		return true
	})

	if len(baggage) > 0 {
		data.SDK.Custom.Baggage = baggage
	}

	data.Service = span.getServiceName()

	var parentID *int64
	if span.ParentSpanID == 0 {
		parentID = nil
	} else {
		parentID = &span.ParentSpanID
	}

	r.Lock()
	defer r.Unlock()

	if len(r.spans) == sensor.options.MaxBufferedSpans {
		r.spans = r.spans[1:]
	}

	r.spans = append(r.spans, jsonSpan{
		TraceID:   span.context.TraceID,
		ParentID:  parentID,
		SpanID:    span.context.SpanID,
		Timestamp: uint64(span.Start.UnixNano()) / uint64(time.Millisecond),
		Duration:  uint64(span.Duration) / uint64(time.Millisecond),
		Name:      "sdk",
		From:      sensor.agent.from,
		Data:      data})

	if r.testMode || !sensor.agent.canSend() {
		return
	}

	if len(r.spans) >= sensor.options.ForceTransmissionStartingAt {
		log.debug("Forcing spans to agent", len(r.spans))

		r.send()
	}
}

func (r *Recorder) send() {
	go func() {
		_, err := sensor.agent.request(sensor.agent.makeURL(agentTracesURL), "POST", r.spans)

		r.Reset()

		if err != nil {
			sensor.agent.reset()
		}
	}()
}
