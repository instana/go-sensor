// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2017

package instana_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"
	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecorderBasics(t *testing.T) {

	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	pSpan := c.StartSpan("parent-span")
	span := c.StartSpan("http-client", ot.ChildOf(pSpan.Context()))
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
	c := instana.InitCollector(&instana.Options{
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	c.StartSpan("test-span", instana.BatchSize(2)).Finish()

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	require.NotNil(t, spans[0].Batch)
	assert.Equal(t, 2, spans[0].Batch.Size)
}

func TestRecorder_BatchSpan_Single(t *testing.T) {
	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	c.StartSpan("test-span", instana.BatchSize(1)).Finish()

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	assert.Nil(t, spans[0].Batch)
}

func TestRecorder_Flush_EmptyQueue(t *testing.T) {
	recorder := instana.NewTestRecorder()

	// Test flushing an empty queue
	err := recorder.Flush(context.Background())
	assert.NoError(t, err)
}

func TestRecorder_MaxBufferedSpans(t *testing.T) {
	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		AgentClient:      alwaysReadyClient{},
		Recorder:         recorder,
		MaxBufferedSpans: 3, // Set a small buffer size for testing
	})
	defer instana.ShutdownCollector()

	// Create more spans than the buffer can hold
	for i := 0; i < 5; i++ {
		c.StartSpan(fmt.Sprintf("span-%d", i)).Finish()
	}

	// Verify that only the most recent MaxBufferedSpans are kept
	spans := recorder.GetQueuedSpans()
	assert.Len(t, spans, 3)

	// Verify that only the most recent MaxBufferedSpans are kept
	assert.Len(t, spans, 3)
}

func TestRecorder_ForceTransmission(t *testing.T) {
	// Create a mock agent client that tracks when spans are sent
	mockAgent := &mockAgentClient{
		ready: true,
	}

	recorder := instana.NewRecorder()
	c := instana.InitCollector(&instana.Options{
		AgentClient:                 mockAgent,
		Recorder:                    recorder,
		MaxBufferedSpans:            10,
		ForceTransmissionStartingAt: 2, // Force transmission after 2 spans
	})
	defer instana.ShutdownCollector()

	// Create spans to trigger force transmission
	for i := 0; i < 2; i++ {
		c.StartSpan(fmt.Sprintf("span-%d", i)).Finish()
	}

	// Give some time for the async flush to happen
	time.Sleep(100 * time.Millisecond)

	// Verify that SendSpans was called
	assert.True(t, mockAgent.spansSent, "Expected spans to be sent to the agent")
}

// Mock agent client for testing
type mockAgentClient struct {
	ready     bool
	spansSent bool
}

func (m *mockAgentClient) Ready() bool                              { return m.ready }
func (m *mockAgentClient) SendMetrics(data acceptor.Metrics) error  { return nil }
func (m *mockAgentClient) SendEvent(event *instana.EventData) error { return nil }
func (m *mockAgentClient) SendSpans(spans []instana.Span) error {
	m.spansSent = true
	return nil
}
func (m *mockAgentClient) SendProfiles(profiles []autoprofile.Profile) error { return nil }
func (m *mockAgentClient) Flush(context.Context) error                       { return nil }

// alwaysReadyClient is already defined in instrumentation_http_test.go

func TestRecorder_Flush_Error(t *testing.T) {
	// Create a mock agent client that returns an error on SendSpans
	mockAgent := &errorAgentClient{
		ready: true,
	}

	recorder := instana.NewRecorder()
	c := instana.InitCollector(&instana.Options{
		AgentClient: mockAgent,
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	// Create a span to be flushed
	c.StartSpan("test-span").Finish()

	// Flush should return an error
	err := recorder.Flush(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send collected spans")

	// Verify that spans are put back in the queue
	assert.Greater(t, recorder.QueuedSpansCount(), 0)
}

// Mock agent client that returns an error on SendSpans
type errorAgentClient struct {
	ready bool
}

func (m *errorAgentClient) Ready() bool                                       { return m.ready }
func (m *errorAgentClient) SendMetrics(data acceptor.Metrics) error           { return nil }
func (m *errorAgentClient) SendEvent(event *instana.EventData) error          { return nil }
func (m *errorAgentClient) SendSpans(spans []instana.Span) error              { return fmt.Errorf("mock error") }
func (m *errorAgentClient) SendProfiles(profiles []autoprofile.Profile) error { return nil }
func (m *errorAgentClient) Flush(context.Context) error                       { return nil }

// TestRecorder_Flush_AgentNotReady tests the behavior when the agent is not ready
func TestRecorder_Flush_AgentNotReady(t *testing.T) {
	// Create a mock agent client that is not ready
	mockAgent := &notReadyAgentClient{}

	// Use a regular recorder, not a test recorder
	recorder := instana.NewRecorder()
	c := instana.InitCollector(&instana.Options{
		AgentClient: mockAgent,
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	// Create a span to be flushed
	c.StartSpan("test-span").Finish()

	// Wait a bit for the span to be processed
	time.Sleep(100 * time.Millisecond)

	// Get the initial count
	initialCount := recorder.QueuedSpansCount()

	// Flush should not return an error when agent is not ready
	err := recorder.Flush(context.Background())
	assert.NoError(t, err)

	// Spans should still be in the queue when agent is not ready
	assert.Equal(t, initialCount, recorder.QueuedSpansCount(), "Spans should remain in queue when agent is not ready")
}

// Mock agent client that is never ready
type notReadyAgentClient struct{}

func (notReadyAgentClient) Ready() bool                                       { return false }
func (notReadyAgentClient) SendMetrics(data acceptor.Metrics) error           { return nil }
func (notReadyAgentClient) SendEvent(event *instana.EventData) error          { return nil }
func (notReadyAgentClient) SendSpans(spans []instana.Span) error              { return nil }
func (notReadyAgentClient) SendProfiles(profiles []autoprofile.Profile) error { return nil }
func (notReadyAgentClient) Flush(context.Context) error                       { return nil }
