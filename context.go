package instana

// SpanContext holds the basic Span metadata.
type SpanContext struct {
	// A probabilistically unique identifier for a [multi-span] trace.
	TraceID int64

	// A probabilistically unique identifier for a span.
	SpanID int64

	// Parent Span ID.
	ParentSpanID int64

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
	var newBaggage map[string]string
	if c.Baggage == nil {
		newBaggage = map[string]string{key: val}
	} else {
		newBaggage = make(map[string]string, len(c.Baggage)+1)
		for k, v := range c.Baggage {
			newBaggage[k] = v
		}
		newBaggage[key] = val
	}

	return SpanContext{
		TraceID:      c.TraceID,
		SpanID:       c.SpanID,
		ParentSpanID: c.ParentSpanID,
		Sampled:      c.Sampled,
		Baggage:      newBaggage,
	}
}
