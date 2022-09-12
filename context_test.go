// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instana_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	instana "github.com/instana/go-sensor"
	"github.com/instana/testify/assert"
	"github.com/instana/testify/require"
	ot "github.com/opentracing/opentracing-go"
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

func TestSpanFromContext_RetrievalWithOpentracing(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)
	sensor := instana.NewSensorWithTracer(tracer)

	handler := http.HandlerFunc(instana.TracingNamedHandlerFunc(sensor, "", "/", func(w http.ResponseWriter, r *http.Request) {
		sp := ot.SpanFromContext(r.Context())

		if sp == nil {
			t.Error("expected Instana span to be retrieved by opentracing")
		}
	}))

	server := httptest.NewServer(handler)
	defer server.Close()

	_, err := http.Get(server.URL)

	if err != nil {
		t.Fatalf("http request error: %s", err)
	}
}
