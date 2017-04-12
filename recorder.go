package instana

import (
	"os"
	"sync"
	"time"

	"github.com/opentracing/basictracer-go"
	ext "github.com/opentracing/opentracing-go/ext"
)

type SpanRecorder struct {
	sync.RWMutex
	spans    []Span
	testMode bool
}

type Span struct {
	TraceID   uint64      `json:"t"`
	ParentID  *uint64     `json:"p,omitempty"`
	SpanID    uint64      `json:"s"`
	Timestamp uint64      `json:"ts"`
	Duration  uint64      `json:"d"`
	Name      string      `json:"n"`
	From      *FromS      `json:"f"`
	Data      interface{} `json:"data"`
}

// NewRecorder Establish a new span recorder
func NewRecorder(testMode bool) *SpanRecorder {
	r := new(SpanRecorder)
	r.testMode = testMode
	r.init()

	return r
}

// GetSpans returns a copy of the array of spans accumulated so far.
func (r *SpanRecorder) GetSpans() []Span {
	r.RLock()
	defer r.RUnlock()
	spans := make([]Span, len(r.spans))
	copy(spans, r.spans)
	return spans
}

func getTag(rawSpan basictracer.RawSpan, tag string) interface{} {
	return rawSpan.Tags[tag]
}

func getIntTag(rawSpan basictracer.RawSpan, tag string) int {
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

func getStringTag(rawSpan basictracer.RawSpan, tag string) string {
	d := getTag(rawSpan, tag)
	if d == nil {
		return ""
	}

	return d.(string)
}

func getHostName(rawSpan basictracer.RawSpan) string {
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

func getServiceName(rawSpan basictracer.RawSpan) string {
	// ServiceName can be determined from multiple sources and has
	// the following priority (preferred first):
	//   1. Added to the span via the OT component tag
	//   2. If span has http.url, then return ""
	//   2. Specified in the tracer instantiation
	component := getStringTag(rawSpan, string(ext.Component))

	if len(component) == 0 {
		httpURL := getStringTag(rawSpan, string(ext.HTTPUrl))

		if len(httpURL) > 0 {
			return httpURL
		}
		return component
	}
	return sensor.serviceName
}

func getSpanKind(rawSpan basictracer.RawSpan) string {
	kind := getStringTag(rawSpan, string(ext.SpanKind))

	switch kind {
	case string(ext.SpanKindRPCServerEnum), "consumer", "entry":
		return "entry"
	case string(ext.SpanKindRPCClientEnum), "producer", "exit":
		return "exit"
	}
	return ""
}

func collectLogs(rawSpan basictracer.RawSpan) map[uint64]map[string]interface{} {
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

func (r *SpanRecorder) init() {
	r.reset()

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

func (r *SpanRecorder) reset() {
	r.Lock()
	defer r.Unlock()
	r.spans = make([]Span, 0, sensor.options.MaxBufferedSpans)
}

func (r *SpanRecorder) RecordSpan(rawSpan basictracer.RawSpan) {
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

	var parentID *uint64
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
		Data:      &data})

	if !r.testMode && (len(r.spans) == sensor.options.ForceTransmissionStartingAt) {
		log.debug("Forcing spans to agent", len(r.spans))

		r.send()
	}
}

func (r *SpanRecorder) send() {
	if sensor.agent.canSend() && !r.testMode {
		go func() {
			_, err := sensor.agent.request(sensor.agent.makeURL(AgentTracesURL), "POST", r.spans)

			r.reset()

			if err != nil {
				sensor.agent.reset()
			}
		}()
	}
}
