// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2017

package instana

import (
	"context"

	ot "github.com/opentracing/opentracing-go"
)

type ContextKey string

var activeSpanKey ContextKey = "active_span"
var redisCommand ContextKey = "redis_command"

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

func AddToContext(ctx context.Context, key ContextKey, value string) context.Context {
	return context.WithValue(ctx, key, value)
}

func GetValueFromContext(ctx context.Context, key ContextKey) string {
	val, ok := ctx.Value(key).(string)
	if !ok {
		return ""
	}
	return val
}
