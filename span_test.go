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
	return
	span := spans[0]

	assert.NotNil(t, span.Raw.Context, "Context is nil!")
	assert.NotNil(t, span.Duration, "Duration is nil!")
	assert.NotNil(t, span.Raw.Operation, "Operation is nil!")
	assert.Equal(t, span.Raw.ParentSpanID, uint64(0), "ParentSpan shouldn't have a value")
	assert.NotNil(t, span.Raw.Start, "Start is nil!")
	assert.Nil(t, span.Raw.Tags, "Tags is nil!")
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

	assert.Equal(t, "child", cSpan.Raw.Operation, "Child span name doesn't compute")
	assert.Equal(t, "parent", pSpan.Raw.Operation, "Parent span name doesn't compute")

	// Parent should not have a parent
	assert.Equal(t, int64(0), pSpan.Raw.ParentSpanID, "ParentSpanID shouldn't have a value")

	// Child must have parent ID set as parent
	assert.Equal(t, pSpan.Raw.Context.SpanID, cSpan.Raw.ParentSpanID, "parentID doesn't match")

	// Must be root span
	assert.Equal(t, pSpan.Raw.Context.TraceID, pSpan.Raw.Context.SpanID, "not a root span")

	// Trace ID must be consistent across spans
	assert.Equal(t, cSpan.Raw.Context.TraceID, pSpan.Raw.Context.TraceID, "trace IDs don't match")

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

	assert.NotNil(t, span.Raw.Context.Baggage, "Missing Baggage")
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

	assert.NotNil(t, span.Raw.Tags, "Missing Tags")
}
