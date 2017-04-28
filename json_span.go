package instana

import (
	ot "github.com/opentracing/opentracing-go"
)

type jsonSpan struct {
	TraceID   int64  `json:"t"`
	ParentID  *int64 `json:"p,omitempty"`
	SpanID    int64  `json:"s"`
	Timestamp uint64 `json:"ts"`
	Duration  uint64 `json:"d"`
	Name      string `json:"n"`
	From      *FromS `json:"f"`
	Data      *Data  `json:"data"`
}

type Data struct {
	Service string   `json:"service,omitempty"`
	SDK     *SDKData `json:"sdk"`
}

type CustomData struct {
	Tags    ot.Tags                           `json:"tags,omitempty"`
	Logs    map[uint64]map[string]interface{} `json:"logs,omitempty"`
	Baggage map[string]string                 `json:"baggage,omitempty"`
}

type SDKData struct {
	Name      string      `json:"name"`
	Type      string      `json:"type,omitempty"`
	Arguments string      `json:"arguments,omitempty"`
	Return    string      `json:"return,omitempty"`
	Custom    *CustomData `json:"custom,omitempty"`
}
