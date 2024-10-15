// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2017

package instana

import (
	"context"

	ot "github.com/opentracing/opentracing-go"
)

type ContextKey int8

const (
	activeSpanKey ContextKey = iota
	redisCommandKey
)

// ContextWithSpan returns a new context.Context holding a reference to an active span
func ContextWithSpan(ctx context.Context, sp ot.Span) context.Context {
	return context.WithValue(ctx, activeSpanKey, sp)
}

// SpanFromContext retrieves previously stored active span from context. If there is no
// span, this method returns false.
func SpanFromContext(ctx context.Context) (ot.Span, bool) {
	sp, ok := ctx.Value(activeSpanKey).(ot.Span)
	return sp, ok
}

func addToContext(ctx context.Context, key ContextKey, value string) context.Context {
	return context.WithValue(ctx, key, value)
}

func getValueFromContext(ctx context.Context, key ContextKey) string {
	val, _ := ctx.Value(key).(string)
	return val
}
