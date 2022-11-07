// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2017

package instana_test

import (
	"errors"
	"testing"
	"time"

	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicSpan(t *testing.T) {
	instana.InitSensor(&instana.Options{
		Service: TestServiceName,
		Tracer: instana.TracerOptions{
			CollectableHTTPHeaders: []string{"x-custom-header-1", "x-custom-header-2"},
		},
		AgentClient: alwaysReadyClient{},
	})
	defer instana.ShutdownSensor()

	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)

	start := time.Now()
	sp := tracer.StartSpan("test")
	time.Sleep(10 * time.Millisecond)
	sp.Finish()
	elapsed := time.Since(start)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)
	span := spans[0]

	assert.NotEmpty(t, span.SpanID)
	assert.NotEmpty(t, span.TraceID)
	assert.NotEmpty(t, span.Timestamp)
	assert.LessOrEqual(t, uint64(10), span.Duration)
	assert.LessOrEqual(t, span.Duration, uint64(elapsed))
	assert.Equal(t, "sdk", span.Name)

	require.IsType(t, instana.SDKSpanData{}, span.Data)
	data := span.Data.(instana.SDKSpanData)
	assert.Equal(t, TestServiceName, data.Service)

	assert.Equal(t, "test", data.Tags.Name)
	assert.Nil(t, data.Tags.Custom["tags"])
	assert.Nil(t, data.Tags.Custom["baggage"])
}

func TestSpanHeritage(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	defer instana.ShutdownSensor()

	parentSpan := tracer.StartSpan("parent")

	childSpan := tracer.StartSpan("child", ot.ChildOf(parentSpan.Context()))
	childSpan.Finish()

	parentSpan.Finish()

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 2)

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
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	defer instana.ShutdownSensor()

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
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	defer instana.ShutdownSensor()

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

func TestOTLogError(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	defer instana.ShutdownSensor()

	sp := tracer.StartSpan("test")
	ext.Error.Set(sp, true)
	sp.Finish()

	spans := recorder.GetQueuedSpans()
	require.Equal(t, len(spans), 1)

	span := spans[0]
	assert.Equal(t, 1, span.Ec)

	require.IsType(t, instana.SDKSpanData{}, span.Data)
	data := span.Data.(instana.SDKSpanData)

	assert.Equal(t, map[string]interface{}{
		"tags": ot.Tags{"error": true},
	}, data.Tags.Custom)
}

func TestSpanErrorLogKV(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	defer instana.ShutdownSensor()

	sp := tracer.StartSpan("test")
	sp.LogKV("error", "simulated error")
	sp.Finish()

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 2)

	span, logSpan := spans[0], spans[1]
	assert.Equal(t, 1, span.Ec)

	assert.Equal(t, span.TraceID, logSpan.TraceID)
	assert.Equal(t, span.SpanID, logSpan.ParentID)
	assert.Equal(t, "log.go", logSpan.Name)

	// assert that log message has been recorded within the span interval
	assert.GreaterOrEqual(t, logSpan.Timestamp, span.Timestamp)
	assert.LessOrEqual(t, logSpan.Duration, span.Duration)

	require.IsType(t, instana.LogSpanData{}, logSpan.Data)
	logData := logSpan.Data.(instana.LogSpanData)

	assert.Equal(t, instana.LogSpanTags{
		Level:   "ERROR",
		Message: `error: "simulated error"`,
	}, logData.Tags)
}

func TestSpan_LogFields(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	defer instana.ShutdownSensor()

	examples := map[string]struct {
		Fields             []log.Field
		ExpectedErrorCount int
		ExpectedTags       instana.LogSpanTags
	}{
		"error object": {
			Fields: []log.Field{
				log.Error(errors.New("simulated error")),
				log.String("function", "ErrorFunc"),
			},
			ExpectedErrorCount: 1,
			ExpectedTags: instana.LogSpanTags{
				Level:   "ERROR",
				Message: `error: "simulated error" function: "ErrorFunc"`,
			},
		},
		"error log": {
			Fields: []log.Field{
				log.String("error", "simulated error"),
				log.String("function", "ErrorFunc"),
			},
			ExpectedErrorCount: 1,
			ExpectedTags: instana.LogSpanTags{
				Level:   "ERROR",
				Message: `error: "simulated error" function: "ErrorFunc"`,
			},
		},
		"warn log": {
			Fields: []log.Field{
				log.String("warn", "simulated warning"),
				log.String("function", "WarnFunc"),
			},
			ExpectedTags: instana.LogSpanTags{
				Level:   "WARN",
				Message: `warn: "simulated warning" function: "WarnFunc"`,
			},
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			sp := tracer.StartSpan("test")
			sp.LogFields(example.Fields...)
			sp.Finish()

			spans := recorder.GetQueuedSpans()
			require.Len(t, spans, 2)

			span, logSpan := spans[0], spans[1]
			assert.Equal(t, example.ExpectedErrorCount, span.Ec)

			assert.Equal(t, span.TraceID, logSpan.TraceID)
			assert.Equal(t, span.SpanID, logSpan.ParentID)
			assert.Equal(t, "log.go", logSpan.Name)

			// assert that log message has been recorded within the span interval
			assert.GreaterOrEqual(t, logSpan.Timestamp, span.Timestamp)
			assert.LessOrEqual(t, logSpan.Duration, span.Duration)

			require.IsType(t, instana.LogSpanData{}, logSpan.Data)
			logData := logSpan.Data.(instana.LogSpanData)

			assert.Equal(t, example.ExpectedTags, logData.Tags)
		})
	}
}

func TestSpan_Suppressed_StartSpanOption(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	defer instana.ShutdownSensor()

	sp := tracer.StartSpan("test", instana.SuppressTracing())
	sp.Finish()

	assert.Empty(t, recorder.GetQueuedSpans())
}

func TestSpan_Suppressed_SetTag(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	defer instana.ShutdownSensor()

	sp := tracer.StartSpan("test")
	instana.SuppressTracing().Set(sp)
	sp.Finish()

	assert.Empty(t, recorder.GetQueuedSpans())
}
