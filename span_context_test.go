// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instana_test

import (
	"testing"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/w3ctrace"
	"github.com/stretchr/testify/assert"
)

func TestNewRootSpanContext(t *testing.T) {
	c := instana.NewRootSpanContext()

	assert.NotEmpty(t, c.TraceID)
	assert.Equal(t, c.SpanID, c.TraceID)
	assert.False(t, c.Sampled)
	assert.False(t, c.Suppressed)
	assert.Empty(t, c.Baggage)

	assert.Equal(t, instana.FormatLongID(c.TraceIDHi, c.TraceID), c.W3CContext.Parent().TraceID)
	assert.Equal(t, instana.FormatID(c.SpanID), c.W3CContext.Parent().ParentID)
	assert.Equal(t, w3ctrace.Flags{
		Sampled: true,
	}, c.W3CContext.Parent().Flags)
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
	}

	for name, parent := range examples {
		t.Run(name, func(t *testing.T) {
			c := instana.NewSpanContext(parent)

			assert.Equal(t, parent.TraceIDHi, c.TraceIDHi)
			assert.Equal(t, parent.TraceID, c.TraceID)
			assert.Equal(t, parent.SpanID, c.ParentID)
			assert.Equal(t, parent.Sampled, c.Sampled)
			assert.Equal(t, parent.Suppressed, c.Suppressed)
			assert.Equal(t, instana.FormatID(c.SpanID), c.W3CContext.Parent().ParentID)
			assert.Equal(t, instana.EUMCorrelationData{}, c.Correlation)
			assert.False(t, c.W3CContext.IsZero())
			assert.Equal(t, parent.Baggage, c.Baggage)

			assert.NotEqual(t, parent.SpanID, c.SpanID)
			assert.NotEmpty(t, c.SpanID)
			assert.False(t, &c.Baggage == &parent.Baggage)
		})
	}
}

func TestNewSpanContext_EmptyParent(t *testing.T) {
	examples := map[string]instana.SpanContext{
		"zero value": {},
		"suppressed": {Suppressed: true},
		"with correlation data": {
			Correlation: instana.EUMCorrelationData{
				Type: "web",
				ID:   "1",
			},
		},
	}

	for name, parent := range examples {
		t.Run(name, func(t *testing.T) {

			instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, nil)
			defer instana.ShutdownSensor()

			c := instana.NewSpanContext(parent)

			assert.NotEmpty(t, c.TraceID)
			assert.Equal(t, c.SpanID, c.TraceID)
			assert.Empty(t, c.ParentID)
			assert.False(t, c.Sampled)
			assert.Equal(t, parent.Suppressed, c.Suppressed)
			assert.Equal(t, instana.EUMCorrelationData{}, c.Correlation)
			assert.False(t, c.W3CContext.IsZero())
			assert.Empty(t, c.Baggage)
		})
	}
}

func TestNewSpanContext_FromW3CTraceContext(t *testing.T) {
	parent := instana.SpanContext{
		W3CContext: w3ctrace.Context{
			RawParent: "00-00000000000000010000000000000002-0000000000000003-01",
			RawState:  "in=1234;5678,vendor1=data",
		},
	}

	instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, nil)
	defer instana.ShutdownSensor()

	c := instana.NewSpanContext(parent)

	assert.NotEqual(t, parent.SpanID, c.SpanID)
	assert.Equal(t, instana.SpanContext{
		TraceIDHi: 0x1,
		TraceID:   0x2,
		ParentID:  0x3,
		SpanID:    c.SpanID,
		W3CContext: w3ctrace.Context{
			RawParent: "00-00000000000000010000000000000002-" + instana.FormatID(c.SpanID) + "-01",
			RawState:  "in=1234;5678,vendor1=data",
		},
		ForeignTrace: true,
		Links: []instana.SpanReference{
			{TraceID: "1234", SpanID: "5678"},
		},
	}, c)
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

func TestSpanContext_IsZero(t *testing.T) {
	examples := map[string]instana.SpanContext{
		"with 64-bit trace ID":  {TraceID: 0x1},
		"with 128-bit trace ID": {TraceIDHi: 0x1},
		"with span ID":          {SpanID: 0x1},
		"with w3c context": {
			W3CContext: w3ctrace.New(w3ctrace.Parent{
				Version:  w3ctrace.Version_Max,
				TraceID:  "abcd",
				ParentID: "1234",
			}),
		},
		"with suppressed option": {
			Suppressed: true,
		},
	}

	for name, sc := range examples {
		t.Run(name, func(t *testing.T) {
			assert.False(t, sc.IsZero())
		})
	}

	t.Run("zero value", func(t *testing.T) {
		assert.True(t, instana.SpanContext{}.IsZero())
	})
}

func TestSpanContext_Clone(t *testing.T) {
	c := instana.SpanContext{
		TraceIDHi:  10,
		TraceID:    1,
		SpanID:     2,
		ParentID:   3,
		Sampled:    true,
		Suppressed: true,
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
