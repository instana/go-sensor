package instana_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/instana/golang-sensor"
	//opentracing "github.com/opentracing/opentracing-go"
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
	assert.NotNil(t, span.Raw.ParentSpanID, "ParentSpan is nil!")
	assert.NotNil(t, span.Raw.Start, "Start is nil!")
	assert.Nil(t, span.Raw.Tags, "Tags is nil!")
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
