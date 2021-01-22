// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instana_test

import (
	"testing"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/w3ctrace"
	"github.com/instana/testify/assert"
)

func TestNewRootSpanContext(t *testing.T) {
	c := instana.NewRootSpanContext()

	assert.NotEmpty(t, c.TraceID)
	assert.Equal(t, c.SpanID, c.TraceID)
	assert.False(t, c.Sampled)
	assert.False(t, c.Suppressed)
	assert.Empty(t, c.Baggage)
	assert.Nil(t, c.ForeignParent)
}

func TestNewSpanContext(t *testing.T) {
	examples := map[string]instana.SpanContext{
		"no w3c trace": {
			TraceIDHi:  10,
			TraceID:    1,
			SpanID:     2,
			ParentID:   3,
			Sampled:    true,
			Suppressed: true,
			Baggage: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
		"with w3c trace, no instana state": {
			TraceIDHi: 10,
			TraceID:   1,
			SpanID:    2,
			W3CContext: w3ctrace.Context{
				RawParent: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
				RawState:  "vendor1=data",
			},
		},
		"with w3c trace, last state from instana": {
			TraceIDHi: 10,
			TraceID:   1,
			SpanID:    2,
			W3CContext: w3ctrace.Context{
				RawParent: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
				RawState:  "in=1234;5678,vendor1=data",
			},
		},
		"with correlation data": {
			TraceIDHi: 10,
			TraceID:   1,
			SpanID:    2,
			Correlation: instana.EUMCorrelationData{
				Type: "web",
				ID:   "1",
			},
		},
	}

	for name, parent := range examples {
		t.Run(name, func(t *testing.T) {
			c := instana.NewSpanContext(parent)

			assert.Equal(t, parent.TraceIDHi, c.TraceIDHi)
			assert.Equal(t, parent.TraceID, c.TraceID)
			assert.Equal(t, parent.SpanID, c.ParentID)
			assert.Equal(t, parent.Sampled, c.Sampled)
			assert.Equal(t, parent.Suppressed, c.Suppressed)
			assert.Equal(t, parent.W3CContext, c.W3CContext)
			assert.Equal(t, instana.EUMCorrelationData{}, c.Correlation)
			assert.Equal(t, parent.Baggage, c.Baggage)

			assert.NotEqual(t, parent.SpanID, c.SpanID)
			assert.NotEmpty(t, c.SpanID)
			assert.False(t, &c.Baggage == &parent.Baggage)
			assert.Nil(t, c.ForeignParent)
		})
	}
}

func TestNewSpanContext_EmptyParent(t *testing.T) {
	c := instana.NewSpanContext(instana.SpanContext{})

	assert.NotEmpty(t, c.TraceID)
	assert.Equal(t, c.SpanID, c.TraceID)
	assert.False(t, c.Sampled)
	assert.False(t, c.Suppressed)
	assert.Equal(t, instana.EUMCorrelationData{}, c.Correlation)
	assert.Empty(t, c.Baggage)
	assert.Nil(t, c.ForeignParent)
}

func TestNewSpanContext_ForeignParent(t *testing.T) {
	examples := map[string]struct {
		Parent           instana.SpanContext
		ExpectedTraceID  int64
		ExpectedParentID int64
	}{
		"no trace, last state from instana": {
			Parent: instana.SpanContext{
				W3CContext: w3ctrace.Context{
					RawParent: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
					RawState:  "in=1234;5678,vendor1=data",
				},
			},
			ExpectedTraceID:  0x1234,
			ExpectedParentID: 0x5678,
		},
		"with trace, last state not from instana": {
			Parent: instana.SpanContext{
				TraceID: 0x4321,
				SpanID:  0x8765,
				W3CContext: w3ctrace.Context{
					RawParent: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
					RawState:  "vendor1=data,in=1234;5678",
				},
			},
			ExpectedTraceID:  0x4321,
			ExpectedParentID: 0x8765,
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			c := instana.NewSpanContext(example.Parent)
			assert.NotEqual(t, example.Parent.SpanID, c.SpanID)
			assert.Equal(t, instana.SpanContext{
				TraceID:       example.ExpectedTraceID,
				SpanID:        c.SpanID,
				ParentID:      example.ExpectedParentID,
				W3CContext:    example.Parent.W3CContext,
				ForeignParent: example.Parent.W3CContext,
			}, c)
		})
	}
}

func TestSpanContext_WithBaggageItem(t *testing.T) {
	c := instana.SpanContext{
		TraceIDHi: 10,
		TraceID:   1,
		SpanID:    2,
		ParentID:  3,
		Sampled:   true,
		Baggage: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}

	updated := c.WithBaggageItem("key3", "value3")
	assert.Equal(t, instana.SpanContext{
		TraceIDHi: 10,
		TraceID:   1,
		SpanID:    2,
		ParentID:  3,
		Sampled:   true,
		Baggage: map[string]string{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
		},
	}, updated)

	assert.Equal(t, instana.SpanContext{
		TraceIDHi: 10,
		TraceID:   1,
		SpanID:    2,
		ParentID:  3,
		Sampled:   true,
		Baggage: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}, c)
}

func TestSpanContext_Clone(t *testing.T) {
	c := instana.SpanContext{
		TraceIDHi:     10,
		TraceID:       1,
		SpanID:        2,
		ParentID:      3,
		Sampled:       true,
		Suppressed:    true,
		ForeignParent: []byte("parent"),
		W3CContext: w3ctrace.New(w3ctrace.Parent{
			TraceID:  "w3ctraceid",
			ParentID: "w3cparentid",
		}),
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
		TraceIDHi: 10,
		TraceID:   1,
		SpanID:    2,
		ParentID:  3,
		Sampled:   true,
	}

	cloned := c.Clone()
	assert.Equal(t, c, cloned)
}
