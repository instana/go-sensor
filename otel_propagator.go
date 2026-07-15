package instana

import (
	"context"

	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// POC propagator using Instana headers
type instanaPropagator struct{}

//add trace information TO OUTGOING headers

func (p instanaPropagator) Inject(
	ctx context.Context,
	carrier propagation.TextMapCarrier,
) {

	//get the current span from the context
	span := trace.SpanFromContext(ctx)
	sc := span.SpanContext()

	//if the span context is invalid then there's nothing to propagate
	if !sc.IsValid() {
		return
	}

	//reusing existing instana headers so tracing remains
	carrier.Set("x-instana-t", sc.TraceID().String())
	carrier.Set("x-instana-s", sc.SpanID().String())
	carrier.Set("x-instana-l", "1") //mark the trace as enabled
}

//read trace information FROM INCOMING headers

func (p instanaPropagator) Extract(
	ctx context.Context,
	carrier propagation.TextMapCarrier,
) context.Context {

	//read the trace ID
	traceID, err := trace.TraceIDFromHex(
		carrier.Get("x-instana-t"),
	)
	if err != nil {
		return ctx
	}

	//read the span ID
	spanID, err := trace.SpanIDFromHex(
		carrier.Get("x-instana-s"),
	)
	if err != nil {
		return ctx
	}

	//create span context from the values found in the headers
	sc := trace.NewSpanContext(
		trace.SpanContextConfig{
			TraceID:    traceID,
			SpanID:     spanID,
			TraceFlags: trace.FlagsSampled,
			Remote:     true,
		},
	)

	//add the span context to the request context
	return trace.ContextWithRemoteSpanContext(ctx, sc)
}

//Headers used for propagation

func (p instanaPropagator) Fields() []string {
	return []string{
		"x-instana-t",
		"x-instana-s",
		"x-instana-l",
	}
}
