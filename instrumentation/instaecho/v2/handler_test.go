// SPDX-FileCopyrightText: 2026 IBM Corp.
//
// SPDX-License-Identifier: MIT

package instaechov2_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"

	instana "github.com/instana/go-sensor"
	instaechov2 "github.com/instana/go-sensor/instrumentation/instaecho/v2"
	"github.com/labstack/echo/v5"
	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testTraceID                = "0000000000001234"
	testSpanID                 = "0000000000004567"
	testURLWithSensitiveParams = "https://example.com/foo/1?SECRET_VALUE=%3Credacted%3E&myPassword=%3Credacted%3E&q=term&sensitive_key=%3Credacted%3E"
	testURLSimple              = "https://example.com/foo?SECRET_VALUE=%3Credacted%3E&myPassword=%3Credacted%3E&q=term&sensitive_key=%3Credacted%3E"
)

func TestMain(m *testing.M) {
	instana.InitSensor(&instana.Options{
		Service: "echo-v2-test",
		Tracer: instana.TracerOptions{
			CollectableHTTPHeaders: []string{"x-custom-header-1", "x-custom-header-2"},
		},
		AgentClient: alwaysReadyClient{},
	})

	os.Exit(m.Run())
}

func TestPropagation(t *testing.T) {
	tests := []struct {
		name              string
		path              string
		routeName         string
		url               string
		handler           func(*testing.T) echo.HandlerFunc
		expectedSpanCount int
		expectedTags      instana.HTTPSpanTags
		validateLog       bool
		collectorOpts     *instana.Options
	}{
		{
			name:              "successful request with path template",
			path:              "/foo/:id",
			routeName:         "foos",
			url:               testURLWithSensitiveParams,
			handler:           createSuccessHandler,
			expectedSpanCount: 2,
			expectedTags: instana.HTTPSpanTags{
				Method:       "GET",
				Status:       http.StatusOK,
				Path:         "/foo/1",
				PathTemplate: "/foo/:id",
				URL:          "",
				Host:         "example.com",
				RouteID:      "foos",
				Protocol:     "https",
				Params:       "SECRET_VALUE=%3Credacted%3E&myPassword=%3Credacted%3E&q=term&sensitive_key=%3Credacted%3E",
				Headers: map[string]string{
					"x-custom-header-1": "request",
					"x-custom-header-2": "response",
				},
			},
			validateLog: false,
			collectorOpts: &instana.Options{
				AgentClient: alwaysReadyClient{},
			},
		},
		{
			name:              "request with error",
			path:              "/foo",
			routeName:         "foos",
			url:               testURLSimple,
			handler:           createErrorHandler,
			expectedSpanCount: 3,
			expectedTags: instana.HTTPSpanTags{
				Method:   "GET",
				Status:   http.StatusInternalServerError,
				Path:     "/foo",
				URL:      "",
				Host:     "example.com",
				RouteID:  "foos",
				Protocol: "https",
				Params:   "SECRET_VALUE=%3Credacted%3E&myPassword=%3Credacted%3E&q=term&sensitive_key=%3Credacted%3E",
				Headers: map[string]string{
					"x-custom-header-1": "request",
					"x-custom-header-2": "response",
				},
				Error: "Internal Server Error",
			},
			validateLog: true,
			collectorOpts: &instana.Options{
				Service: "test_service",
				Tracer: instana.TracerOptions{
					CollectableHTTPHeaders: []string{"x-custom-header-1", "x-custom-header-2"},
				},
				AgentClient: alwaysReadyClient{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := instana.NewTestRecorder()
			tt.collectorOpts.Recorder = recorder
			c := instana.InitCollector(tt.collectorOpts)
			t.Cleanup(instana.ShutdownCollector)

			engine := instaechov2.New(c)
			setupTestRoute(t, engine, tt.path, tt.routeName, tt.handler(t))

			req := setupTracedRequest("GET", tt.url, testTraceID, testSpanID)
			w := httptest.NewRecorder()

			engine.ServeHTTP(w, req)

			assertTraceHeaders(t, w)

			spans := recorder.GetQueuedSpans()
			require.Len(t, spans, tt.expectedSpanCount)

			entrySpan, interSpan := spans[1], spans[0]

			assertSpanRelationship(t, entrySpan, interSpan, testTraceID, testSpanID)
			assertHTTPSpanData(t, entrySpan, tt.expectedTags)

			// Validate log span if error case
			if tt.validateLog {
				logSpan := spans[2]

				// assert that log message has been recorded within the span interval
				assert.GreaterOrEqual(t, logSpan.Timestamp, entrySpan.Timestamp)
				assert.LessOrEqual(t, logSpan.Duration, entrySpan.Duration)

				require.IsType(t, instana.LogSpanData{}, logSpan.Data)
				logData := logSpan.Data.(instana.LogSpanData)

				assert.Equal(t, instana.LogSpanTags{
					Level:   "ERROR",
					Message: `error: "Internal Server Error"`,
				}, logData.Tags)
			}
		})
	}
}

// setupTracedRequest creates an HTTP request with Instana trace headers
func setupTracedRequest(method, url, traceID, spanID string) *http.Request {
	req := httptest.NewRequest(method, url, nil)
	req.Header.Add(instana.FieldT, traceID)
	req.Header.Add(instana.FieldS, spanID)
	req.Header.Add(instana.FieldL, "1")
	req.Header.Set("X-Custom-Header-1", "request")
	return req
}

// assertTraceHeaders verifies that all required trace headers are present in the response
func assertTraceHeaders(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	assert.NotEmpty(t, w.Header().Get("X-Instana-T"))
	assert.NotEmpty(t, w.Header().Get("X-Instana-S"))
	assert.NotEmpty(t, w.Header().Get("X-Instana-L"))
	assert.NotEmpty(t, w.Header().Get("Traceparent"))
	assert.NotEmpty(t, w.Header().Get("Tracestate"))
}

// assertSpanRelationship verifies the parent-child relationship between spans
func assertSpanRelationship(t *testing.T, entrySpan, childSpan instana.Span, expectedTraceID, expectedParentID string) {
	t.Helper()
	assert.EqualValues(t, instana.EntrySpanKind, entrySpan.Kind)
	assert.EqualValues(t, instana.IntermediateSpanKind, childSpan.Kind)
	assert.Equal(t, entrySpan.TraceID, childSpan.TraceID)
	assert.Equal(t, entrySpan.SpanID, childSpan.ParentID)
	assert.Equal(t, expectedTraceID, instana.FormatID(entrySpan.TraceID))
	assert.Equal(t, expectedParentID, instana.FormatID(entrySpan.ParentID))
}

// setupTestRoute registers a test route with the given handler
func setupTestRoute(t *testing.T, engine *echo.Echo, path, name string, handler echo.HandlerFunc) {
	t.Helper()
	_, err := engine.AddRoute(echo.Route{
		Method:  "GET",
		Path:    path,
		Name:    name,
		Handler: handler,
	})
	require.NoError(t, err)
}

// createSuccessHandler creates a handler that completes successfully with a child span
func createSuccessHandler(t *testing.T) echo.HandlerFunc {
	return func(c *echo.Context) error {
		parent, ok := instana.SpanFromContext(c.Request().Context())
		assert.True(t, ok)

		sp := parent.Tracer().StartSpan("sub-call", opentracing.ChildOf(parent.Context()))
		sp.Finish()

		c.Response().Header().Set("x-custom-header-2", "response")
		return c.JSON(200, []byte("{}"))
	}
}

// createErrorHandler creates a handler that returns an error with a child span
func createErrorHandler(t *testing.T) echo.HandlerFunc {
	return func(c *echo.Context) error {
		parent, ok := instana.SpanFromContext(c.Request().Context())
		assert.True(t, ok)

		sp := parent.Tracer().StartSpan("sub-call", opentracing.ChildOf(parent.Context()))
		sp.Finish()
		c.Response().Header().Set("x-custom-header-2", "response")

		return errors.New("MY-BAD-TEST-ERROR")
	}
}

// assertHTTPSpanData validates HTTP span data against expected tags
func assertHTTPSpanData(t *testing.T, span instana.Span, expected instana.HTTPSpanTags) {
	t.Helper()
	require.IsType(t, instana.HTTPSpanData{}, span.Data)
	spanData := span.Data.(instana.HTTPSpanData)
	assert.Equal(t, expected, spanData.Tags)
}

type alwaysReadyClient struct{}

func (alwaysReadyClient) Ready() bool                                       { return true }
func (alwaysReadyClient) SendMetrics(data acceptor.Metrics) error           { return nil }
func (alwaysReadyClient) SendEvent(event *instana.EventData) error          { return nil }
func (alwaysReadyClient) SendSpans(spans []instana.Span) error              { return nil }
func (alwaysReadyClient) SendProfiles(profiles []autoprofile.Profile) error { return nil }
func (alwaysReadyClient) Flush(context.Context) error                       { return nil }
