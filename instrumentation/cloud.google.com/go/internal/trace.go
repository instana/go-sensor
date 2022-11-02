// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package internal

import (
	"context"

	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"
)

// StartExitSpan starts a new span and injects in into returned context. If provided context already
// contains an active span, it will be used as a parent.
func StartExitSpan(ctx context.Context, op string, opts ...ot.StartSpanOption) context.Context {
	sp, ok := instana.SpanFromContext(ctx)
	if !ok {
		return ctx
	}

	opts = append(opts, ot.ChildOf(sp.Context()))

	return instana.ContextWithSpan(ctx, sp.Tracer().StartSpan(op, opts...))
}

// FinishSpan finishes an active span found in context, optionally logging an error if it's not nil.
// This function is a noop if provided context does not contain an active span.
func FinishSpan(ctx context.Context, err error) {
	sp, ok := instana.SpanFromContext(ctx)
	if !ok {
		return
	}

	if err != nil {
		sp.LogFields(otlog.Error(err))
	}

	sp.Finish()
}
