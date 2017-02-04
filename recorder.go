package instana

import (
	"os"
	"time"

	"github.com/opentracing/basictracer-go"
	ext "github.com/opentracing/opentracing-go/ext"
)

type SpanRecorder struct {
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
	return new(SpanRecorder)
}

func getTag(rawSpan basictracer.RawSpan, tag string) interface{} {
	return rawSpan.Tags[tag]
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
		return HTTPClient
	}

	return HTTPServer
}

func (r *SpanRecorder) RecordSpan(rawSpan basictracer.RawSpan) {
	var data = &Data{}
	var tp string
	h := getHostName(rawSpan)
	status := getTag(rawSpan, string(ext.HTTPStatusCode))
	if status != nil {
		tp = getHTTPType(rawSpan)
		data = &Data{HTTP: &HTTPData{
			Host:   h,
			URL:    getStringTag(rawSpan, string(ext.HTTPUrl)),
			Method: getStringTag(rawSpan, string(ext.HTTPMethod)),
			Status: status.(int)}}
	} else {
		tp = RPC
		data = &Data{RPC: &RPCData{
			Host: h,
			Call: rawSpan.Operation}}
	}

	data.Custom = &CustomData{Tags: rawSpan.Tags,
		Logs: rawSpan.Logs}

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

	if sensor.agent.canSend() {
		span := &Span{
			TraceID:   rawSpan.Context.TraceID,
			ParentID:  parentID,
			SpanID:    rawSpan.Context.SpanID,
			Timestamp: uint64(rawSpan.Start.UnixNano()) / uint64(time.Millisecond),
			Duration:  uint64(rawSpan.Duration) / uint64(time.Millisecond),
			Name:      tp,
			From:      sensor.agent.from,
			Data:      &data}

		go sensor.agent.request(sensor.agent.makeURL(AgentTracesURL), "POST", []interface{}{span})
	}
}
