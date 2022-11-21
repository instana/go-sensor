// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instana_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/w3ctrace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func BenchmarkTracingNamedHandlerFunc(b *testing.B) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{
		Service:     "go-sensor-test",
		AgentClient: alwaysReadyClient{},
	}, recorder))
	defer instana.ShutdownSensor()

	h := instana.TracingNamedHandlerFunc(s, "action", "/{action}", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(w, "Ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test?q=term", nil)

	rec := httptest.NewRecorder()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		h.ServeHTTP(rec, req)
	}
}

func TestTracingNamedHandlerFunc_Write(t *testing.T) {
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

	h := instana.TracingNamedHandlerFunc(s, "action", "/{action}", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("X-Response", "true")
		w.Header().Set("X-Custom-Header-2", "response")
		fmt.Fprintln(w, "Ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test?q=term", nil)
	req.Header.Set("Authorization", "Basic blah")
	req.Header.Set("X-Custom-Header-1", "request")

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "Ok\n", rec.Body.String())

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
	}, data.Tags)

	// check whether the trace context has been sent back to the client
	assert.Equal(t, instana.FormatID(span.TraceID), rec.Header().Get(instana.FieldT))
	assert.Equal(t, instana.FormatID(span.SpanID), rec.Header().Get(instana.FieldS))

	// w3c trace context
	traceparent := rec.Header().Get(w3ctrace.TraceParentHeader)
	assert.Contains(t, traceparent, instana.FormatLongID(span.TraceIDHi, span.TraceID))
	assert.Contains(t, traceparent, instana.FormatID(span.SpanID))

	tracestate := rec.Header().Get(w3ctrace.TraceStateHeader)
	assert.True(t, strings.HasPrefix(
		tracestate,
		"in="+instana.FormatID(span.TraceID)+";"+instana.FormatID(span.SpanID),
	), tracestate)
}

func TestTracingNamedHandlerFunc_InstanaFieldLPriorityOverTraceParentHeader(t *testing.T) {
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

	h := instana.TracingNamedHandlerFunc(s, "action", "/test", func(w http.ResponseWriter, req *http.Request) {})

	for name, testCase := range testCases {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header = testCase.headers

		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, strings.HasSuffix(rec.Header().Get(w3ctrace.TraceParentHeader), testCase.traceParentHeaderSuffix), "case '"+name+"' failed")
	}
}

func TestTracingNamedHandlerFunc_WriteHeaders(t *testing.T) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder))
	defer instana.ShutdownSensor()

	h := instana.TracingNamedHandlerFunc(s, "test", "/test", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/test?q=term", nil))

	assert.Equal(t, http.StatusNotFound, rec.Code)

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
		Status:  http.StatusNotFound,
		Method:  "GET",
		Host:    "example.com",
		Path:    "/test",
		Params:  "q=term",
		RouteID: "test",
	}, data.Tags)

	// check whether the trace context has been sent back to the client
	assert.Equal(t, instana.FormatID(span.TraceID), rec.Header().Get(instana.FieldT))
	assert.Equal(t, instana.FormatID(span.SpanID), rec.Header().Get(instana.FieldS))

	// w3c trace context
	traceparent := rec.Header().Get(w3ctrace.TraceParentHeader)
	assert.Contains(t, traceparent, instana.FormatLongID(span.TraceIDHi, span.TraceID))
	assert.Contains(t, traceparent, instana.FormatID(span.SpanID))

	tracestate := rec.Header().Get(w3ctrace.TraceStateHeader)
	assert.True(t, strings.HasPrefix(
		tracestate,
		"in="+instana.FormatID(span.TraceID)+";"+instana.FormatID(span.SpanID),
	), tracestate)
}

func TestTracingNamedHandlerFunc_W3CTraceContext(t *testing.T) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder))
	defer instana.ShutdownSensor()

	h := instana.TracingNamedHandlerFunc(s, "test", "/test", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(w, "Ok")
	})

	rec := httptest.NewRecorder()

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(w3ctrace.TraceParentHeader, "00-00000000000000010000000000000002-0000000000000003-01")
	req.Header.Set(w3ctrace.TraceStateHeader, "in=1234;5678,rojo=00f067aa0ba902b7")

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

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
		Host:    "example.com",
		Status:  http.StatusOK,
		Method:  "GET",
		Path:    "/test",
		RouteID: "test",
	}, data.Tags)

	// check whether the trace context has been sent back to the client
	assert.Equal(t, instana.FormatID(span.TraceID), rec.Header().Get(instana.FieldT))
	assert.Equal(t, instana.FormatID(span.SpanID), rec.Header().Get(instana.FieldS))

	// w3c trace context
	traceparent := rec.Header().Get(w3ctrace.TraceParentHeader)
	assert.Contains(t, traceparent, instana.FormatLongID(span.TraceIDHi, span.TraceID))
	assert.Contains(t, traceparent, instana.FormatID(span.SpanID))

	tracestate := rec.Header().Get(w3ctrace.TraceStateHeader)
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

	h := instana.TracingNamedHandlerFunc(s, "action", "/{action}", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(w, "Ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test?q=term&sensitive_key=s3cr3t&myPassword=qwerty&SECRET_VALUE=1", nil)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "Ok\n", rec.Body.String())

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
	}, data.Tags)

	// check whether the trace context has been sent back to the client
	assert.Equal(t, instana.FormatID(span.TraceID), rec.Header().Get(instana.FieldT))
	assert.Equal(t, instana.FormatID(span.SpanID), rec.Header().Get(instana.FieldS))
}

func TestTracingHandlerFunc_Error(t *testing.T) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder))
	defer instana.ShutdownSensor()

	h := instana.TracingNamedHandlerFunc(s, "test", "/test", func(w http.ResponseWriter, req *http.Request) {
		http.Error(w, "something went wrong", http.StatusInternalServerError)
	})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/test", nil))

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 2)

	span, logSpan := spans[0], spans[1]
	assert.Equal(t, 1, span.Ec)
	assert.EqualValues(t, instana.EntrySpanKind, span.Kind)
	assert.False(t, span.Synthetic)

	require.IsType(t, instana.HTTPSpanData{}, span.Data)
	data := span.Data.(instana.HTTPSpanData)

	assert.Equal(t, instana.HTTPSpanTags{
		Status:  http.StatusInternalServerError,
		Method:  "GET",
		Host:    "example.com",
		Path:    "/test",
		RouteID: "test",
		Error:   "Internal Server Error",
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

	h := instana.TracingNamedHandlerFunc(s, "test-handler", "/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(w, "Ok")
	})

	rec := httptest.NewRecorder()

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(instana.FieldSynthetic, "1")

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)
	assert.True(t, spans[0].Synthetic)
}

func TestTracingHandlerFunc_EUMCall(t *testing.T) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder))
	defer instana.ShutdownSensor()

	h := instana.TracingNamedHandlerFunc(s, "test-handler", "/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(w, "Ok")
	})

	rec := httptest.NewRecorder()

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(instana.FieldL, "1,correlationType=web;correlationId=eum correlation id")

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)
	assert.Equal(t, "web", spans[0].CorrelationType)
	assert.Equal(t, "eum correlation id", spans[0].CorrelationID)
}

func TestTracingHandlerFunc_PanicHandling(t *testing.T) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder))
	defer instana.ShutdownSensor()

	h := instana.TracingNamedHandlerFunc(s, "test", "/test", func(w http.ResponseWriter, req *http.Request) {
		panic("something went wrong")
	})

	rec := httptest.NewRecorder()
	assert.Panics(t, func() {
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/test?q=term", nil))
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
		Status:  http.StatusInternalServerError,
		Method:  "GET",
		Host:    "example.com",
		Path:    "/test",
		Params:  "q=term",
		RouteID: "test",
		Error:   "something went wrong",
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

func TestRoundTripper(t *testing.T) {
	recorder := instana.NewTestRecorder()
	opts := &instana.Options{
		Service: TestServiceName,
		Tracer: instana.TracerOptions{
			CollectableHTTPHeaders: []string{"x-custom-header-1", "x-custom-header-2"},
		},
		AgentClient: alwaysReadyClient{},
	}
	tracer := instana.NewTracerWithEverything(opts, recorder)
	s := instana.NewSensorWithTracer(tracer)
	defer instana.ShutdownSensor()

	parentSpan := tracer.StartSpan("parent")

	var traceIDHeader, spanIDHeader string
	rt := instana.RoundTripper(s, testRoundTripper(func(req *http.Request) (*http.Response, error) {
		traceIDHeader = req.Header.Get(instana.FieldT)
		spanIDHeader = req.Header.Get(instana.FieldS)

		return &http.Response{
			Status:     http.StatusText(http.StatusNotImplemented),
			StatusCode: http.StatusNotImplemented,
			Header: http.Header{
				"X-Response":        []string{"true"},
				"X-Custom-Header-2": []string{"response"},
			},
		}, nil
	}))

	ctx := instana.ContextWithSpan(context.Background(), parentSpan)
	req := httptest.NewRequest("GET", "http://user:password@example.com/hello?q=term&sensitive_key=s3cr3t&myPassword=qwerty&SECRET_VALUE=1", nil)
	req.Header.Set("X-Custom-Header-1", "request")
	req.Header.Set("Authorization", "Basic blah")

	_, err := rt.RoundTrip(req.WithContext(ctx))
	require.NoError(t, err)

	parentSpan.Finish()

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 2)

	cSpan, pSpan := spans[0], spans[1]
	assert.Equal(t, 0, cSpan.Ec)
	assert.EqualValues(t, instana.ExitSpanKind, cSpan.Kind)

	assert.Equal(t, pSpan.TraceID, cSpan.TraceID)
	assert.Equal(t, pSpan.SpanID, cSpan.ParentID)

	assert.Equal(t, instana.FormatID(cSpan.TraceID), traceIDHeader)
	assert.Equal(t, instana.FormatID(cSpan.SpanID), spanIDHeader)

	require.IsType(t, instana.HTTPSpanData{}, cSpan.Data)
	data := cSpan.Data.(instana.HTTPSpanData)

	assert.Equal(t, instana.HTTPSpanTags{
		Method: "GET",
		Status: http.StatusNotImplemented,
		URL:    "http://example.com/hello",
		Params: "SECRET_VALUE=%3Credacted%3E&myPassword=%3Credacted%3E&q=term&sensitive_key=%3Credacted%3E",
		Headers: map[string]string{
			"x-custom-header-1": "request",
			"x-custom-header-2": "response",
		},
	}, data.Tags)
}

func TestRoundTripper_WithoutParentSpan(t *testing.T) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder))
	defer instana.ShutdownSensor()

	rt := instana.RoundTripper(s, testRoundTripper(func(req *http.Request) (*http.Response, error) {
		assert.Empty(t, req.Header.Get(instana.FieldT))
		assert.Empty(t, req.Header.Get(instana.FieldS))

		return &http.Response{
			Status:     http.StatusText(http.StatusNotImplemented),
			StatusCode: http.StatusNotImplemented,
		}, nil
	}))

	resp, err := rt.RoundTrip(httptest.NewRequest("GET", "http://example.com/hello", nil))
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotImplemented, resp.StatusCode)

	assert.Empty(t, recorder.GetQueuedSpans())
}

func TestRoundTripper_Error(t *testing.T) {
	serverErr := errors.New("something went wrong")

	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder))
	defer instana.ShutdownSensor()

	rt := instana.RoundTripper(s, testRoundTripper(func(req *http.Request) (*http.Response, error) {
		return nil, serverErr
	}))

	ctx := instana.ContextWithSpan(context.Background(), s.Tracer().StartSpan("parent"))
	req := httptest.NewRequest("GET", "http://example.com/hello?q=term&key=s3cr3t", nil)

	_, err := rt.RoundTrip(req.WithContext(ctx))
	assert.Error(t, err)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 2)

	span, logSpan := spans[0], spans[1]
	assert.Equal(t, 1, span.Ec)
	assert.EqualValues(t, instana.ExitSpanKind, span.Kind)

	require.IsType(t, instana.HTTPSpanData{}, span.Data)
	data := span.Data.(instana.HTTPSpanData)

	assert.Equal(t, instana.HTTPSpanTags{
		Method: "GET",
		URL:    "http://example.com/hello",
		Params: "key=%3Credacted%3E&q=term",
		Error:  "something went wrong",
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

func TestRoundTripper_DefaultTransport(t *testing.T) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder))
	defer instana.ShutdownSensor()

	var numCalls int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		numCalls++

		assert.NotEmpty(t, req.Header.Get(instana.FieldT))
		assert.NotEmpty(t, req.Header.Get(instana.FieldS))

		w.Write([]byte("OK"))
	}))
	defer ts.Close()

	rt := instana.RoundTripper(s, nil)

	ctx := instana.ContextWithSpan(context.Background(), s.Tracer().StartSpan("parent"))
	req := httptest.NewRequest("GET", ts.URL+"/hello", nil)

	resp, err := rt.RoundTrip(req.WithContext(ctx))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	assert.Equal(t, 1, numCalls)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	assert.Equal(t, 0, span.Ec)
	assert.EqualValues(t, instana.ExitSpanKind, span.Kind)

	require.IsType(t, instana.HTTPSpanData{}, span.Data)
	data := span.Data.(instana.HTTPSpanData)

	assert.Equal(t, instana.HTTPSpanTags{
		Status: http.StatusOK,
		Method: "GET",
		URL:    ts.URL + "/hello",
	}, data.Tags)
}

type testRoundTripper func(*http.Request) (*http.Response, error)

func (rt testRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return rt(req)
}

type alwaysReadyClient struct{}

func (alwaysReadyClient) Ready() bool                                       { return true }
func (alwaysReadyClient) SendMetrics(data acceptor.Metrics) error           { return nil }
func (alwaysReadyClient) SendEvent(event *instana.EventData) error          { return nil }
func (alwaysReadyClient) SendSpans(spans []instana.Span) error              { return nil }
func (alwaysReadyClient) SendProfiles(profiles []autoprofile.Profile) error { return nil }
func (alwaysReadyClient) Flush(context.Context) error                       { return nil }
