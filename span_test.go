package instana_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/instana/golang-sensor"
	bt "github.com/opentracing/basictracer-go"
	ot "github.com/opentracing/opentracing-go"
)

func TestBasicSpan(t *testing.T) {
	const op = "test"
	recorder := bt.NewInMemoryRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

	sp := tracer.StartSpan(op)
	sp.Finish()

	spans := recorder.GetSpans()
	assert.Equal(t, len(spans), 1)
	span := spans[0]

	assert.NotNil(t, span.Context, "Context is nil!")
	assert.NotNil(t, span.Duration, "Duration is nil!")
	assert.NotNil(t, span.Operation, "Operation is nil!")
	assert.Equal(t, span.ParentSpanID, uint64(0), "ParentSpan shouldn't have a value")
	assert.NotNil(t, span.Start, "Start is nil!")
	assert.Nil(t, span.Tags, "Tags is nil!")
}

func TestSpanHeritage(t *testing.T) {
	recorder := bt.NewInMemoryRecorder()
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

	assert.Equal(t, "child", cSpan.Operation, "Child span name doesn't compute")
	assert.Equal(t, "parent", pSpan.Operation, "Parent span name doesn't compute")

	// Parent should not have a parent
	assert.Equal(t, pSpan.ParentSpanID, uint64(0), "ParentSpanID shouldn't have a value")

	// Child must have parent ID set as parent
	assert.Equal(t, pSpan.Context.SpanID, cSpan.ParentSpanID, "parentID doesn't match")

	// Must be root span
	assert.Equal(t, pSpan.Context.TraceID, pSpan.Context.SpanID, "not a root span")

	// Trace ID must be consistent across spans
	assert.Equal(t, cSpan.Context.TraceID, pSpan.Context.TraceID, "trace IDs don't match")

}

func TestSpanBaggage(t *testing.T) {
	const op = "test"
	recorder := bt.NewInMemoryRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

	sp := tracer.StartSpan(op)
	sp.SetBaggageItem("foo", "bar")
	sp.Finish()

	spans := recorder.GetSpans()
	assert.Equal(t, len(spans), 1)
	span := spans[0]

	assert.NotNil(t, span.Context.Baggage, "Missing Baggage")
}

func TestSpanTags(t *testing.T) {
	const op = "test"
	recorder := bt.NewInMemoryRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

	sp := tracer.StartSpan(op)
	sp.SetTag("foo", "bar")
	sp.Finish()

	spans := recorder.GetSpans()
	assert.Equal(t, len(spans), 1)
	span := spans[0]

	assert.NotNil(t, span.Tags, "Missing Tags")
}
