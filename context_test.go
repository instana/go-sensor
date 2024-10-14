// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instana_test

import (
	"context"
	"testing"

	instana "github.com/instana/go-sensor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpanFromContext_WithActiveSpan(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	defer instana.ShutdownSensor()

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

func TestAddToContext_WithAnExistingKey(t *testing.T) {
	ctx := instana.AddToContext(context.Background(), instana.ContextKey("redis_command"), "GET")

	val := instana.GetValueFromContext(ctx, instana.ContextKey("redis_command"))
	assert.Equal(t, val, "GET")
}

func TestAddToContext_WithOutAnExistingKey(t *testing.T) {
	ctx := context.Background()

	val := instana.GetValueFromContext(ctx, instana.ContextKey("redis_command"))
	assert.Equal(t, val, "")
}
