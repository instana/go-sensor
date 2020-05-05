package instana_test

import (
	"net/http/httptest"
	"testing"

	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func TestWithTracingSyntheticSpan(t *testing.T) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("x-instana-synthetic", "1")

	s.WithTracingSpan("test-span", rec, req, func(sp ot.Span) {
		sp.SetTag("custom-tag", "value")
	})

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	assert.Empty(t, span.ParentID)
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
			"sy":            true,
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
