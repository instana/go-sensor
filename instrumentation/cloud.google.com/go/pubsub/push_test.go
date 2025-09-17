// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package pubsub_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/cloud.google.com/go/pubsub"
	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTracingHandler(t *testing.T) {
	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	payload, err := os.ReadFile("testdata/message.json")
	require.NoError(t, err)

	var (
		numCalls int
		reqSpan  opentracing.Span
	)
	h := pubsub.TracingHandlerFunc(c, "/", func(w http.ResponseWriter, req *http.Request) {
		numCalls++

		var ok bool
		reqSpan, ok = instana.SpanFromContext(req.Context())
		require.True(t, ok)

		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		assert.JSONEq(t, string(payload), string(body))

		w.WriteHeader(http.StatusAccepted)
	})

	rec := httptest.NewRecorder()

	h(rec, httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(payload)))

	assert.Equal(t, http.StatusAccepted, rec.Result().StatusCode)
	assert.Equal(t, 1, numCalls)

	_ = reqSpan

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	gcpsSpan := spans[0]

	// new trace started
	assert.NotEmpty(t, gcpsSpan.TraceID)
	assert.Empty(t, gcpsSpan.ParentID)
	assert.NotEmpty(t, gcpsSpan.SpanID)

	// span tags
	assert.Equal(t, "gcps", gcpsSpan.Name)
	assert.EqualValues(t, instana.EntrySpanKind, gcpsSpan.Kind)
	assert.Equal(t, 0, gcpsSpan.Ec)

	require.IsType(t, instana.GCPPubSubSpanData{}, gcpsSpan.Data)

	data := gcpsSpan.Data.(instana.GCPPubSubSpanData)
	assert.Equal(t, instana.GCPPubSubSpanTags{
		Operation:    "CONSUME",
		ProjectID:    "myproject",
		Subscription: "mysubscription",
		MessageID:    "136969346945",
	}, data.Tags)
}

func TestTracingHandlerFunc_TracePropagation(t *testing.T) {
	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	payload, err := os.ReadFile("testdata/message_with_context.json")
	require.NoError(t, err)

	var numCalls int
	h := pubsub.TracingHandlerFunc(c, "/", func(w http.ResponseWriter, req *http.Request) {
		numCalls++

		_, ok := instana.SpanFromContext(req.Context())
		assert.True(t, ok)

		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		assert.JSONEq(t, string(payload), string(body))

		w.WriteHeader(http.StatusAccepted)
	})

	rec := httptest.NewRecorder()

	h(rec, httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(payload)))

	assert.Equal(t, http.StatusAccepted, rec.Result().StatusCode)
	assert.Equal(t, 1, numCalls)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	gcpsSpan := spans[0]

	// trace continuation
	assert.EqualValues(t, 0x1234, gcpsSpan.TraceID)
	assert.EqualValues(t, 0x5678, gcpsSpan.ParentID)
	assert.NotEmpty(t, gcpsSpan.SpanID)

	// span tags
	assert.Equal(t, "gcps", gcpsSpan.Name)
	assert.EqualValues(t, instana.EntrySpanKind, gcpsSpan.Kind)
	assert.Equal(t, 0, gcpsSpan.Ec)

	require.IsType(t, instana.GCPPubSubSpanData{}, gcpsSpan.Data)

	data := gcpsSpan.Data.(instana.GCPPubSubSpanData)
	assert.Equal(t, instana.GCPPubSubSpanTags{
		Operation:    "CONSUME",
		ProjectID:    "myproject",
		Subscription: "mysubscription",
		MessageID:    "136969346945",
	}, data.Tags)
}

func TestTracingHandlerFunc_NotPubSub(t *testing.T) {
	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	var numCalls int
	h := pubsub.TracingHandlerFunc(c, "/", func(w http.ResponseWriter, req *http.Request) {
		numCalls++

		_, ok := instana.SpanFromContext(req.Context())
		assert.True(t, ok)

		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		assert.Equal(t, "request payload", string(body))

		w.WriteHeader(http.StatusAccepted)
	})

	rec := httptest.NewRecorder()

	h(rec, httptest.NewRequest(http.MethodPost, "/", strings.NewReader("request payload")))

	assert.Equal(t, http.StatusAccepted, rec.Result().StatusCode)
	assert.Equal(t, 1, numCalls)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)
	assert.Equal(t, "g.http", spans[0].Name)
}

type alwaysReadyClient struct{}

func (alwaysReadyClient) Ready() bool                                       { return true }
func (alwaysReadyClient) SendMetrics(data acceptor.Metrics) error           { return nil }
func (alwaysReadyClient) SendEvent(event *instana.EventData) error          { return nil }
func (alwaysReadyClient) SendSpans(spans []instana.Span) error              { return nil }
func (alwaysReadyClient) SendProfiles(profiles []autoprofile.Profile) error { return nil }
func (alwaysReadyClient) Flush(context.Context) error                       { return nil }

// errorReader is a custom io.Reader that always returns an error
type errorReader struct{}

func (errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("forced read error")
}

// TestTracingHandlerFunc_ReadBodyError tests error handling when reading request body fails
func TestTracingHandlerFunc_ReadBodyError(t *testing.T) {
	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	var numCalls int
	h := pubsub.TracingHandlerFunc(c, "/", func(w http.ResponseWriter, req *http.Request) {
		numCalls++
	})

	rec := httptest.NewRecorder()

	// Create a request with a body that will fail to read
	req := httptest.NewRequest(http.MethodPost, "/", errorReader{})

	h(rec, req)

	// Verify that the handler returned an error and didn't call the wrapped handler
	assert.Equal(t, http.StatusInternalServerError, rec.Result().StatusCode)
	assert.Equal(t, 0, numCalls)
}

// TestTracingHandlerFunc_InvalidJSON tests handling of malformed JSON messages
func TestTracingHandlerFunc_InvalidJSON(t *testing.T) {
	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	var numCalls int
	h := pubsub.TracingHandlerFunc(c, "/", func(w http.ResponseWriter, req *http.Request) {
		numCalls++
	})

	rec := httptest.NewRecorder()

	// Create a request with invalid JSON
	invalidJSON := []byte(`{"message": "this is not valid pubsub message format"}`)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(invalidJSON))

	h(rec, req)

	// Verify that the handler falls back to regular HTTP tracing
	assert.Equal(t, http.StatusOK, rec.Result().StatusCode)
	assert.Equal(t, 1, numCalls)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)
	assert.Equal(t, "g.http", spans[0].Name) // Regular HTTP span, not PubSub span
}

// TestTracingHandlerFunc_InvalidSubscriptionFormat tests handling of invalid subscription formats
func TestTracingHandlerFunc_InvalidSubscriptionFormat(t *testing.T) {
	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	var numCalls int
	h := pubsub.TracingHandlerFunc(c, "/", func(w http.ResponseWriter, req *http.Request) {
		numCalls++
	})

	rec := httptest.NewRecorder()

	// Create a request with invalid subscription format
	invalidSubscription := []byte(`{
		"message": {
			"attributes": {},
			"messageId": "136969346945"
		},
		"subscription": "invalid-subscription-format"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(invalidSubscription))

	h(rec, req)

	// Verify that the handler falls back to regular HTTP tracing
	assert.Equal(t, http.StatusOK, rec.Result().StatusCode)
	assert.Equal(t, 1, numCalls)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)
	assert.Equal(t, "g.http", spans[0].Name) // Regular HTTP span, not PubSub span
}

// TestParsePathResourceIDExported tests the parsePathResourceID function indirectly
// by creating test cases that will exercise different paths in the subscription parsing logic
func TestParsePathResourceIDExported(t *testing.T) {
	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	tests := []struct {
		name          string
		subscription  string
		expectSuccess bool
	}{
		{
			name:          "Valid subscription format",
			subscription:  "projects/myproject/subscriptions/mysubscription",
			expectSuccess: true,
		},
		{
			name:          "Missing projects prefix",
			subscription:  "notprojects/myproject/subscriptions/mysubscription",
			expectSuccess: false,
		},
		{
			name:          "Missing subscriptions part",
			subscription:  "projects/myproject/notsubscriptions/mysubscription",
			expectSuccess: false,
		},
		{
			name:          "Empty string",
			subscription:  "",
			expectSuccess: false,
		},
		{
			name:          "Just projects prefix",
			subscription:  "projects/",
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We'll test the parsePathResourceID function indirectly through startConsumePushSpan
			payload := []byte(fmt.Sprintf(`{
				"message": {
					"attributes": {},
					"messageId": "136969346945"
				},
				"subscription": "%s"
			}`, tt.subscription))

			var numCalls int
			h := pubsub.TracingHandlerFunc(c, "/", func(w http.ResponseWriter, req *http.Request) {
				numCalls++
			})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(payload))

			h(rec, req)

			if tt.expectSuccess {
				// Should create a PubSub span
				assert.Equal(t, 1, numCalls)
				spans := recorder.GetQueuedSpans()
				require.Len(t, spans, 1)
				assert.Equal(t, "gcps", spans[0].Name)
			} else {
				// Should fall back to HTTP span
				assert.Equal(t, 1, numCalls)
				spans := recorder.GetQueuedSpans()
				require.Len(t, spans, 1)
				assert.Equal(t, "g.http", spans[0].Name)
			}

			// Create a new recorder for each test case instead of resetting
		})
	}
}
