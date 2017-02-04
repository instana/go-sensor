package instana

import (
	ot "github.com/opentracing/opentracing-go"
)

type CustomData struct {
	Tags ot.Tags        `json:"tags,omitempty"`
	Logs []ot.LogRecord `json:"logs,omitempty"`
}
