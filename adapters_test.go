// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

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
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder))
	defer instana.ShutdownSensor()

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

func TestWithTracingSpan_PanicHandling(t *testing.T) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder))
	defer instana.ShutdownSensor()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	require.Panics(t, func() {
		s.WithTracingSpan("test-span", rec, req, func(sp ot.Span) {
			panic("something went wrong")
		})
	})

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 2)

	span, logSpan := spans[0], spans[1]
	assert.Empty(t, span.ParentID)
	assert.Equal(t, 1, span.Ec)

	require.IsType(t, instana.SDKSpanData{}, span.Data)
	data := span.Data.(instana.SDKSpanData)

	assert.Equal(t, "test-span", data.Tags.Name)
	assert.Equal(t, "entry", data.Tags.Type)

	assert.Len(t, data.Tags.Custom, 1)
	assert.Equal(t, ot.Tags{
		"http.method":   "GET",
		"http.url":      "/test",
		"peer.hostname": "example.com",
		"span.kind":     ext.SpanKindRPCServerEnum,
	}, data.Tags.Custom["tags"])

	assert.Equal(t, span.TraceID, logSpan.TraceID)
	assert.Equal(t, span.SpanID, logSpan.ParentID)
	assert.Equal(t, "log.go", logSpan.Name)

	// assert that log message has been recorded within the span interval
	assert.GreaterOrEqual(t, logSpan.Timestamp, span.Timestamp)
	assert.LessOrEqual(t, logSpan.Duration, span.Duration)

	require.IsType(t, instana.LogSpanData{}, logSpan.Data)
	logData := logSpan.Data.(instana.LogSpanData)

	assert.Equal(t, instana.LogSpanTags{
		Level:   "ERROR",
		Message: `error: "something went wrong"`,
	}, logData.Tags)
}

func TestWithTracingSpan_WithActiveParentSpan(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	s := instana.NewSensorWithTracer(tracer)
	defer instana.ShutdownSensor()

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
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder))
	defer instana.ShutdownSensor()

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
