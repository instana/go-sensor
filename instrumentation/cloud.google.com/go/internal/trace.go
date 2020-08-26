package internal

import (
	"context"

	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"
)

func StartExitSpan(ctx context.Context, op string, opts ...ot.StartSpanOption) context.Context {
	sp, ok := instana.SpanFromContext(ctx)
	if !ok {
		return ctx
	}

	opts = append(opts, ot.ChildOf(sp.Context()))

	return instana.ContextWithSpan(ctx, sp.Tracer().StartSpan(op, opts...))
}

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
