package instalambda

import (
	"encoding/json"

	"github.com/opentracing/opentracing-go"
)

func extractTriggerEventTags(payload []byte) opentracing.Tags {
	var v struct {
	}

	if err := json.Unmarshal(payload, &v); err != nil {
		return opentracing.Tags{}
	}

	switch {
	default:
		return opentracing.Tags{}
	}
}
