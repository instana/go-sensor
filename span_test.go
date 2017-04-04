package instana_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/instana/golang-sensor"
	bt "github.com/opentracing/basictracer-go"
	//opentracing "github.com/opentracing/opentracing-go"
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
	assert.NotNil(t, span.ParentSpanID, "ParentSpan is nil!")
	assert.NotNil(t, span.Start, "Start is nil!")
	assert.Nil(t, span.Tags, "Tags is nil!")
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
