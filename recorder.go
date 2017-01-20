package instana

import (
	"os"
	"time"

	"github.com/opentracing/basictracer-go"
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

func getSpanLogField(rawSpan basictracer.RawSpan, field string) interface{} {
	for _, log := range rawSpan.Logs {
		for _, f := range log.Fields {
			if f.Key() == field {
				return f.Value()
			}
		}
	}

	return nil
}

func getStringSpanLogField(rawSpan basictracer.RawSpan, field string) string {
	d := getSpanLogField(rawSpan, field)
	if d == nil {
		return ""
	}

	return d.(string)
}

func getDataLogField(rawSpan basictracer.RawSpan) *Data {
	d := getSpanLogField(rawSpan, "data")
	if d != nil {
		return getSpanLogField(rawSpan, "data").(*Data)
	}

	return nil
}

func (r *SpanRecorder) RecordSpan(rawSpan basictracer.RawSpan) {
	data := getDataLogField(rawSpan)
	tp := getStringSpanLogField(rawSpan, "type")
	if data == nil {
		h, err := os.Hostname()
		if err != nil {
			h = "localhost"
		}

		data = &Data{RPC: &RPCData{
			Host: h,
			Call: rawSpan.Operation}}
		tp = RPC
	}

	baggage := make(map[string]string)
	rawSpan.Context.ForeachBaggageItem(func(k string, v string) bool {
		baggage[k] = v

		return true
	})

	if len(baggage) > 0 {
		data.Baggage = baggage
	}

	if data.Service == "" {
		data.Service = sensor.serviceName
	}

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
