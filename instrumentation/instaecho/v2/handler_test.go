// SPDX-FileCopyrightText: 2026 IBM Corp.
// SPDX-FileCopyrightText: 2026 Instana Inc.
//
// SPDX-License-Identifier: MIT

//go:build go1.18
// +build go1.18

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
	traceIDHeader := "0000000000001234"
	spanIDHeader := "0000000000004567"

	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	engine := instaechov2.New(c)
	_, err := engine.AddRoute(echo.Route{
		Method: "GET",
		Path:   "/foo/:id",
		Name:   "foos",
		Handler: func(c *echo.Context) error {

			parent, ok := instana.SpanFromContext(c.Request().Context())
			assert.True(t, ok)

			sp := parent.Tracer().StartSpan("sub-call", opentracing.ChildOf(parent.Context()))
			sp.Finish()

			c.Response().Header().Set("x-custom-header-2", "response")
			return c.JSON(200, []byte("{}"))
		},
	})
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "https://example.com/foo/1?SECRET_VALUE=%3Credacted%3E&myPassword=%3Credacted%3E&q=term&sensitive_key=%3Credacted%3E", nil)

	req.Header.Add(instana.FieldT, traceIDHeader)
	req.Header.Add(instana.FieldS, spanIDHeader)
	req.Header.Add(instana.FieldL, "1")
	req.Header.Set("X-Custom-Header-1", "request")

	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	// Response headers assertions
	assert.NotEmpty(t, w.Header().Get("X-Instana-T"))
	assert.NotEmpty(t, w.Header().Get("X-Instana-S"))
	assert.NotEmpty(t, w.Header().Get("X-Instana-L"))
	assert.NotEmpty(t, w.Header().Get("Traceparent"))
	assert.NotEmpty(t, w.Header().Get("Tracestate"))

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 2)

	entrySpan, interSpan := spans[1], spans[0]

	assert.EqualValues(t, instana.EntrySpanKind, entrySpan.Kind)
	assert.EqualValues(t, instana.IntermediateSpanKind, interSpan.Kind)

	assert.Equal(t, entrySpan.TraceID, interSpan.TraceID)
	assert.Equal(t, entrySpan.SpanID, interSpan.ParentID)

	assert.Equal(t, traceIDHeader, instana.FormatID(entrySpan.TraceID))
	assert.Equal(t, spanIDHeader, instana.FormatID(entrySpan.ParentID))

	// ensure that entry span contains all necessary data
	require.IsType(t, instana.HTTPSpanData{}, entrySpan.Data)
	entrySpanData := entrySpan.Data.(instana.HTTPSpanData)

	assert.Equal(t, instana.HTTPSpanTags{
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
	}, entrySpanData.Tags)
}

func TestPropagationWithError(t *testing.T) {
	traceIDHeader := "0000000000001234"
	spanIDHeader := "0000000000004567"

	recorder := instana.NewTestRecorder()
	opts := &instana.Options{
		Service: "test_service",
		Tracer: instana.TracerOptions{
			CollectableHTTPHeaders: []string{"x-custom-header-1", "x-custom-header-2"},
		},
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	}
	c := instana.InitCollector(opts)
	defer instana.ShutdownCollector()

	engine := instaechov2.New(c)
	_, err := engine.AddRoute(echo.Route{
		Method: "GET",
		Path:   "/foo",
		Name:   "foos",
		Handler: func(c *echo.Context) error {
			parent, ok := instana.SpanFromContext(c.Request().Context())
			assert.True(t, ok)

			sp := parent.Tracer().StartSpan("sub-call", opentracing.ChildOf(parent.Context()))
			sp.Finish()
			c.Response().Header().Set("x-custom-header-2", "response")

			return errors.New("MY-BAD-TEST-ERROR")
		},
	})
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "https://example.com/foo?SECRET_VALUE=%3Credacted%3E&myPassword=%3Credacted%3E&q=term&sensitive_key=%3Credacted%3E", nil)

	req.Header.Add(instana.FieldT, traceIDHeader)
	req.Header.Add(instana.FieldS, spanIDHeader)
	req.Header.Add(instana.FieldL, "1")
	req.Header.Set("X-Custom-Header-1", "request")

	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	// Response headers assertions
	assert.NotEmpty(t, w.Header().Get("X-Instana-T"))
	assert.NotEmpty(t, w.Header().Get("X-Instana-S"))
	assert.NotEmpty(t, w.Header().Get("X-Instana-L"))
	assert.NotEmpty(t, w.Header().Get("Traceparent"))
	assert.NotEmpty(t, w.Header().Get("Tracestate"))

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 3)

	interSpan, entrySpan, logSpan := spans[0], spans[1], spans[2]

	assert.EqualValues(t, instana.EntrySpanKind, entrySpan.Kind)
	assert.EqualValues(t, instana.IntermediateSpanKind, interSpan.Kind)

	assert.Equal(t, entrySpan.TraceID, interSpan.TraceID)
	assert.Equal(t, entrySpan.SpanID, interSpan.ParentID)

	assert.Equal(t, traceIDHeader, instana.FormatID(entrySpan.TraceID))
	assert.Equal(t, spanIDHeader, instana.FormatID(entrySpan.ParentID))

	// ensure that entry span contains all necessary data
	require.IsType(t, instana.HTTPSpanData{}, entrySpan.Data)
	entrySpanData := entrySpan.Data.(instana.HTTPSpanData)

	assert.Equal(t, instana.HTTPSpanTags{
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
	}, entrySpanData.Tags)

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

type alwaysReadyClient struct{}

func (alwaysReadyClient) Ready() bool                                       { return true }
func (alwaysReadyClient) SendMetrics(data acceptor.Metrics) error           { return nil }
func (alwaysReadyClient) SendEvent(event *instana.EventData) error          { return nil }
func (alwaysReadyClient) SendSpans(spans []instana.Span) error              { return nil }
func (alwaysReadyClient) SendProfiles(profiles []autoprofile.Profile) error { return nil }
func (alwaysReadyClient) Flush(context.Context) error                       { return nil }
