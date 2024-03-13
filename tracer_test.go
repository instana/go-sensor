// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2017

package instana_test

import (
	"os"
	"testing"

	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
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

func Test_tracerS_SuppressTracing(t *testing.T) {
	opName := "my_operation"
	suppressTracingTag := "suppress_tracing"
	exitSpan := ext.SpanKindRPCClientEnum
	entrySpan := ext.SpanKindRPCServerEnum
	allowExitAsRoot := "INSTANA_ALLOW_EXIT_AS_ROOT"

	getSpanTags := func(kind ext.SpanKindEnum, suppressTracing bool) ot.Tags {
		return ot.Tags{
			"span.kind":        kind,
			suppressTracingTag: suppressTracing,
		}
	}

	type args struct {
		operationName string
		opts          ot.StartSpanOptions
	}
	tests := []struct {
		name      string
		exportEnv bool
		args      args
		want      bool
	}{
		{
			name:      "env_disable_tag_false_exit_true",
			exportEnv: false,
			args: args{
				operationName: opName,
				opts: ot.StartSpanOptions{
					Tags: getSpanTags(exitSpan, false),
				},
			},
			want: true,
		},
		{
			name:      "env_disable_tag_true_exit_true",
			exportEnv: false,
			args: args{
				operationName: opName,
				opts: ot.StartSpanOptions{
					Tags: getSpanTags(exitSpan, true),
				},
			},
			want: true,
		},
		{
			name:      "env_enable_tag_false_exit_true",
			exportEnv: true,
			args: args{
				operationName: opName,
				opts: ot.StartSpanOptions{
					Tags: getSpanTags(exitSpan, false),
				},
			},
			want: false,
		},
		{
			name:      "env_enable_tag_true_exit_true",
			exportEnv: true,
			args: args{
				operationName: opName,
				opts: ot.StartSpanOptions{
					Tags: getSpanTags(exitSpan, true),
				},
			},
			want: false,
		}, // ------------------
		{
			name:      "env_disable_tag_false_exit_false",
			exportEnv: false,
			args: args{
				operationName: opName,
				opts: ot.StartSpanOptions{
					Tags: getSpanTags(entrySpan, false),
				},
			},
			want: false,
		},
		{
			name:      "env_disable_tag_true_exit_false",
			exportEnv: false,
			args: args{
				operationName: opName,
				opts: ot.StartSpanOptions{
					Tags: getSpanTags(entrySpan, true),
				},
			},
			want: true,
		},
		{
			name:      "env_enable_tag_false_exit_false",
			exportEnv: true,
			args: args{
				operationName: opName,
				opts: ot.StartSpanOptions{
					Tags: getSpanTags(entrySpan, false),
				},
			},
			want: false,
		},
		{
			name:      "env_enable_tag_true_exit_false",
			exportEnv: true,
			args: args{
				operationName: opName,
				opts: ot.StartSpanOptions{
					Tags: getSpanTags(entrySpan, true),
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.exportEnv {
				os.Setenv(allowExitAsRoot, "1")
			} else {
				os.Unsetenv(allowExitAsRoot)
			}

			tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, nil)
			got := tracer.StartSpanWithOptions(tt.args.operationName, tt.args.opts)
			sc := got.Context().(instana.SpanContext)
			assert.Equal(t, tt.want, sc.Suppressed)
		})
	}
}
