package instana

import "github.com/instana/go-sensor/w3ctrace"

// SpanContext holds the basic Span metadata.
type SpanContext struct {
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
	// A flag that signals that this context comes from a service that is
	// not monitored by Instana
	Foreign bool
	// The 3rd party parent if the context is derived from non-Instana trace
	ForeignParent interface{}
}

// NewRootSpanContext initializes a new root span context issuing a new trace ID
func NewRootSpanContext() SpanContext {
	spanID := randomID()

	return SpanContext{
		TraceID: spanID,
		SpanID:  spanID,
	}
}

// NewSpanContext initializes a new child span context from its parent
func NewSpanContext(parent SpanContext) SpanContext {
	c := parent.Clone()
	c.SpanID, c.ParentID = randomID(), parent.SpanID
	c.Foreign = false

	return c
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
		TraceID:    c.TraceID,
		SpanID:     c.SpanID,
		ParentID:   c.ParentID,
		Sampled:    c.Sampled,
		Suppressed: c.Suppressed,
		Foreign:    c.Foreign,
		W3CContext: c.W3CContext,
	}

	if c.Baggage != nil {
		res.Baggage = make(map[string]string, len(c.Baggage))
		for k, v := range c.Baggage {
			res.Baggage[k] = v
		}
	}

	return res
}
