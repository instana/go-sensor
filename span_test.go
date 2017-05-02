package instana_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/instana/golang-sensor"
	ot "github.com/opentracing/opentracing-go"
)

func TestBasicSpan(t *testing.T) {
	const op = "test"
	opts := instana.Options{LogLevel: instana.Debug}
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&opts, recorder)

	sp := tracer.StartSpan(op)
	sp.Finish()

	spans := recorder.GetSpans()
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

	spans := recorder.GetSpans()
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

	spans := recorder.GetSpans()
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

	spans := recorder.GetSpans()
	assert.Equal(t, len(spans), 1)
	span := spans[0]

	assert.NotNil(t, span.Data.SDK.Custom.Tags, "Missing Tags")
}
