// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2017

package instana_test

import (
	"testing"

	instana "github.com/instana/go-sensor"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecorderBasics(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	defer instana.ShutdownSensor()

	span := tracer.StartSpan("http-client")
	span.SetTag(string(ext.SpanKind), "exit")
	span.SetTag("http.status", 200)
	span.SetTag("http.url", "https://www.instana.com/product/")
	span.SetTag(string(ext.HTTPMethod), "GET")
	span.Finish()

	// Validate GetQueuedSpans returns queued spans and clears the queue
	spans := recorder.GetQueuedSpans()
	assert.Len(t, spans, 1)
	assert.Equal(t, 0, recorder.QueuedSpansCount())
}

func TestRecorder_BatchSpan(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	defer instana.ShutdownSensor()

	tracer.StartSpan("test-span", instana.BatchSize(2)).Finish()

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	require.NotNil(t, spans[0].Batch)
	assert.Equal(t, 2, spans[0].Batch.Size)
}

func TestRecorder_BatchSpan_Single(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	defer instana.ShutdownSensor()

	tracer.StartSpan("test-span", instana.BatchSize(1)).Finish()

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	assert.Nil(t, spans[0].Batch)
}
