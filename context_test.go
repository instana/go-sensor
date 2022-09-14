// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instana_test

import (
	"context"
	"github.com/opentracing/opentracing-go"
	"testing"

	instana "github.com/instana/go-sensor"
	"github.com/instana/testify/assert"
	"github.com/instana/testify/require"
)

func TestSpanFromContext_WithActiveSpan(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

	span := tracer.StartSpan("test")
	ctx := instana.ContextWithSpan(context.Background(), span)

	sp, ok := instana.SpanFromContext(ctx)
	require.True(t, ok)
	assert.Equal(t, span, sp)
}

func TestSpanFromContext_NoActiveSpan(t *testing.T) {
	_, ok := instana.SpanFromContext(context.Background())
	assert.False(t, ok)
}

func TestInteroperabilityWithOpenTracing(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

	// If a span was stored in context via instana (for example by http middleware)
	span := tracer.StartSpan("test")
	ctx := instana.ContextWithSpan(context.Background(), span)

	// It should be possible to retrieve it with opentracing API
	otSpan := opentracing.SpanFromContext(ctx)

	assert.NotNil(t, otSpan)
}
