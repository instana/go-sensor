package instana_test

import (
	"errors"
	"testing"
	"time"

	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
	ext "github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicSpan(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

	sp := tracer.StartSpan("test")
	time.Sleep(2 * time.Millisecond)
	sp.Finish()

	spans := recorder.GetQueuedSpans()
	assert.Equal(t, 1, len(spans))
	span := spans[0]

	assert.NotEmpty(t, span.SpanID)
	assert.NotEmpty(t, span.TraceID)
	assert.NotEmpty(t, span.Timestamp)
	assert.Equal(t, uint64(2), span.Duration)
	assert.Equal(t, "sdk", span.Name)

	require.IsType(t, instana.SDKSpanData{}, span.Data)
	data := span.Data.(instana.SDKSpanData)
	assert.Equal(t, "go-sensor.test", data.Service)

	assert.Equal(t, "test", data.Tags.Name)
	assert.Nil(t, data.Tags.Custom["tags"])
	assert.Nil(t, data.Tags.Custom["baggage"])
	assert.Equal(t, "go", span.Lang)
}

func TestSpanHeritage(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

	parentSpan := tracer.StartSpan("parent")

	childSpan := tracer.StartSpan("child", ot.ChildOf(parentSpan.Context()))
	childSpan.Finish()

	parentSpan.Finish()

	spans := recorder.GetQueuedSpans()
	assert.Equal(t, len(spans), 2)

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
	assert.True(t, span.Error)
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
	require.Len(t, spans, 1)
	span := spans[0]

	assert.Equal(t, 1, span.Ec)
	assert.True(t, span.Error)

	require.IsType(t, instana.SDKSpanData{}, span.Data)
	data := span.Data.(instana.SDKSpanData)

	require.IsType(t, map[uint64]map[string]interface{}{}, data.Tags.Custom["logs"])
	logRecords := data.Tags.Custom["logs"].(map[uint64]map[string]interface{})

	require.Len(t, logRecords, 1)
	for _, v := range logRecords {
		assert.Equal(t, map[string]interface{}{"error": "simulated error"}, v)
	}
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
	require.Len(t, spans, 1)

	span := spans[0]
	assert.True(t, span.Error)
	assert.Equal(t, 2, span.Ec)

	require.IsType(t, instana.SDKSpanData{}, span.Data)
	data := span.Data.(instana.SDKSpanData)

	require.IsType(t, map[uint64]map[string]interface{}{}, data.Tags.Custom["logs"])
	logRecords := data.Tags.Custom["logs"].(map[uint64]map[string]interface{})

	assert.Len(t, logRecords, 1)
}
