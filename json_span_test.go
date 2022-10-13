// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instana_test

import (
	"testing"

	instana "github.com/instana/go-sensor"
	"github.com/instana/testify/assert"
	"github.com/instana/testify/require"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

func TestSpanKind_String(t *testing.T) {
	examples := map[string]struct {
		Kind     instana.SpanKind
		Expected string
	}{
		"entry": {
			Kind:     instana.EntrySpanKind,
			Expected: "entry",
		},
		"exit": {
			Kind:     instana.ExitSpanKind,
			Expected: "exit",
		},
		"intermediate": {
			Kind:     instana.IntermediateSpanKind,
			Expected: "intermediate",
		},
		"unknown": {
			Kind:     instana.SpanKind(0),
			Expected: "intermediate",
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, example.Expected, example.Kind.String())
		})
	}
}

func TestNewSDKSpanData(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)
	defer instana.TestOnlyStopSensor()

	sp := tracer.StartSpan("sdk",
		ext.SpanKindRPCServer,
		opentracing.Tags{
			"host":       "localhost",
			"custom.tag": "42",
		})
	sp.Finish()

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	require.IsType(t, instana.SDKSpanData{}, span.Data)

	data := span.Data.(instana.SDKSpanData)
	assert.Equal(t, instana.SDKSpanTags{
		Name: "sdk",
		Type: "entry",
		Custom: map[string]interface{}{
			"tags": opentracing.Tags{
				"span.kind":  ext.SpanKindRPCServerEnum,
				"host":       "localhost",
				"custom.tag": "42",
			},
		},
	}, data.Tags)
}

func TestSpanData_CustomTags(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)
	defer instana.TestOnlyStopSensor()

	sp := tracer.StartSpan("g.http", opentracing.Tags{
		"http.host":   "localhost",
		"http.path":   "/",
		"custom.tag":  "42",
		"another.tag": true,
	})
	sp.Finish()

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	require.IsType(t, instana.HTTPSpanData{}, span.Data)

	data := span.Data.(instana.HTTPSpanData)

	assert.Equal(t, instana.HTTPSpanTags{
		Host: "localhost",
		Path: "/",
	}, data.Tags)
	assert.Equal(t, &instana.CustomSpanData{
		Tags: map[string]interface{}{
			"custom.tag":  "42",
			"another.tag": true,
		},
	}, data.Custom)
}
