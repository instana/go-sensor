// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2017

package instana

import (
	"context"

	ot "github.com/opentracing/opentracing-go"
)

// ContextWithSpan returns a new context.Context holding a reference to an active span
func ContextWithSpan(ctx context.Context, sp ot.Span) context.Context {
	return ot.ContextWithSpan(ctx, sp)
}

// SpanFromContext retrieves previously stored active span from context. If there is no
// span, this method returns false.
func SpanFromContext(ctx context.Context) (ot.Span, bool) {
	span := ot.SpanFromContext(ctx)
	if span != nil {
		return span, true
	} else {
		return nil, false
	}
}
