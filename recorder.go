package instana

import (
	"os"
	"time"

	"github.com/opentracing/basictracer-go"
)

type InstanaSpanRecorder struct {
	sensor *sensorS
}

type InstanaSpan struct {
	TraceId   uint64      `json:"t"`
	ParentId  *uint64     `json:"p,omitempty"`
	SpanId    uint64      `json:"s"`
	Timestamp uint64      `json:"ts"`
	Duration  uint64      `json:"d"`
	Name      string      `json:"n"`
	From      *FromS      `json:"f"`
	Data      interface{} `json:"data"`
}

func NewRecorder() *InstanaSpanRecorder {
	ret := new(InstanaSpanRecorder)
	ret.sensor = sensor

	return ret
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

func (r *InstanaSpanRecorder) RecordSpan(rawSpan basictracer.RawSpan) {
	data := getDataLogField(rawSpan)
	tp := getStringSpanLogField(rawSpan, "type")
	if data == nil {
		h, err := os.Hostname()
		if err != nil {
			h = "localhost"
		}

		data = &Data{Rpc: &RpcData{
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
		data.Service = r.sensor.serviceName
	}

	var parentId *uint64
	if rawSpan.ParentSpanID == 0 {
		parentId = nil
	} else {
		parentId = &rawSpan.ParentSpanID
	}

	if r.sensor.agent.canSend() {
		span := &InstanaSpan{
			TraceId:   rawSpan.Context.TraceID,
			ParentId:  parentId,
			SpanId:    rawSpan.Context.SpanID,
			Timestamp: uint64(rawSpan.Start.UnixNano()) / uint64(time.Millisecond),
			Duration:  uint64(rawSpan.Duration) / uint64(time.Millisecond),
			Name:      tp,
			From:      r.sensor.agent.from,
			Data:      &data}

		go r.sensor.agent.request(r.sensor.agent.makeUrl(AGENT_TRACES_URL), "POST", []interface{}{span})
	}
}
