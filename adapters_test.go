package instana_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSensor_TracingHandler_Write(t *testing.T) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{
		Service: "go-sensor-test",
	}, recorder))

	h := s.TracingHandler("test-handler", func(w http.ResponseWriter, req *http.Request) {
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

	require.IsType(t, instana.SDKSpanData{}, span.Data)
	data := span.Data.(instana.SDKSpanData)

	assert.Equal(t, "test-handler", data.Tags.Name)
	assert.Equal(t, "entry", data.Tags.Type)

	assert.Equal(t, map[string]interface{}{
		"tags": ot.Tags{
			"http.status_code": http.StatusOK,
			"http.method":      "GET",
			"http.url":         "/test",
			"peer.hostname":    "example.com",
			"span.kind":        ext.SpanKindRPCServerEnum,
		},
	}, data.Tags.Custom)
}

func TestSensor_TracingHandler_WriteHeaders(t *testing.T) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	h := s.TracingHandler("test-handler", func(w http.ResponseWriter, req *http.Request) {
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

	require.IsType(t, instana.SDKSpanData{}, span.Data)
	data := span.Data.(instana.SDKSpanData)

	assert.Equal(t, "test-handler", data.Tags.Name)
	assert.Equal(t, "entry", data.Tags.Type)

	assert.Equal(t, map[string]interface{}{
		"tags": ot.Tags{
			"http.method":      "GET",
			"http.status_code": 501,
			"http.url":         "/test",
			"peer.hostname":    "example.com",
			"span.kind":        ext.SpanKindRPCServerEnum,
		},
	}, data.Tags.Custom)
}

func TestTracingHttpRequest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		http.Error(w, "Not Found", http.StatusNotFound)
	}))
	defer ts.Close()

	tsURL, err := url.Parse(ts.URL)
	require.NoError(t, err)

	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	req, err := http.NewRequest("GET", ts.URL+"/path?q=s", nil)
	require.NoError(t, err)

	resp, err := s.TracingHttpRequest("test-request", httptest.NewRequest("GET", "/parent", nil), req, http.Client{})
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	assert.False(t, span.Error)
	assert.Equal(t, 0, span.Ec)

	require.IsType(t, instana.SDKSpanData{}, span.Data)
	data := span.Data.(instana.SDKSpanData)

	assert.Equal(t, "net/http.Client", data.Tags.Name)
	assert.Equal(t, "exit", data.Tags.Type)

	assert.Equal(t, map[string]interface{}{
		"tags": ot.Tags{
			"http.method":      "GET",
			"http.status_code": 404,
			"http.url":         ts.URL + "/path?q=s",
			"peer.hostname":    tsURL.Host,
			"span.kind":        ext.SpanKindRPCClientEnum,
		},
	}, data.Tags.Custom)
}

func TestWithTracingSpan(t *testing.T) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	s.WithTracingSpan("test-span", rec, req, func(sp ot.Span) {
		sp.SetTag("custom-tag", "value")
	})

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	assert.Empty(t, span.ParentID)
	assert.False(t, span.Error)
	assert.Equal(t, 0, span.Ec)

	require.IsType(t, instana.SDKSpanData{}, span.Data)
	data := span.Data.(instana.SDKSpanData)

	assert.Equal(t, "test-span", data.Tags.Name)
	assert.Equal(t, "entry", data.Tags.Type)

	assert.Equal(t, map[string]interface{}{
		"tags": ot.Tags{
			"http.method":   "GET",
			"http.url":      "/test",
			"peer.hostname": "example.com",
			"span.kind":     ext.SpanKindRPCServerEnum,
			"custom-tag":    "value",
		},
	}, data.Tags.Custom)
}

func TestWithTracingSpan_PanicHandling(t *testing.T) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	require.Panics(t, func() {
		s.WithTracingSpan("test-span", rec, req, func(sp ot.Span) {
			panic("something went wrong")
		})
	})

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	assert.Empty(t, span.ParentID)
	assert.True(t, span.Error)
	assert.Equal(t, 1, span.Ec)

	require.IsType(t, instana.SDKSpanData{}, span.Data)
	data := span.Data.(instana.SDKSpanData)

	assert.Equal(t, "test-span", data.Tags.Name)
	assert.Equal(t, "entry", data.Tags.Type)

	assert.Len(t, data.Tags.Custom, 2)
	assert.Equal(t, ot.Tags{
		"http.method":   "GET",
		"http.url":      "/test",
		"peer.hostname": "example.com",
		"span.kind":     ext.SpanKindRPCServerEnum,
	}, data.Tags.Custom["tags"])

	require.IsType(t, map[uint64]map[string]interface{}{}, data.Tags.Custom["logs"])
	logRecords := data.Tags.Custom["logs"].(map[uint64]map[string]interface{})

	assert.Len(t, logRecords, 1)
	for _, v := range logRecords {
		assert.Equal(t, map[string]interface{}{"error": "something went wrong"}, v)
	}
}

func TestWithTracingSpan_WithActiveParentSpan(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)
	s := instana.NewSensorWithTracer(tracer)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	parentSpan := tracer.StartSpan("parent-span")
	ctx := instana.ContextWithSpan(req.Context(), parentSpan)

	s.WithTracingSpan("test-span", rec, req.WithContext(ctx), func(sp ot.Span) {})
	parentSpan.Finish()

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 2)

	assert.Equal(t, spans[1].TraceID, spans[0].TraceID)
	assert.Equal(t, spans[1].SpanID, spans[0].ParentID)
}

func TestWithTracingSpan_WithWireContext(t *testing.T) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	traceID := instana.FormatID(1234567890)
	parentSpanID := instana.FormatID(1)

	req.Header.Set(instana.FieldT, traceID)
	req.Header.Set(instana.FieldS, parentSpanID)

	s.WithTracingSpan("test-span", rec, req, func(sp ot.Span) {})

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	assert.Equal(t, int64(1234567890), spans[0].TraceID)
	assert.Equal(t, int64(1), spans[0].ParentID)
}

func TestWithTracingContext(t *testing.T) {}
