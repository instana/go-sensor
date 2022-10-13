// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2017

package instana_test

import (
	"testing"

	instana "github.com/instana/go-sensor"
	"github.com/instana/testify/assert"
	ot "github.com/opentracing/opentracing-go"
)

func TestTracerAPI(t *testing.T) {
	tracer := instana.NewTracer()
	assert.NotNil(t, tracer)

	recorder := instana.NewTestRecorder()

	tracer = instana.NewTracerWithEverything(&instana.Options{}, recorder)
	defer instana.TestOnlyStopSensor()
	assert.NotNil(t, tracer)

	tracer = instana.NewTracerWithOptions(&instana.Options{})
	assert.NotNil(t, tracer)
}

func TestTracerBasics(t *testing.T) {
	opts := instana.Options{LogLevel: instana.Debug}
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&opts, recorder)
	defer instana.TestOnlyStopSensor()

	sp := tracer.StartSpan("test")
	sp.SetBaggageItem("foo", "bar")
	sp.Finish()

	spans := recorder.GetQueuedSpans()
	assert.Equal(t, len(spans), 1)
}

func TestTracer_StartSpan_SuppressTracing(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)
	defer instana.TestOnlyStopSensor()

	sp := tracer.StartSpan("test", instana.SuppressTracing())

	sc := sp.Context().(instana.SpanContext)
	assert.True(t, sc.Suppressed)
}

func TestTracer_StartSpan_WithCorrelationData(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)
	defer instana.TestOnlyStopSensor()

	sp := tracer.StartSpan("test", ot.ChildOf(instana.SpanContext{
		Correlation: instana.EUMCorrelationData{
			Type: "type1",
			ID:   "id1",
		},
	}))

	sc := sp.Context().(instana.SpanContext)
	assert.Equal(t, instana.EUMCorrelationData{}, sc.Correlation)
}

type strangeContext struct{}

func (c *strangeContext) ForeachBaggageItem(handler func(k, v string) bool) {}

func TestTracer_NonInstanaSpan(t *testing.T) {
	tracer := instana.NewTracerWithEverything(&instana.Options{}, nil)
	defer instana.TestOnlyStopSensor()

	ref := ot.SpanReference{
		Type:              ot.ChildOfRef,
		ReferencedContext: &strangeContext{},
	}

	opts := ot.StartSpanOptions{
		References: []ot.SpanReference{
			ref,
		},
	}

	assert.NotPanics(t, func() {
		tracer.StartSpanWithOptions("my_operation", opts)
	})
}
