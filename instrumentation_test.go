package instana_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestRoundTripper(t *testing.T) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	rt := instana.RoundTripper(s, testRoundTripper(func(req *http.Request) (*http.Response, error) {
		assert.NotEmpty(t, req.Header.Get(instana.FieldT))
		assert.NotEmpty(t, req.Header.Get(instana.FieldS))

		return &http.Response{
			Status:     http.StatusText(http.StatusNotImplemented),
			StatusCode: http.StatusNotImplemented,
		}, nil
	}))

	resp, err := rt.RoundTrip(httptest.NewRequest("GET", "http://example.com/hello", nil))
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotImplemented, resp.StatusCode)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	assert.False(t, span.Error)
	assert.Equal(t, 0, span.Ec)

	require.NotNil(t, span.Data)
	require.NotNil(t, span.Data.SDK)
	assert.Equal(t, "net/http.Client", span.Data.SDK.Name)
	assert.Equal(t, "exit", span.Data.SDK.Type)

	require.NotNil(t, span.Data.SDK.Custom)
	assert.Equal(t, ot.Tags{
		"http.method":      "GET",
		"http.status_code": http.StatusNotImplemented,
		"http.url":         "http://example.com/hello",
		"peer.hostname":    "example.com",
		"span.kind":        ext.SpanKindRPCClientEnum,
	}, span.Data.SDK.Custom.Tags)
}

func TestRoundTripper_WithParentSpan(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)
	s := instana.NewSensorWithTracer(tracer)

	span := tracer.StartSpan("parent")

	var traceIDHeader, spanIDHeader string
	rt := instana.RoundTripper(s, testRoundTripper(func(req *http.Request) (*http.Response, error) {
		traceIDHeader = req.Header.Get(instana.FieldT)
		spanIDHeader = req.Header.Get(instana.FieldS)

		return &http.Response{
			Status:     http.StatusText(http.StatusNotImplemented),
			StatusCode: http.StatusNotImplemented,
		}, nil
	}))

	ctx := instana.ContextWithSpan(context.Background(), span)
	req := httptest.NewRequest("GET", "http://example.com/hello", nil)

	_, err := rt.RoundTrip(req.WithContext(ctx))
	require.NoError(t, err)

	span.Finish()

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 2)

	assert.Equal(t, spans[1].TraceID, spans[0].TraceID)

	require.NotNil(t, spans[0].ParentID)
	assert.Equal(t, spans[1].SpanID, *spans[0].ParentID)

	traceID, err := instana.Header2ID(traceIDHeader)
	require.NoError(t, err)
	assert.Equal(t, spans[0].TraceID, traceID)

	spanID, err := instana.Header2ID(spanIDHeader)
	require.NoError(t, err)
	assert.Equal(t, spans[0].SpanID, spanID)
}

func TestRoundTripper_Error(t *testing.T) {
	serverErr := errors.New("something went wrong")

	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	rt := instana.RoundTripper(s, testRoundTripper(func(req *http.Request) (*http.Response, error) {
		return nil, serverErr
	}))

	_, err := rt.RoundTrip(httptest.NewRequest("GET", "http://example.com/hello", nil))
	assert.Error(t, err)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	assert.True(t, span.Error)
	assert.Equal(t, 1, span.Ec)

	require.NotNil(t, span.Data)
	require.NotNil(t, span.Data.SDK)
	assert.Equal(t, "net/http.Client", span.Data.SDK.Name)
	assert.Equal(t, "exit", span.Data.SDK.Type)

	require.NotNil(t, span.Data.SDK.Custom)
	assert.Equal(t, ot.Tags{
		"message":       "something went wrong",
		"http.error":    "something went wrong",
		"http.method":   "GET",
		"http.url":      "http://example.com/hello",
		"peer.hostname": "example.com",
		"span.kind":     ext.SpanKindRPCClientEnum,
	}, span.Data.SDK.Custom.Tags)

	var logRecords []map[string]interface{}
	for _, v := range span.Data.SDK.Custom.Logs {
		logRecords = append(logRecords, v)
	}

	require.Len(t, logRecords, 1)
	assert.Equal(t, serverErr, logRecords[0]["error"])
}

func TestRoundTripper_DefaultTransport(t *testing.T) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	var numCalls int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		numCalls++

		assert.NotEmpty(t, req.Header.Get(instana.FieldT))
		assert.NotEmpty(t, req.Header.Get(instana.FieldS))

		w.Write([]byte("OK"))
	}))
	defer ts.Close()

	rt := instana.RoundTripper(s, nil)

	resp, err := rt.RoundTrip(httptest.NewRequest("GET", ts.URL+"/hello", nil))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	assert.Equal(t, 1, numCalls)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	assert.False(t, span.Error)
	assert.Equal(t, 0, span.Ec)

	require.NotNil(t, span.Data)
	require.NotNil(t, span.Data.SDK)
	assert.Equal(t, "net/http.Client", span.Data.SDK.Name)
	assert.Equal(t, "exit", span.Data.SDK.Type)

	require.NotNil(t, span.Data.SDK.Custom)
	assert.Equal(t, ot.Tags{
		"http.method":      "GET",
		"http.status_code": http.StatusOK,
		"http.url":         ts.URL + "/hello",
		"peer.hostname":    strings.TrimPrefix(ts.URL, "http://"),
		"span.kind":        ext.SpanKindRPCClientEnum,
	}, span.Data.SDK.Custom.Tags)
}

type testRoundTripper func(*http.Request) (*http.Response, error)

func (rt testRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return rt(req)
}
