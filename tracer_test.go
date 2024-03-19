// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2017

package instana_test

import (
	"testing"

	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
)

func TestTracerAPI(t *testing.T) {
	tracer := instana.NewTracer()
	assert.NotNil(t, tracer)

	recorder := instana.NewTestRecorder()

	tracer = instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	defer instana.ShutdownSensor()
	assert.NotNil(t, tracer)

	tracer = instana.NewTracerWithOptions(&instana.Options{AgentClient: alwaysReadyClient{}})
	assert.NotNil(t, tracer)
}

func TestTracerBasics(t *testing.T) {
	opts := instana.Options{LogLevel: instana.Debug, AgentClient: alwaysReadyClient{}}
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&opts, recorder)
	defer instana.ShutdownSensor()

	sp := tracer.StartSpan("test")
	sp.SetBaggageItem("foo", "bar")
	sp.Finish()

	spans := recorder.GetQueuedSpans()
	assert.Equal(t, len(spans), 1)
}

func TestTracer_StartSpan_SuppressTracing(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	defer instana.ShutdownSensor()

	sp := tracer.StartSpan("test", instana.SuppressTracing())

	sc := sp.Context().(instana.SpanContext)
	assert.True(t, sc.Suppressed)
}

func TestTracer_StartSpan_WithCorrelationData(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	defer instana.ShutdownSensor()

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
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, nil)
	defer instana.ShutdownSensor()

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
