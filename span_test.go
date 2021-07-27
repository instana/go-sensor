// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2017

package instana_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	instana "github.com/instana/go-sensor"
	"github.com/instana/testify/assert"
	"github.com/instana/testify/require"
	ot "github.com/opentracing/opentracing-go"
	ext "github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
)

func TestBasicSpan(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

	sp := tracer.StartSpan("test")
	time.Sleep(10 * time.Millisecond)
	sp.Finish()

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)
	span := spans[0]

	assert.NotEmpty(t, span.SpanID)
	assert.NotEmpty(t, span.TraceID)
	assert.NotEmpty(t, span.Timestamp)
	assert.InDelta(t, uint64(10), span.Duration, 5.0)
	assert.Equal(t, "sdk", span.Name)

	require.IsType(t, instana.SDKSpanData{}, span.Data)
	data := span.Data.(instana.SDKSpanData)
	assert.Equal(t, TestServiceName, data.Service)

	assert.Equal(t, "test", data.Tags.Name)
	assert.Nil(t, data.Tags.Custom["tags"])
	assert.Nil(t, data.Tags.Custom["baggage"])
}

func TestSpanHeritage(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

	parentSpan := tracer.StartSpan("parent")

	childSpan := tracer.StartSpan("child", ot.ChildOf(parentSpan.Context()))
	childSpan.Finish()

	parentSpan.Finish()

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 2)

	cSpan, pSpan := spans[0], spans[1]

	// Parent should not have a parent
	assert.Empty(t, pSpan.ParentID)

	// Child must have parent ID set as parent
	assert.Equal(t, pSpan.SpanID, cSpan.ParentID)

	// Must be root span
	assert.Equal(t, pSpan.TraceID, pSpan.SpanID)

	// Trace ID must be consistent across spans
	assert.Equal(t, cSpan.TraceID, pSpan.TraceID)

	require.IsType(t, cSpan.Data, instana.SDKSpanData{})
	cData := cSpan.Data.(instana.SDKSpanData)
	assert.Equal(t, "child", cData.Tags.Name)

	require.IsType(t, pSpan.Data, instana.SDKSpanData{})
	pData := pSpan.Data.(instana.SDKSpanData)
	assert.Equal(t, "parent", pData.Tags.Name)
}

func TestSpanBaggage(t *testing.T) {
	const op = "test"
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

	sp := tracer.StartSpan(op)
	sp.SetBaggageItem("foo", "bar")
	sp.Finish()

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)
	span := spans[0]

	require.IsType(t, instana.SDKSpanData{}, span.Data)
	data := span.Data.(instana.SDKSpanData)

	assert.Equal(t, map[string]string{"foo": "bar"}, data.Tags.Custom["baggage"])
}

func TestSpanTags(t *testing.T) {
	const op = "test"
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

	sp := tracer.StartSpan(op)
	sp.SetTag("foo", "bar")
	sp.Finish()

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)
	span := spans[0]

	require.IsType(t, instana.SDKSpanData{}, span.Data)
	data := span.Data.(instana.SDKSpanData)

	assert.Equal(t, ot.Tags{"foo": "bar"}, data.Tags.Custom["tags"])
}

func TestSpanLogFields(t *testing.T) {
	const op = "test"
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

	sp := tracer.StartSpan(op)
	sp.LogFields(
		log.String("event", "soft error"),
		log.String("type", "cache timeout"),
		log.Int("waited.millis", 1500),
	)
	sp.Finish()

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)
	span := spans[0]

	require.IsType(t, instana.SDKSpanData{}, span.Data)
	data := span.Data.(instana.SDKSpanData)

	require.IsType(t, map[uint64]map[string]interface{}{}, data.Tags.Custom["logs"])
	logRecords := data.Tags.Custom["logs"].(map[uint64]map[string]interface{})

	require.Len(t, logRecords, 1)
	for _, v := range logRecords {
		assert.Equal(t, map[string]interface{}{
			"event":         "soft error",
			"type":          "cache timeout",
			"waited.millis": 1500,
		}, v)
	}
}

func TestSpanLogKVs(t *testing.T) {
	const op = "test"
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

	sp := tracer.StartSpan(op)
	sp.LogKV(
		"event", "soft error",
		"type", "cache timeout",
		"waited.millis", 1500,
	)
	sp.Finish()

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)
	span := spans[0]

	require.IsType(t, instana.SDKSpanData{}, span.Data)
	data := span.Data.(instana.SDKSpanData)

	require.IsType(t, map[uint64]map[string]interface{}{}, data.Tags.Custom["logs"])
	logRecords := data.Tags.Custom["logs"].(map[uint64]map[string]interface{})

	require.Len(t, logRecords, 1)
	for _, v := range logRecords {
		assert.Equal(t, map[string]interface{}{
			"event":         "soft error",
			"type":          "cache timeout",
			"waited.millis": 1500,
		}, v)
	}
}

func TestOTLogError(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

	sp := tracer.StartSpan("test")
	ext.Error.Set(sp, true)
	sp.Finish()

	spans := recorder.GetQueuedSpans()
	require.Equal(t, len(spans), 1)

	span := spans[0]
	assert.Equal(t, 1, span.Ec)

	require.IsType(t, instana.SDKSpanData{}, span.Data)
	data := span.Data.(instana.SDKSpanData)

	assert.Equal(t, map[string]interface{}{
		"tags": ot.Tags{"error": true},
	}, data.Tags.Custom)
}

func TestSpanErrorLogKV(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

	sp := tracer.StartSpan("test")
	sp.LogKV("error", "simulated error")
	sp.Finish()

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 2)

	span, logSpan := spans[0], spans[1]
	assert.Equal(t, 1, span.Ec)

	require.IsType(t, instana.SDKSpanData{}, span.Data)
	data := span.Data.(instana.SDKSpanData)

	require.IsType(t, map[uint64]map[string]interface{}{}, data.Tags.Custom["logs"])
	logRecords := data.Tags.Custom["logs"].(map[uint64]map[string]interface{})

	require.Len(t, logRecords, 1)
	for _, v := range logRecords {
		assert.Equal(t, map[string]interface{}{"error": "simulated error"}, v)
	}

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
		Message: `error: "simulated error"`,
	}, logData.Tags)
}

func TestSpanErrorLogFields(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

	sp := tracer.StartSpan("test")

	err := errors.New("simulated error")
	sp.LogFields(log.Error(err), log.String("function", "TestspanErrorLogFields"))
	sp.LogFields(log.Error(err), log.String("function", "TestspanErrorLogFields"))
	sp.Finish()

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 3)

	span, logSpans := spans[0], spans[1:]
	assert.Equal(t, 2, span.Ec)

	require.IsType(t, instana.SDKSpanData{}, span.Data)
	data := span.Data.(instana.SDKSpanData)

	require.IsType(t, map[uint64]map[string]interface{}{}, data.Tags.Custom["logs"])
	logRecords := data.Tags.Custom["logs"].(map[uint64]map[string]interface{})

	assert.Len(t, logRecords, 1)

	require.Len(t, logSpans, 2)
	for i, logSpan := range logSpans {
		assert.Equal(t, span.TraceID, logSpan.TraceID, fmt.Sprintf("log span %d", i))
		assert.Equal(t, span.SpanID, logSpan.ParentID, fmt.Sprintf("log span %d", i))
		assert.Equal(t, "log.go", logSpan.Name, fmt.Sprintf("log span %d", i))

		// assert that log message has been recorded within the span interval
		assert.GreaterOrEqual(t, logSpan.Timestamp, span.Timestamp, fmt.Sprintf("log span %d", i))
		assert.LessOrEqual(t, logSpan.Duration, span.Duration, fmt.Sprintf("log span %d", i))

		require.IsType(t, instana.LogSpanData{}, logSpan.Data, fmt.Sprintf("log span %d", i))
		logData := logSpan.Data.(instana.LogSpanData)

		assert.Equal(t, instana.LogSpanTags{
			Level:   "ERROR",
			Message: `error: "simulated error" function: "TestspanErrorLogFields"`,
		}, logData.Tags, fmt.Sprintf("log span %d", i))
	}
}

func TestSpan_Suppressed_StartSpanOption(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

	sp := tracer.StartSpan("test", instana.SuppressTracing())
	sp.Finish()

	assert.Empty(t, recorder.GetQueuedSpans())
}

func TestSpan_Suppressed_SetTag(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

	sp := tracer.StartSpan("test")
	instana.SuppressTracing().Set(sp)
	sp.Finish()

	assert.Empty(t, recorder.GetQueuedSpans())
}
