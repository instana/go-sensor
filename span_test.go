// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2017

package instana_test

import (
	"errors"
	"os"
	"path/filepath"
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
	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		Service: TestServiceName,
		Tracer: instana.TracerOptions{
			CollectableHTTPHeaders: []string{"x-custom-header-1", "x-custom-header-2"},
		},
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	start := time.Now()
	sp := c.StartSpan("test")
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
	c := instana.InitCollector(&instana.Options{
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	parentSpan := c.StartSpan("parent")

	childSpan := c.StartSpan("child", ot.ChildOf(parentSpan.Context()))
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
	c := instana.InitCollector(&instana.Options{
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	sp := c.StartSpan(op)
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
	c := instana.InitCollector(&instana.Options{
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	sp := c.StartSpan(op)
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
	c := instana.InitCollector(&instana.Options{
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	sp := c.StartSpan("test")
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
	c := instana.InitCollector(&instana.Options{
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	sp := c.StartSpan("test")
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
	c := instana.InitCollector(&instana.Options{
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

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
				Message: `error.object: "simulated error" function: "ErrorFunc"`,
			},
		},
		"error log": {
			Fields: []log.Field{
				log.String("error.object", "simulated error"),
				log.String("function", "ErrorFunc"),
			},
			ExpectedErrorCount: 1,
			ExpectedTags: instana.LogSpanTags{
				Level:   "ERROR",
				Message: `error.object: "simulated error" function: "ErrorFunc"`,
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
			sp := c.StartSpan("test")
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
	c := instana.InitCollector(&instana.Options{
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	sp := c.StartSpan("test", instana.SuppressTracing())
	sp.Finish()

	assert.Empty(t, recorder.GetQueuedSpans())
}

func TestSpan_Suppressed_SetTag(t *testing.T) {
	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	sp := c.StartSpan("test")
	instana.SuppressTracing().Set(sp)
	sp.Finish()

	assert.Empty(t, recorder.GetQueuedSpans())
}

func Test_tracerS_SuppressTracing(t *testing.T) {
	opName := "my_operation"
	suppressTracingTag := "suppress_tracing"
	exitSpan := ext.SpanKindRPCClientEnum
	entrySpan := ext.SpanKindRPCServerEnum
	allowRootExitSpanEnv := "INSTANA_ALLOW_ROOT_EXIT_SPAN"

	c := instana.InitCollector(&instana.Options{
		AgentClient: alwaysReadyClient{},
	})
	defer instana.ShutdownCollector()
	parentSpan := c.StartSpan("parent-span")

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
		want      int
	}{
		{
			name:      "env_unset_suppress_false_spanType_exit",
			exportEnv: false,
			args: args{
				operationName: opName,
				opts: ot.StartSpanOptions{
					Tags: getSpanTags(exitSpan, false),
				},
			},
			want: 0,
		},
		{
			name:      "env_unset_suppress_true_spanType_exit",
			exportEnv: false,
			args: args{
				operationName: opName,
				opts: ot.StartSpanOptions{
					Tags: getSpanTags(exitSpan, true),
				},
			},
			want: 0,
		},
		{
			name:      "env_set_suppress_false_spanType_exit",
			exportEnv: true,
			args: args{
				operationName: opName,
				opts: ot.StartSpanOptions{
					Tags: getSpanTags(exitSpan, false),
				},
			},
			want: 1,
		},
		{
			name:      "env_set_suppress_true_spanType_exit",
			exportEnv: true,
			args: args{
				operationName: opName,
				opts: ot.StartSpanOptions{
					Tags: getSpanTags(exitSpan, true),
				},
			},
			want: 0,
		},
		{
			name:      "env_unset_suppress_false_spanType_entry",
			exportEnv: false,
			args: args{
				operationName: opName,
				opts: ot.StartSpanOptions{
					Tags: getSpanTags(entrySpan, false),
				},
			},
			want: 1,
		},
		{
			name:      "env_unset_suppress_true_spanType_entry",
			exportEnv: false,
			args: args{
				operationName: opName,
				opts: ot.StartSpanOptions{
					Tags: getSpanTags(entrySpan, true),
				},
			},
			want: 0,
		},
		{
			name:      "env_set_suppress_false_spanType_entry",
			exportEnv: true,
			args: args{
				operationName: opName,
				opts: ot.StartSpanOptions{
					Tags: getSpanTags(entrySpan, false),
				},
			},
			want: 1,
		},
		{
			name:      "env_set_suppress_true_spanType_entry",
			exportEnv: true,
			args: args{
				operationName: opName,
				opts: ot.StartSpanOptions{
					Tags: getSpanTags(entrySpan, true),
				},
			},
			want: 0,
		},
		{
			name:      "env_unset_suppress_false_spanType_ExitSpanButNotRoot",
			exportEnv: false,
			args: args{
				operationName: opName,
				opts: ot.StartSpanOptions{
					Tags: getSpanTags(exitSpan, false),
					References: []ot.SpanReference{
						ot.ChildOf(parentSpan.Context()),
					},
				},
			},
			want: 1,
		},
		{
			name:      "env_set_suppress_false_spanType_ExitSpanButNotRoot",
			exportEnv: true,
			args: args{
				operationName: opName,
				opts: ot.StartSpanOptions{
					Tags: getSpanTags(exitSpan, false),
					References: []ot.SpanReference{
						ot.ChildOf(parentSpan.Context()),
					},
				},
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.exportEnv {
				os.Setenv(allowRootExitSpanEnv, "1")

				defer func() {
					os.Unsetenv(allowRootExitSpanEnv)
				}()
			}

			recorder := instana.NewTestRecorder()
			c := instana.InitCollector(&instana.Options{
				AgentClient: alwaysReadyClient{},
				Recorder:    recorder,
			})
			defer instana.ShutdownCollector()

			sp := c.StartSpanWithOptions(tt.args.operationName, tt.args.opts)
			sp.Finish()
			assert.Equal(t, tt.want, len(recorder.GetQueuedSpans()))
		})
	}
}

// TestDisableSpanCategoryFromConfig tests that log spans are properly filtered when disabled via in-code configuration
func TestDisableSpanCategory(t *testing.T) {
	tests := []struct {
		name             string
		disabledCategory string
		spanOperations   []string
		expectedRecorded []bool
	}{
		{
			name:             "Disable logging category",
			disabledCategory: "logging",
			spanOperations:   []string{string(instana.LogSpanType), string(instana.HTTPServerSpanType), string(instana.KafkaSpanType)},
			expectedRecorded: []bool{false, true, true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := instana.NewTestRecorder()

			opts := &instana.Options{
				Tracer: instana.TracerOptions{
					DisableSpans: map[string]bool{
						tt.disabledCategory: true,
					},
				},
				Recorder:    recorder,
				AgentClient: &alwaysReadyClient{},
			}

			c := instana.InitCollector(opts)
			defer instana.ShutdownCollector()

			for _, operation := range tt.spanOperations {
				span := c.StartSpan(operation)
				span.Finish()
			}

			spans := recorder.GetQueuedSpans()

			assert.Equal(t, 2, len(spans),
				"Expected number of recorded spans doesn't match")

			// Create a map of recorded operations for easier lookup
			recordedOps := make(map[string]bool)
			for _, span := range spans {
				recordedOps[span.Name] = true
			}

			// Verify each span was recorded or not as expected
			for i, operation := range tt.spanOperations {
				if tt.expectedRecorded[i] {
					assert.True(t, recordedOps[operation],
						"Expected operation %s to be recorded but it wasn't", operation)
				} else {
					assert.False(t, recordedOps[operation],
						"Expected operation %s to be filtered out but it was recorded", operation)
				}
			}
		})
	}
}

// TestDisableSpanCategoryFromConfig tests that log spans are properly filtered when disabled via configuration file
func TestDisableSpanCategoryFromConfig(t *testing.T) {

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	yamlContent := `tracing:
  disable:
    - logging: true
`
	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	require.NoError(t, err, "Failed to create test config file")

	t.Setenv("INSTANA_CONFIG_PATH", configPath)

	recorder := instana.NewTestRecorder()

	opts := &instana.Options{
		Recorder:    recorder,
		AgentClient: alwaysReadyClient{},
	}

	c := instana.InitCollector(opts)
	defer instana.ShutdownCollector()

	operations := []string{
		string(instana.HTTPServerSpanType),
		string(instana.KafkaSpanType),
		string(instana.LogSpanType),
	}

	// Expected recording status for each operation above
	expectedRecorded := []bool{true, true, false}

	for _, operation := range operations {
		span := c.StartSpan(operation)
		span.Finish()
	}

	spans := recorder.GetQueuedSpans()

	// Verify that spans of the disabled categories were not recorded
	assert.Equal(t, 2, len(spans),
		"Expected number of recorded spans doesn't match")

	// Create a map of recorded operations for easier lookup
	recordedOps := make(map[string]bool)
	for _, span := range spans {
		recordedOps[span.Name] = true
	}

	// Verify each span was recorded or not as expected
	for i, operation := range operations {
		if expectedRecorded[i] {
			assert.True(t, recordedOps[operation],
				"Expected operation %s to be recorded but it wasn't", operation)
		} else {
			assert.False(t, recordedOps[operation],
				"Expected operation %s to be filtered out but it was recorded", operation)
		}
	}
}

// TestDisableAllCategories tests that log spans are filtered when logging category is disabled via environment variable
func TestDisableAllCategories(t *testing.T) {
	t.Setenv("INSTANA_TRACING_DISABLE", "logging")

	recorder := instana.NewTestRecorder()
	// Initialize tracer with all categories disabled (which now only includes logging)
	opts := &instana.Options{
		Tracer:      instana.TracerOptions{},
		Recorder:    recorder,
		AgentClient: alwaysReadyClient{},
	}

	tracer := instana.InitCollector(opts)
	defer instana.ShutdownCollector()

	operations := []string{
		string(instana.HTTPServerSpanType),
		string(instana.KafkaSpanType),
		string(instana.LogSpanType),
	}

	for _, operation := range operations {
		span := tracer.StartSpan(operation)
		span.Finish()
	}

	spans := recorder.GetQueuedSpans()

	// Verify that only non-logging spans were recorded
	assert.Equal(t, 2, len(spans), "Expected only non-logging spans to be recorded")

	// Create a map of recorded operations for easier lookup
	recordedOps := make(map[string]bool)
	for _, span := range spans {
		recordedOps[span.Name] = true
	}

	// Verify log spans were not recorded
	assert.False(t, recordedOps[string(instana.LogSpanType)],
		"Expected log spans to be filtered out but they were recorded")

	// Verify other spans were recorded
	assert.True(t, recordedOps[string(instana.HTTPServerSpanType)],
		"Expected HTTP spans to be recorded")
	assert.True(t, recordedOps[string(instana.KafkaSpanType)],
		"Expected Kafka spans to be recorded")
}
