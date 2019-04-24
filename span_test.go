package instana_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
	ext "github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"github.com/stretchr/testify/assert"
)

func TestBasicSpan(t *testing.T) {
	const op = "test"
	opts := instana.Options{LogLevel: instana.Debug}
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&opts, recorder)

	sp := tracer.StartSpan(op)
	sp.Finish()

	spans := recorder.GetQueuedSpans()
	assert.Equal(t, 1, len(spans))
	span := spans[0]

	assert.NotZero(t, span.SpanID, "Missing span ID for this span")
	assert.NotZero(t, span.TraceID, "Missing trace ID for this span")
	assert.NotZero(t, span.Timestamp, "Missing timestamp for this span")
	assert.NotNil(t, span.Duration, "Duration is nil")
	assert.Equal(t, "sdk", span.Name, "Missing sdk span name")
	assert.Equal(t, "test", span.Data.SDK.Name, "Missing span name")
	assert.Nil(t, span.Data.SDK.Custom.Tags, "Tags has an unexpected value")
	assert.Nil(t, span.Data.SDK.Custom.Baggage, "Baggage has an unexpected value")
	assert.Equal(t, "go", span.Lang, "Missing or wrong ta/lang")
}

func TestSpanHeritage(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

	parentSpan := tracer.StartSpan("parent")

	childSpan := tracer.StartSpan("child", ot.ChildOf(parentSpan.Context()))
	time.Sleep(2 * time.Millisecond)
	childSpan.Finish()

	time.Sleep(2 * time.Millisecond)
	parentSpan.Finish()

	spans := recorder.GetQueuedSpans()
	assert.Equal(t, len(spans), 2)
	cSpan := spans[0]
	pSpan := spans[1]

	assert.Equal(t, "child", cSpan.Data.SDK.Name, "Child span name doesn't compute")
	assert.Equal(t, "parent", pSpan.Data.SDK.Name, "Parent span name doesn't compute")

	// Parent should not have a parent
	assert.Nil(t, pSpan.ParentID, "ParentID shouldn't have a value")

	// Child must have parent ID set as parent
	assert.Equal(t, pSpan.SpanID, *cSpan.ParentID, "parentID doesn't match")

	// Must be root span
	assert.Equal(t, pSpan.TraceID, pSpan.SpanID, "not a root span")

	// Trace ID must be consistent across spans
	assert.Equal(t, cSpan.TraceID, pSpan.TraceID, "trace IDs don't match")

}

func TestSpanBaggage(t *testing.T) {
	const op = "test"
	opts := instana.Options{LogLevel: instana.Debug}
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&opts, recorder)

	sp := tracer.StartSpan(op)
	sp.SetBaggageItem("foo", "bar")
	sp.Finish()

	spans := recorder.GetQueuedSpans()
	assert.Equal(t, len(spans), 1)
	span := spans[0]

	assert.NotNil(t, span.Data.SDK.Custom.Baggage, "Missing Baggage")
}

func TestSpanTags(t *testing.T) {
	const op = "test"
	opts := instana.Options{LogLevel: instana.Debug}
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&opts, recorder)

	sp := tracer.StartSpan(op)
	sp.SetTag("foo", "bar")
	sp.Finish()

	spans := recorder.GetQueuedSpans()
	assert.Equal(t, len(spans), 1)
	span := spans[0]

	assert.NotNil(t, span.Data.SDK.Custom.Tags, "Missing Tags")
}

func TestSpanLogFields(t *testing.T) {
	const op = "test"
	opts := instana.Options{LogLevel: instana.Debug}
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&opts, recorder)

	span := tracer.StartSpan(op)
	span.LogFields(
		log.String("event", "soft error"),
		log.String("type", "cache timeout"),
		log.Int("waited.millis", 1500))
	span.Finish()

	spans := recorder.GetQueuedSpans()
	assert.Equal(t, len(spans), 1)
	firstSpan := spans[0]

	// j, _ := json.MarshalIndent(spans, "", "  ")
	// fmt.Printf("spans:", bytes.NewBuffer(j))

	logData := firstSpan.Data.SDK.Custom.Logs
	assert.NotNil(t, logData, "Missing logged fields")
	assert.Equal(t, 1, len(logData), "Unexpected log count")

	for _, v := range logData {
		assert.Equal(t, "soft error", v["event"], "Wrong or missing log")
		assert.Equal(t, "cache timeout", v["type"], "Wrong or missing log")
		assert.Equal(t, 1500, v["waited.millis"], "Wrong or missing log")
	}
}

func TestSpanLogKVs(t *testing.T) {
	const op = "test"
	opts := instana.Options{LogLevel: instana.Debug}
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&opts, recorder)

	span := tracer.StartSpan(op)
	span.LogKV(
		"event", "soft error",
		"type", "cache timeout",
		"waited.millis", 1500)
	span.Finish()

	spans := recorder.GetQueuedSpans()
	assert.Equal(t, len(spans), 1)
	firstSpan := spans[0]

	// j, _ := json.MarshalIndent(spans, "", "  ")
	// fmt.Printf("spans:", bytes.NewBuffer(j))

	logData := firstSpan.Data.SDK.Custom.Logs
	assert.NotNil(t, logData, "Missing logged fields")
	assert.Equal(t, 1, len(logData), "Unexpected log count")

	for _, v := range logData {
		assert.Equal(t, "soft error", v["event"], "Wrong or missing log")
		assert.Equal(t, "cache timeout", v["type"], "Wrong or missing log")
		assert.Equal(t, 1500, v["waited.millis"], "Wrong or missing log")
	}
}

func TestOTLogError(t *testing.T) {
	const op = "test"
	opts := instana.Options{LogLevel: instana.Debug}
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&opts, recorder)

	span := tracer.StartSpan(op)
	ext.Error.Set(span, true)
	span.Finish()

	spans := recorder.GetQueuedSpans()
	assert.Equal(t, len(spans), 1)
	firstSpan := spans[0]

	logData := firstSpan.Data.SDK.Custom.Logs
	tagData := firstSpan.Data.SDK.Custom.Tags
	assert.Equal(t, 1, len(tagData), "Unexpected log count")
	assert.Equal(t, true, firstSpan.Error, "Span should be marked as errored")
	assert.Equal(t, 1, firstSpan.Ec, "Error count should be 1")

	for _, v := range logData {
		for sk, sv := range v {
			fmt.Print(v)
			assert.Equal(t, "error", sk, "Wrong or missing log")
			assert.Equal(t, "simulated error", sv, "Wrong or missing log")
		}
	}
}

func TestSpanErrorLogKV(t *testing.T) {
	const op = "test"
	opts := instana.Options{LogLevel: instana.Debug}
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&opts, recorder)

	span := tracer.StartSpan(op)
	span.LogKV("error", "simulated error")
	span.Finish()

	spans := recorder.GetQueuedSpans()
	assert.Equal(t, len(spans), 1)
	firstSpan := spans[0]

	assert.Equal(t, 1, firstSpan.Ec, "Error count should be 1")
	assert.Equal(t, true, firstSpan.Error, "Span should be marked as errored")

	logData := firstSpan.Data.SDK.Custom.Logs
	assert.NotNil(t, logData, "Missing logged fields")
	assert.Equal(t, 1, len(logData), "Unexpected log count")

	for _, v := range logData {
		for sk, sv := range v {
			assert.Equal(t, "error", sk, "Wrong or missing log")
			assert.Equal(t, "simulated error", sv, "Wrong or missing log")
		}
	}
}

func TestSpanErrorLogFields(t *testing.T) {
	const op = "test"
	opts := instana.Options{LogLevel: instana.Debug}
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&opts, recorder)

	span := tracer.StartSpan(op)

	err := errors.New("simulated error")
	span.LogFields(log.Error(err), log.String("function", "TestspanErrorLogFields"))
	span.LogFields(log.Error(err), log.String("function", "TestspanErrorLogFields"))
	span.Finish()

	spans := recorder.GetQueuedSpans()
	assert.Equal(t, len(spans), 1)
	firstSpan := spans[0]

	logData := firstSpan.Data.SDK.Custom.Logs
	assert.Equal(t, 1, len(logData), "Unexpected tag count")
	assert.Equal(t, true, firstSpan.Error, "Span should be marked as errored")
	assert.Equal(t, 2, firstSpan.Ec, "Error count should be 2")
}
