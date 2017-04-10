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
	spans []Span
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

func NewRecorder() *SpanRecorder {
	r := new(SpanRecorder)
	r.init()

	return r
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
	s := getStringTag(rawSpan, string(ext.Component))
	if s == "" {
		s = getStringTag(rawSpan, string(ext.PeerService))
		if s == "" {
			return sensor.serviceName
		}
	}

	return s
}

func getHTTPType(rawSpan basictracer.RawSpan) string {
	kind := getStringTag(rawSpan, string(ext.SpanKind))
	if kind == string(ext.SpanKindRPCServerEnum) {
		return HTTPServer
	}

	return HTTPClient
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
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for range ticker.C {
			log.debug("Sending spans to agent", len(r.spans))

			r.send()
		}
	}()
}

func (r *SpanRecorder) reset() {
	r.Lock()
	defer r.Unlock()
	r.spans = make([]Span, 0, sensor.options.MaxBufferedSpans)
}

func (r *SpanRecorder) RecordSpan(rawSpan basictracer.RawSpan) {
	var data = &Data{}
	var tp string
	h := getHostName(rawSpan)
	status := getIntTag(rawSpan, string(ext.HTTPStatusCode))
	if status >= 0 {
		tp = getHTTPType(rawSpan)
		data = &Data{HTTP: &HTTPData{
			Host:   h,
			URL:    getStringTag(rawSpan, string(ext.HTTPUrl)),
			Method: getStringTag(rawSpan, string(ext.HTTPMethod)),
			Status: status},
			SDK: &SDKData{Name: tp}}
	} else {
		log.debug("No HTTP status code provided or invalid status code, opting out to RPC")

		tp = RPC
		data = &Data{RPC: &RPCData{
			Host: h,
			Call: rawSpan.Operation},
			SDK: &SDKData{Name: tp}}
	}

	data.Custom = &CustomData{Tags: rawSpan.Tags, Logs: collectLogs(rawSpan)}

	baggage := make(map[string]string)
	rawSpan.Context.ForeachBaggageItem(func(k string, v string) bool {
		baggage[k] = v

		return true
	})

	if len(baggage) > 0 {
		data.Baggage = baggage
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

	if len(r.spans) == sensor.options.ForceTransmissionStartingAt {
		log.debug("Forcing spans to agent", len(r.spans))

		r.send()
	}
}

func (r *SpanRecorder) send() {
	if sensor.agent.canSend() {
		go func() {
			_, err := sensor.agent.request(sensor.agent.makeURL(AgentTracesURL), "POST", r.spans)

			r.reset()

			if err != nil {
				sensor.agent.reset()
			}
		}()
	}
}
