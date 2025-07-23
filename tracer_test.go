// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2017

package instana_test

import (
	"fmt"
	"testing"

	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestTracerLogSpans(t *testing.T) {
	type args struct {
		fields        []otlog.Field
		TracerOptions instana.TracerOptions
	}

	type expectedResult struct {
		errorCount int
		spanCount  int
	}

	testCases := []struct {
		desc           string
		args           args
		expectedResult expectedResult
	}{
		{
			desc: "log span count equals default MaxLogsPerSpan limit",
			args: args{
				fields: []otlog.Field{
					otlog.Error(fmt.Errorf("error1")),
					otlog.Error(fmt.Errorf("error2")),
				},
				TracerOptions: instana.DefaultTracerOptions(),
			},
			expectedResult: expectedResult{
				errorCount: 2,
				spanCount:  3, // span + 2 log spans
			},
		},
		{
			desc: "log span count is greater than default MaxLogsPerSpan limit",
			args: args{
				fields: []otlog.Field{
					otlog.Error(fmt.Errorf("error1")),
					otlog.Error(fmt.Errorf("error2")),
					otlog.Error(fmt.Errorf("error2")),
				},
				TracerOptions: instana.DefaultTracerOptions(),
			},
			expectedResult: expectedResult{
				errorCount: 3,
				spanCount:  3, // span + 2 log spans
			},
		},
		{
			desc: "log span count is 2, MaxLogsPerSpan is 1",
			args: args{
				fields: []otlog.Field{
					otlog.Error(fmt.Errorf("error1")),
					otlog.Error(fmt.Errorf("error2")),
				},
				TracerOptions: instana.TracerOptions{
					MaxLogsPerSpan: 1,
				},
			},
			expectedResult: expectedResult{
				errorCount: 2,
				spanCount:  2, // span + 1 log span
			},
		},
		{
			desc: "log span count is 2, MaxLogsPerSpan is set to 0",
			args: args{
				fields: []otlog.Field{
					otlog.Error(fmt.Errorf("error1")),
					otlog.Error(fmt.Errorf("error2")),
				},
				TracerOptions: instana.TracerOptions{
					MaxLogsPerSpan: 0,
				},
			},
			expectedResult: expectedResult{
				errorCount: 2,
				spanCount:  3, // span + 2 log span (as default MaxLogsPerSpan value will be set here)
			},
		},
		{
			desc: "appended logs of level greater than warn",
			args: args{
				fields: []otlog.Field{
					otlog.Object("debug", "log1"),
					otlog.Object("debug", "log2"),
				},
				TracerOptions: instana.DefaultTracerOptions(),
			},
			expectedResult: expectedResult{
				errorCount: 0,
				spanCount:  1, // span + no spans created for debug logs
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {

			opts := instana.Options{
				LogLevel:    instana.Debug,
				AgentClient: alwaysReadyClient{},
				Tracer:      tC.args.TracerOptions,
			}

			recorder := instana.NewTestRecorder()
			tracer := instana.NewTracerWithEverything(&opts, recorder)
			defer instana.ShutdownSensor()

			sp := tracer.StartSpan("test")
			sp.SetTag(string(ext.SpanKind), "entry")
			for _, field := range tC.args.fields {
				sp.LogFields(field)
			}
			sp.Finish()

			spans := recorder.GetQueuedSpans()
			assert.Equal(t, tC.expectedResult.spanCount, len(spans))

			span := spans[0]
			assert.Empty(t, span.ParentID)
			assert.Equal(t, tC.expectedResult.errorCount, span.Ec)

			require.IsType(t, instana.SDKSpanData{}, span.Data)
			data := span.Data.(instana.SDKSpanData)

			assert.Equal(t, "test", data.Tags.Name)
			assert.Equal(t, "entry", data.Tags.Type)

			validateLogSpan := func(span, logSpan instana.Span, logSpanMessage string) {
				assert.Equal(t, span.TraceID, logSpan.TraceID)
				assert.Equal(t, span.SpanID, logSpan.ParentID)
				assert.Equal(t, "log.go", logSpan.Name)

				require.IsType(t, instana.LogSpanData{}, logSpan.Data)
				data := logSpan.Data.(instana.LogSpanData)

				assert.Equal(t, logSpanMessage, data.Tags.Message)
			}

			for i := 1; i < tC.expectedResult.spanCount; i++ {
				logSpan := spans[i]
				expectedLog := fmt.Sprintf(`error.object: "error%d"`, i)
				validateLogSpan(span, logSpan, expectedLog)
			}
		})
	}
}
