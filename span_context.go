package instana

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

	// The span's associated baggage.
	Baggage map[string]string // initialized on first use
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
		TraceID:  c.TraceID,
		SpanID:   c.SpanID,
		ParentID: c.ParentID,
		Sampled:  c.Sampled,
	}

	if c.Baggage != nil {
		res.Baggage = make(map[string]string, len(c.Baggage))
		for k, v := range c.Baggage {
			res.Baggage[k] = v
		}
	}

	return res
}
