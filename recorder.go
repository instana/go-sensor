package instana

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	ext "github.com/opentracing/opentracing-go/ext"
)

// A SpanRecorder handles all of the `RawSpan` data generated via an
// associated `Tracer` (see `NewStandardTracer`) instance. It also names
// the containing process and provides access to a straightforward tag map.
type SpanRecorder interface {
	// Implementations must determine whether and where to store `span`.
	RecordSpan(span RawSpan)
}

type InstanaRecorder struct {
	sync.RWMutex
	spans    []Span
	testMode bool
}

type Span struct {
	TraceID   int64       `json:"t"`
	ParentID  *int64      `json:"p,omitempty"`
	SpanID    int64       `json:"s"`
	Timestamp uint64      `json:"ts"`
	Duration  uint64      `json:"d"`
	Name      string      `json:"n"`
	From      *FromS      `json:"f"`
	Data      interface{} `json:"data"`
	Raw       RawSpan     `json:"-"`
}

// NewRecorder Establish a InstanaRecorder span recorder
func NewRecorder() *InstanaRecorder {
	r := new(InstanaRecorder)
	r.init()
	return r
}

// NewTestRecorder Establish a new span recorder used for testing
func NewTestRecorder() *InstanaRecorder {
	r := new(InstanaRecorder)
	r.testMode = true
	r.init()
	return r
}

// GetSpans returns a copy of the array of spans accumulated so far.
func (r *InstanaRecorder) GetSpans() []Span {
	r.RLock()
	defer r.RUnlock()
	spans := make([]Span, len(r.spans))
	copy(spans, r.spans)
	return spans
}

func (r *InstanaRecorder) init() {
	r.Reset()

	if r.testMode {
		log.debug("Recorder in test mode.  Not reporting spans to the backend.")
	} else {
		ticker := time.NewTicker(1 * time.Second)
		go func() {
			for range ticker.C {
				log.debug("Sending spans to agent", len(r.spans))

				r.send()
			}
		}()
	}
}

func (r *InstanaRecorder) Reset() {
	r.Lock()
	defer r.Unlock()
	r.spans = make([]Span, 0, sensor.options.MaxBufferedSpans)
}

func (r *InstanaRecorder) RecordSpan(rawSpan RawSpan) {
	var data = &Data{}
	kind := getSpanKind(rawSpan)

	data.SDK = &SDKData{
		Name:   rawSpan.Operation,
		Type:   kind,
		Custom: &CustomData{Tags: rawSpan.Tags, Logs: collectLogs(rawSpan)}}

	baggage := make(map[string]string)
	rawSpan.Context.ForeachBaggageItem(func(k string, v string) bool {
		baggage[k] = v

		return true
	})

	if len(baggage) > 0 {
		data.SDK.Custom.Baggage = baggage
	}

	data.Service = getServiceName(rawSpan)

	var parentID *int64
	if rawSpan.ParentSpanID == 0 {
		parentID = nil
	} else {
		parentID = &rawSpan.ParentSpanID
	}

	r.Lock()
	defer r.Unlock()

	if len(r.spans) == sensor.options.MaxBufferedSpans {
		r.spans = r.spans[1:]
	}

	r.spans = append(r.spans, Span{
		TraceID:   rawSpan.Context.TraceID,
		ParentID:  parentID,
		SpanID:    rawSpan.Context.SpanID,
		Timestamp: uint64(rawSpan.Start.UnixNano()) / uint64(time.Millisecond),
		Duration:  uint64(rawSpan.Duration) / uint64(time.Millisecond),
		Name:      "sdk",
		From:      sensor.agent.from,
		Data:      &data,
		Raw:       rawSpan})

	if !r.testMode && (len(r.spans) == sensor.options.ForceTransmissionStartingAt) {
		log.debug("Forcing spans to agent", len(r.spans))

		r.send()
	}
}

func (r *InstanaRecorder) send() {
	if sensor.agent.canSend() && !r.testMode {
		go func() {
			j, _ := json.MarshalIndent(r.spans, "", "  ")
			log.debug("spans:", bytes.NewBuffer(j))

			_, err := sensor.agent.request(sensor.agent.makeURL(AgentTracesURL), "POST", r.spans)

			r.Reset()

			if err != nil {
				sensor.agent.reset()
			}
		}()
	}
}
func getTag(rawSpan RawSpan, tag string) interface{} {
	var x, ok = rawSpan.Tags[tag]
	if !ok {
		x = ""
	}
	return x
}

func getIntTag(rawSpan RawSpan, tag string) int {
	d := rawSpan.Tags[tag]
	if d == nil {
		return -1
	}

	r, ok := d.(int)
	if !ok {
		return -1
	}

	return r
}

func getStringTag(rawSpan RawSpan, tag string) string {
	d := rawSpan.Tags[tag]
	if d == nil {
		return ""
	}
	return fmt.Sprint(d)
}

func getHostName(rawSpan RawSpan) string {
	hostTag := getStringTag(rawSpan, string(ext.PeerHostname))
	if hostTag != "" {
		return hostTag
	}

	h, err := os.Hostname()
	if err != nil {
		h = "localhost"
	}

	return h
}

func getServiceName(rawSpan RawSpan) string {
	// ServiceName can be determined from multiple sources and has
	// the following priority (preferred first):
	//   1. If added to the span via the OT component tag
	//   2. If added to the span via the OT http.url tag
	//   3. Specified in the tracer instantiation via Service option
	component := getStringTag(rawSpan, string(ext.Component))
	if len(component) > 0 {
		return component
	}

	httpURL := getStringTag(rawSpan, string(ext.HTTPUrl))
	if len(httpURL) > 0 {
		return httpURL
	}
	return sensor.serviceName
}

func getSpanKind(rawSpan RawSpan) string {
	kind := getStringTag(rawSpan, string(ext.SpanKind))

	switch kind {
	case string(ext.SpanKindRPCServerEnum), "consumer", "entry":
		return "entry"
	case string(ext.SpanKindRPCClientEnum), "producer", "exit":
		return "exit"
	}
	return ""
}

func collectLogs(rawSpan RawSpan) map[uint64]map[string]interface{} {
	logs := make(map[uint64]map[string]interface{})
	for _, l := range rawSpan.Logs {
		if _, ok := logs[uint64(l.Timestamp.UnixNano())/uint64(time.Millisecond)]; !ok {
			logs[uint64(l.Timestamp.UnixNano())/uint64(time.Millisecond)] = make(map[string]interface{})
		}

		for _, f := range l.Fields {
			logs[uint64(l.Timestamp.UnixNano())/uint64(time.Millisecond)][f.Key()] = f.Value()
		}
	}

	return logs
}
