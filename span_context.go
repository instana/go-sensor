// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instana

import (
	"strings"

	"github.com/instana/go-sensor/w3ctrace"
)

// EUMCorrelationData represents the data sent by the Instana End-User Monitoring script
// integrated into frontend
type EUMCorrelationData struct {
	Type string
	ID   string
}

// SpanContext holds the basic Span metadata.
type SpanContext struct {
	// The higher 4 bytes of a 128-bit trace ID
	TraceIDHi int64
	// A probabilistically unique identifier for a [multi-span] trace.
	TraceID int64
	// A probabilistically unique identifier for a span.
	SpanID int64
	// An optional parent span ID, 0 if this is the root span context.
	ParentID int64
	// Whether the trace is sampled.
	Sampled bool
	// Whether the trace is suppressed and should not be sent to the agent.
	Suppressed bool
	// The span's associated baggage.
	Baggage map[string]string // initialized on first use
	// The W3C trace context
	W3CContext w3ctrace.Context
	// Correlation is the correlation data sent by the frontend EUM script
	Correlation EUMCorrelationData
	// The 3rd party parent if the context is derived from non-Instana trace
	ForeignParent interface{}
}

// NewRootSpanContext initializes a new root span context issuing a new trace ID
func NewRootSpanContext() SpanContext {
	spanID := randomID()

	return SpanContext{
		TraceIDHi: randomID(),
		TraceID:   spanID,
		SpanID:    spanID,
	}
}

// NewSpanContext initializes a new child span context from its parent. It will
// ignore the parent context if it contains neither Instana trace and span IDs
// nor a W3C trace context
func NewSpanContext(parent SpanContext) SpanContext {
	foreign := parent.restoreFromForeignTraceContext(parent.W3CContext)

	if parent.TraceIDHi == 0 && parent.TraceID == 0 && parent.SpanID == 0 {
		return NewRootSpanContext()
	}

	c := parent.Clone()
	c.SpanID, c.ParentID = randomID(), parent.SpanID

	if foreign {
		c.ForeignParent = c.W3CContext
	}

	return c
}

func (c *SpanContext) restoreFromForeignTraceContext(trCtx w3ctrace.Context) bool {
	if trCtx.IsZero() {
		return false
	}

	st := c.W3CContext.State()

	if (c.TraceIDHi != 0 || c.TraceID != 0) && c.SpanID != 0 {
		// we've got Instana trace parent, but still need to check if the last
		// service upstream is instrumented with Instana, i.e. is the last one
		// to update the `tracestate`
		return st.Index(w3ctrace.VendorInstana) > 0
	}

	// we've got only have the 3rd-party context, which means that upstream is
	// tracing, but not with Instana, so need to either start a new trace or
	// try to pickup the existing one from `tracestate`
	c.TraceID = randomID()

	vd, ok := st.Fetch(w3ctrace.VendorInstana)
	if !ok {
		return true
	}

	i := strings.Index(vd, ";")
	if i < 0 {
		return true
	}

	traceIDHi, traceIDLo, err := ParseLongID(vd[:i])
	if err != nil {
		return true
	}

	spanID, err := ParseID(vd[i+1:])
	if err != nil {
		return true
	}

	c.TraceIDHi, c.TraceID, c.SpanID = traceIDHi, traceIDLo, spanID

	return true
}

// ForeachBaggageItem belongs to the opentracing.SpanContext interface
func (c SpanContext) ForeachBaggageItem(handler func(k, v string) bool) {
	for k, v := range c.Baggage {
		if !handler(k, v) {
			break
		}
	}
}

// WithBaggageItem returns an entirely new SpanContext with the
// given key:value baggage pair set.
func (c SpanContext) WithBaggageItem(key, val string) SpanContext {
	res := c.Clone()

	if res.Baggage == nil {
		res.Baggage = make(map[string]string, 1)
	}
	res.Baggage[key] = val

	return res
}

// Clone returns a deep copy of a SpanContext
func (c SpanContext) Clone() SpanContext {
	res := SpanContext{
		TraceIDHi:     c.TraceIDHi,
		TraceID:       c.TraceID,
		SpanID:        c.SpanID,
		ParentID:      c.ParentID,
		Sampled:       c.Sampled,
		Suppressed:    c.Suppressed,
		ForeignParent: c.ForeignParent,
		W3CContext:    c.W3CContext,
	}

	if c.Baggage != nil {
		res.Baggage = make(map[string]string, len(c.Baggage))
		for k, v := range c.Baggage {
			res.Baggage[k] = v
		}
	}

	return res
}
