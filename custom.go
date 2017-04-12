package instana

import (
	ot "github.com/opentracing/opentracing-go"
)

type CustomData struct {
	Tags    ot.Tags                           `json:"tags,omitempty"`
	Logs    map[uint64]map[string]interface{} `json:"logs,omitempty"`
	Baggage map[string]string                 `json:"baggage,omitempty"`
}
