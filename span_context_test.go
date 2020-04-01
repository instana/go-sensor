package instana_test

import (
	"testing"

	instana "github.com/instana/go-sensor"
	"github.com/stretchr/testify/assert"
)

func TestSpanContext_WithBaggageItem(t *testing.T) {
	c := instana.SpanContext{
		TraceID: 1,
		SpanID:  2,
		Sampled: true,
		Baggage: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}

	updated := c.WithBaggageItem("key3", "value3")
	assert.Equal(t, instana.SpanContext{
		TraceID:  1,
		SpanID:   2,
		ParentID: 3,
		Sampled:  true,
		Baggage: map[string]string{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
		},
	}, updated)

	assert.Equal(t, instana.SpanContext{
		TraceID: 1,
		SpanID:  2,
		Sampled: true,
		Baggage: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}, c)
}

func TestSpanContext_Clone(t *testing.T) {
	c := instana.SpanContext{
		TraceID: 1,
		SpanID:  2,
		Sampled: true,
		Baggage: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}

	cloned := c.Clone()
	assert.Equal(t, c, cloned)
	assert.False(t, &c == &cloned)
	assert.False(t, &c.Baggage == &cloned.Baggage)
}

func TestSpanContext_Clone_NoBaggage(t *testing.T) {
	c := instana.SpanContext{
		TraceID: 1,
		SpanID:  2,
		Sampled: true,
	}

	cloned := c.Clone()
	assert.Equal(t, c, cloned)
}
