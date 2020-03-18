package instana_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTracingHandlerFunc_Write(t *testing.T) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{
		Service: "go-sensor-test",
	}, recorder))

	h := instana.TracingHandlerFunc(s, "test-handler", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(w, "Ok")
	})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/test?q=classified", nil))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "Ok\n", rec.Body.String())

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	assert.False(t, span.Error)
	assert.Equal(t, 0, span.Ec)

	require.NotNil(t, span.Data)
	require.NotNil(t, span.Data.SDK)
	assert.Equal(t, "test-handler", span.Data.SDK.Name)
	assert.Equal(t, "entry", span.Data.SDK.Type)

	require.NotNil(t, span.Data.SDK.Custom)
	assert.Equal(t, ot.Tags{
		"http.status_code": http.StatusOK,
		"http.method":      "GET",
		"http.url":         "/test",
		"peer.hostname":    "example.com",
		"span.kind":        ext.SpanKindRPCServerEnum,
	}, span.Data.SDK.Custom.Tags)

	// check whether the trace context has been sent back to the client
	traceID, err := instana.Header2ID(rec.Header().Get(instana.FieldT))
	require.NoError(t, err)
	assert.Equal(t, span.TraceID, traceID)

	spanID, err := instana.Header2ID(rec.Header().Get(instana.FieldS))
	require.NoError(t, err)
	assert.Equal(t, span.SpanID, spanID)
}

func TestTracingHandlerFunc_WriteHeaders(t *testing.T) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	h := instana.TracingHandlerFunc(s, "test-handler", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/test?q=classified", nil))

	assert.Equal(t, http.StatusNotImplemented, rec.Code)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	assert.False(t, span.Error)
	assert.Equal(t, 0, span.Ec)

	require.NotNil(t, span.Data)
	require.NotNil(t, span.Data.SDK)
	assert.Equal(t, "test-handler", span.Data.SDK.Name)
	assert.Equal(t, "entry", span.Data.SDK.Type)

	require.NotNil(t, span.Data.SDK.Custom)
	assert.Equal(t, ot.Tags{
		"http.method":      "GET",
		"http.status_code": http.StatusNotImplemented,
		"http.url":         "/test",
		"peer.hostname":    "example.com",
		"span.kind":        ext.SpanKindRPCServerEnum,
	}, span.Data.SDK.Custom.Tags)
}

func TestTracingHandlerFunc_PanicHandling(t *testing.T) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	h := instana.TracingHandlerFunc(s, "test-handler", func(w http.ResponseWriter, req *http.Request) {
		panic("something went wrong")
	})

	rec := httptest.NewRecorder()
	assert.Panics(t, func() {
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/test?q=classified", nil))
	})

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	assert.True(t, span.Error)
	assert.Equal(t, 1, span.Ec)

	require.NotNil(t, span.Data)
	require.NotNil(t, span.Data.SDK)
	assert.Equal(t, "test-handler", span.Data.SDK.Name)
	assert.Equal(t, "entry", span.Data.SDK.Type)

	require.NotNil(t, span.Data.SDK.Custom)
	assert.Equal(t, ot.Tags{
		"message":          "something went wrong",
		"http.error":       "something went wrong",
		"http.method":      "GET",
		"http.status_code": http.StatusInternalServerError,
		"http.url":         "/test",
		"peer.hostname":    "example.com",
		"span.kind":        ext.SpanKindRPCServerEnum,
	}, span.Data.SDK.Custom.Tags)

	var logRecords []map[string]interface{}
	for _, v := range span.Data.SDK.Custom.Logs {
		logRecords = append(logRecords, v)
	}

	require.Len(t, logRecords, 1)
	assert.Equal(t, "something went wrong", logRecords[0]["error"])
}
