// (c) Copyright IBM Corp. 2023

//go:build go1.17
// +build go1.17

package instafiber_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"
	"github.com/instana/go-sensor/instrumentation/instafiber"
	"github.com/instana/go-sensor/w3ctrace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

func BenchmarkTracingHandlerFunc(b *testing.B) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{
		Service:     "go-sensor-test",
		AgentClient: alwaysReadyClient{},
	}, recorder))
	defer instana.ShutdownSensor()

	h := instafiber.TraceHandler(s, "action", "/{action}", func(c *fiber.Ctx) error {
		return c.SendString("Ok")
	})

	app := fiber.New()
	app.Get("/test", h)

	req := httptest.NewRequest(http.MethodGet, "/test?q=term", nil)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = app.Test(req)
	}
}

func TestTracingHandlerFunc_Write(t *testing.T) {
	opts := &instana.Options{
		Service: "go-sensor-test",
		Tracer: instana.TracerOptions{
			CollectableHTTPHeaders: []string{"x-custom-header-1", "x-custom-header-2"},
		},
		AgentClient: alwaysReadyClient{},
	}

	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(opts, recorder))
	defer instana.ShutdownSensor()

	h := instafiber.TraceHandler(s, "action", "/{action}", func(c *fiber.Ctx) error {
		c.Set("X-Response", "true")
		c.Set("X-Custom-Header-2", "response")
		return c.SendString("Ok\n")
	})

	app := fiber.New()
	app.Get("/test", h)

	req := httptest.NewRequest(http.MethodGet, "/test?q=term", nil)
	req.Header.Set("Authorization", "Basic blah")
	req.Header.Set("X-Custom-Header-1", "request")

	resp, err := app.Test(req)
	assert.Equal(t, err, nil)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	b, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "Ok\n", string(b))

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	assert.Equal(t, 0, span.Ec)
	assert.EqualValues(t, instana.EntrySpanKind, span.Kind)
	assert.False(t, span.Synthetic)
	assert.Empty(t, span.CorrelationType)
	assert.Empty(t, span.CorrelationID)
	assert.False(t, span.ForeignTrace)
	assert.Empty(t, span.Ancestor)

	require.IsType(t, instana.HTTPSpanData{}, span.Data)
	data := span.Data.(instana.HTTPSpanData)

	assert.Equal(t, instana.HTTPSpanTags{
		Host:   "example.com",
		Status: http.StatusOK,
		Method: "GET",
		Path:   "/test",
		Params: "q=term",
		Headers: map[string]string{
			"x-custom-header-1": "request",
			"x-custom-header-2": "response",
		},
		PathTemplate: "/{action}",
		RouteID:      "action",
		Protocol:     "http",
	}, data.Tags)

	// check whether the trace context has been sent back to the client
	assert.Equal(t, instana.FormatID(span.TraceID), resp.Header.Get(instana.FieldT))
	assert.Equal(t, instana.FormatID(span.SpanID), resp.Header.Get(instana.FieldS))

	// w3c trace context
	traceparent := resp.Header.Get(w3ctrace.TraceParentHeader)
	assert.Contains(t, traceparent, instana.FormatLongID(span.TraceIDHi, span.TraceID))
	assert.Contains(t, traceparent, instana.FormatID(span.SpanID))

	tracestate := resp.Header.Get(w3ctrace.TraceStateHeader)
	assert.True(t, strings.HasPrefix(
		tracestate,
		"in="+instana.FormatID(span.TraceID)+";"+instana.FormatID(span.SpanID),
	), tracestate)
}

func TestTracingHandlerFunc_InstanaFieldLPriorityOverTraceParentHeader(t *testing.T) {
	type testCase struct {
		headers                 http.Header
		traceParentHeaderSuffix string
	}

	testCases := map[string]testCase{
		"traceparent is suppressed, x-instana-l is not suppressed": {
			headers: http.Header{
				w3ctrace.TraceParentHeader: []string{"00-00000000000000000000000000000001-0000000000000001-00"},
				instana.FieldL:             []string{"1"},
			},
			traceParentHeaderSuffix: "-01",
		},
		"traceparent is suppressed, x-instana-l is absent (is not suppressed by default)": {
			headers: http.Header{
				w3ctrace.TraceParentHeader: []string{"00-00000000000000000000000000000001-0000000000000001-00"},
			},
			traceParentHeaderSuffix: "-01",
		},
		"traceparent is not suppressed, x-instana-l is absent (tracing enabled by default)": {
			headers: http.Header{
				w3ctrace.TraceParentHeader: []string{"00-00000000000000000000000000000001-0000000000000001-01"},
			},
			traceParentHeaderSuffix: "-01",
		},
		"traceparent is not suppressed, x-instana-l is not suppressed": {
			headers: http.Header{
				w3ctrace.TraceParentHeader: []string{"00-00000000000000000000000000000001-0000000000000001-01"},
				instana.FieldL:             []string{"1"},
			},
			traceParentHeaderSuffix: "-01",
		},
		"traceparent is suppressed, x-instana-l is suppressed": {
			headers: http.Header{
				w3ctrace.TraceParentHeader: []string{"00-00000000000000000000000000000001-0000000000000001-00"},
				instana.FieldL:             []string{"0"},
			},
			traceParentHeaderSuffix: "-00",
		},
		"traceparent is not suppressed, x-instana-l is suppressed": {
			headers: http.Header{
				w3ctrace.TraceParentHeader: []string{"00-00000000000000000000000000000001-0000000000000001-01"},
				instana.FieldL:             []string{"0"},
			},
			traceParentHeaderSuffix: "-00",
		},
	}

	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{
		Service:     "go-sensor-test",
		AgentClient: alwaysReadyClient{},
	}, recorder))
	defer instana.ShutdownSensor()

	h := instafiber.TraceHandler(s, "action", "/test", func(c *fiber.Ctx) error { return nil })

	app := fiber.New()
	app.Get("/test", h)

	for name, testCase := range testCases {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header = testCase.headers

		resp, err := app.Test(req)
		assert.Equal(t, err, nil)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, strings.HasSuffix(resp.Header.Get(w3ctrace.TraceParentHeader), testCase.traceParentHeaderSuffix), "case '"+name+"' failed")
	}
}

func TestTracingHandlerFunc_WriteHeaders(t *testing.T) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder))
	defer instana.ShutdownSensor()

	h := instafiber.TraceHandler(s, "test", "/test", func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusNotFound)
	})

	req := httptest.NewRequest(http.MethodGet, "/test?q=term", nil)

	app := fiber.New()
	app.Get("/test", h)

	resp, err := app.Test(req)
	assert.Equal(t, err, nil)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	assert.Equal(t, 0, span.Ec)
	assert.EqualValues(t, instana.EntrySpanKind, span.Kind)
	assert.False(t, span.Synthetic)
	assert.Empty(t, span.CorrelationType)
	assert.Empty(t, span.CorrelationID)
	assert.False(t, span.ForeignTrace)
	assert.Empty(t, span.Ancestor)

	require.IsType(t, instana.HTTPSpanData{}, span.Data)
	data := span.Data.(instana.HTTPSpanData)

	assert.Equal(t, instana.HTTPSpanTags{
		Status:   http.StatusNotFound,
		Method:   "GET",
		Host:     "example.com",
		Path:     "/test",
		Params:   "q=term",
		RouteID:  "test",
		Protocol: "http",
	}, data.Tags)

	// check whether the trace context has been sent back to the client
	assert.Equal(t, instana.FormatID(span.TraceID), resp.Header.Get(instana.FieldT))
	assert.Equal(t, instana.FormatID(span.SpanID), resp.Header.Get(instana.FieldS))

	// w3c trace context
	traceparent := resp.Header.Get(w3ctrace.TraceParentHeader)
	assert.Contains(t, traceparent, instana.FormatLongID(span.TraceIDHi, span.TraceID))
	assert.Contains(t, traceparent, instana.FormatID(span.SpanID))

	tracestate := resp.Header.Get(w3ctrace.TraceStateHeader)
	assert.True(t, strings.HasPrefix(
		tracestate,
		"in="+instana.FormatID(span.TraceID)+";"+instana.FormatID(span.SpanID),
	), tracestate)
}

func TestTracingHandlerFunc_W3CTraceContext(t *testing.T) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder))
	defer instana.ShutdownSensor()

	h := instafiber.TraceHandler(s, "test", "/test", func(c *fiber.Ctx) error {
		return c.SendString("Ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(w3ctrace.TraceParentHeader, "00-00000000000000010000000000000002-0000000000000003-01")
	req.Header.Set(w3ctrace.TraceStateHeader, "in=1234;5678,rojo=00f067aa0ba902b7")

	app := fiber.New()
	app.Get("/test", h)

	resp, err := app.Test(req)
	assert.Equal(t, err, nil)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span := spans[0]

	assert.EqualValues(t, 0x1, span.TraceIDHi)
	assert.EqualValues(t, 0x2, span.TraceID)
	assert.EqualValues(t, 0x3, span.ParentID)

	assert.Equal(t, 0, span.Ec)
	assert.EqualValues(t, instana.EntrySpanKind, span.Kind)
	assert.False(t, span.Synthetic)
	assert.Empty(t, span.CorrelationType)
	assert.Empty(t, span.CorrelationID)
	assert.True(t, span.ForeignTrace)
	assert.Equal(t, &instana.TraceReference{
		TraceID:  "1234",
		ParentID: "5678",
	}, span.Ancestor)

	require.IsType(t, instana.HTTPSpanData{}, span.Data)
	data := span.Data.(instana.HTTPSpanData)

	assert.Equal(t, instana.HTTPSpanTags{
		Host:     "example.com",
		Status:   http.StatusOK,
		Method:   "GET",
		Path:     "/test",
		RouteID:  "test",
		Protocol: "http",
	}, data.Tags)

	// check whether the trace context has been sent back to the client
	assert.Equal(t, instana.FormatID(span.TraceID), resp.Header.Get(instana.FieldT))
	assert.Equal(t, instana.FormatID(span.SpanID), resp.Header.Get(instana.FieldS))

	// w3c trace context
	traceparent := resp.Header.Get(w3ctrace.TraceParentHeader)
	assert.Contains(t, traceparent, instana.FormatLongID(span.TraceIDHi, span.TraceID))
	assert.Contains(t, traceparent, instana.FormatID(span.SpanID))

	tracestate := resp.Header.Get(w3ctrace.TraceStateHeader)
	assert.True(t, strings.HasPrefix(
		tracestate,
		"in="+instana.FormatID(span.TraceID)+";"+instana.FormatID(span.SpanID),
	), tracestate)
}

func TestTracingHandlerFunc_SecretsFiltering(t *testing.T) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{
		Service:     "go-sensor-test",
		AgentClient: alwaysReadyClient{},
	}, recorder))
	defer instana.ShutdownSensor()

	h := instafiber.TraceHandler(s, "action", "/{action}", func(c *fiber.Ctx) error {
		return c.SendString("Ok\n")
	})

	req := httptest.NewRequest(http.MethodGet, "/test?q=term&sensitive_key=s3cr3t&myPassword=qwerty&SECRET_VALUE=1", nil)

	app := fiber.New()
	app.Get("/test", h)

	resp, err := app.Test(req)
	assert.Equal(t, err, nil)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	b, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "Ok\n", string(b))

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	assert.Equal(t, 0, span.Ec)
	assert.EqualValues(t, instana.EntrySpanKind, span.Kind)
	assert.False(t, span.Synthetic)
	assert.Empty(t, span.CorrelationType)
	assert.Empty(t, span.CorrelationID)

	require.IsType(t, instana.HTTPSpanData{}, span.Data)
	data := span.Data.(instana.HTTPSpanData)

	assert.Equal(t, instana.HTTPSpanTags{
		Host:         "example.com",
		Status:       http.StatusOK,
		Method:       "GET",
		Path:         "/test",
		Params:       "SECRET_VALUE=%3Credacted%3E&myPassword=%3Credacted%3E&q=term&sensitive_key=%3Credacted%3E",
		PathTemplate: "/{action}",
		RouteID:      "action",
		Protocol:     "http",
	}, data.Tags)

	// check whether the trace context has been sent back to the client
	assert.Equal(t, instana.FormatID(span.TraceID), resp.Header.Get(instana.FieldT))
	assert.Equal(t, instana.FormatID(span.SpanID), resp.Header.Get(instana.FieldS))
}

func TestTracingHandlerFunc_Error(t *testing.T) {
	// Create a sensor for tracing
	recorder := instana.NewTestRecorder()
	s := instana.InitCollector(&instana.Options{AgentClient: alwaysReadyClient{}, Recorder: recorder})
	defer instana.ShutdownSensor()

	// Wrap the request handler with instrumentation code
	h := instafiber.TraceHandler(s, "test", "/test", func(c *fiber.Ctx) error {
		c.Status(http.StatusInternalServerError)
		return fiber.NewError(http.StatusInternalServerError, "something went wrong")
	})

	// Create a sample request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	// Create a fiber app
	app := fiber.New()
	app.Get("/test", h)

	// Handle the request
	resp, err := app.Test(req)
	assert.Equal(t, err, nil, "error occurred while fetching the response")

	// Verify the response
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	// Fetch and verify the recorded spans
	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 2)

	span, logSpan := spans[0], spans[1]
	assert.Equal(t, 1, span.Ec)
	assert.EqualValues(t, instana.EntrySpanKind, span.Kind)
	assert.False(t, span.Synthetic)

	require.IsType(t, instana.HTTPSpanData{}, span.Data)
	data := span.Data.(instana.HTTPSpanData)

	assert.Equal(t, instana.HTTPSpanTags{
		Status:   http.StatusInternalServerError,
		Method:   "GET",
		Host:     "example.com",
		Path:     "/test",
		RouteID:  "test",
		Error:    "Internal Server Error",
		Protocol: "http",
	}, data.Tags)

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
		Message: `error: "Internal Server Error"`,
	}, logData.Tags)
}

func TestTracingHandlerFunc_SyntheticCall(t *testing.T) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder))
	defer instana.ShutdownSensor()

	h := instafiber.TraceHandler(s, "test-handler", "/", func(c *fiber.Ctx) error {
		return c.SendString("Ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(instana.FieldSynthetic, "1")

	app := fiber.New()
	app.Get("/test", h)

	resp, err := app.Test(req)

	assert.Equal(t, err, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)
	assert.True(t, spans[0].Synthetic)
}

func TestTracingHandlerFunc_EUMCall(t *testing.T) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder))
	defer instana.ShutdownSensor()

	h := instafiber.TraceHandler(s, "test-handler", "/", func(c *fiber.Ctx) error {
		return c.SendString("Ok")
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(instana.FieldL, "1,correlationType=web;correlationId=eum correlation id")

	app := fiber.New()
	app.Get("/test", h)

	resp, err := app.Test(req)
	assert.Equal(t, err, nil)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)
	assert.Equal(t, "web", spans[0].CorrelationType)
	assert.Equal(t, "eum correlation id", spans[0].CorrelationID)
}

func TestTracingHandlerFunc_PanicHandling(t *testing.T) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder))
	defer instana.ShutdownSensor()

	h := instafiber.TraceHandler(s, "test", "/test", func(c *fiber.Ctx) error {
		panic("something went wrong")
	})

	c := &fasthttp.RequestCtx{}

	c.Request.Header.SetMethod(fiber.MethodGet)
	c.Request.Header.Set(instana.FieldL, "1,correlationType=web;correlationId=eum correlation id")
	c.URI().SetPath("/test")
	c.URI().SetQueryString("q=term")
	c.URI().SetHost("example.com")

	app := fiber.New()
	app.Get("/test", h)

	assert.Panics(t, func() {
		handler := app.Handler()
		handler(c)
	})

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 2)

	span, logSpan := spans[0], spans[1]
	assert.Equal(t, 1, span.Ec)
	assert.EqualValues(t, instana.EntrySpanKind, span.Kind)
	assert.False(t, span.Synthetic)

	require.IsType(t, instana.HTTPSpanData{}, span.Data)
	data := span.Data.(instana.HTTPSpanData)

	assert.Equal(t, instana.HTTPSpanTags{
		Status:   http.StatusInternalServerError,
		Method:   "GET",
		Host:     "example.com",
		Path:     "/test",
		Params:   "q=term",
		RouteID:  "test",
		Error:    "something went wrong",
		Protocol: "http",
	}, data.Tags)

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
		Message: `error: "something went wrong"`,
	}, logData.Tags)
}

type alwaysReadyClient struct{}

func (alwaysReadyClient) Ready() bool                              { return true }
func (alwaysReadyClient) SendMetrics(acceptor.Metrics) error       { return nil }
func (alwaysReadyClient) SendEvent(*instana.EventData) error       { return nil }
func (alwaysReadyClient) SendSpans([]instana.Span) error           { return nil }
func (alwaysReadyClient) SendProfiles([]autoprofile.Profile) error { return nil }
func (alwaysReadyClient) Flush(context.Context) error              { return nil }
