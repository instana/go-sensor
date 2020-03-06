package instana

import (
	"context"

	ot "github.com/opentracing/opentracing-go"
)

type contextKey struct{}

var activeSpanKey contextKey

// ContextWithSpan returns a new context.Context holding a reference to an active span
func ContextWithSpan(ctx context.Context, sp ot.Span) context.Context {
	return context.WithValue(ctx, activeSpanKey, sp)
}

// SpanFromContext retrieves previously stored active span from context. If there is no
// span, this method returns false.
func SpanFromContext(ctx context.Context) (ot.Span, bool) {
	sp, ok := ctx.Value(activeSpanKey).(ot.Span)
	if !ok {
		return nil, false
	}

	return sp, true
}
