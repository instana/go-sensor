package instana_test

import (
	"testing"

	instana "github.com/instana/go-sensor"
	"github.com/stretchr/testify/assert"
	//opentracing "github.com/opentracing/opentracing-go"
)

func TestTracerAPI(t *testing.T) {
	tracer := instana.NewTracer()
	assert.NotNil(t, tracer, "NewTracer returned nil")

	opts := instana.Options{LogLevel: instana.Debug}
	recorder := instana.NewTestRecorder()
	tracer = instana.NewTracerWithEverything(&opts, recorder)
	assert.NotNil(t, tracer, "NewTracerWithEverything returned nil")

	tracer = instana.NewTracerWithOptions(&instana.Options{})
	assert.NotNil(t, tracer, "NewTracerWithOptions returned nil")
}

func TestTracerBasics(t *testing.T) {
	opts := instana.Options{LogLevel: instana.Debug}
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&opts, recorder)

	sp := tracer.StartSpan("test")
	sp.SetBaggageItem("foo", "bar")
	sp.Finish()

	spans := recorder.GetQueuedSpans()
	assert.Equal(t, len(spans), 1)
}
