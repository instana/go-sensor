// (c) Copyright IBM Corp. 2024

package instafasthttp_test

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"testing"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"
	"github.com/instana/go-sensor/instrumentation/instafasthttp"
	"github.com/instana/go-sensor/w3ctrace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
)

func BenchmarkTracingHandlerFunc(b *testing.B) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	s := instana.NewSensorWithTracer(tracer)

	h := instafasthttp.TraceHandler(s, "action", "/{action}", func(ctx *fasthttp.RequestCtx) {
		ctx.SetStatusCode(fasthttp.StatusOK)
		fmt.Fprintf(ctx, "Ok")
	})

	server := &fasthttp.Server{
		Handler: h,
	}

	ln := fasthttputil.NewInmemoryListener()

	go func() {
		if err := server.Serve(ln); err != nil {
			b.Errorf("unexpected error: %v", err)
		}
	}()

	b.ResetTimer()

	for i := 0; i < 10; i++ {
		conn, err := ln.Dial()
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}

		if _, err = conn.Write([]byte("GET / HTTP/1.1\r\nHost: google.com\r\n\r\n")); err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}

	if err := ln.Close(); err != nil {
		b.Fatalf("unexpected error: %v", err)
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

	h := instafasthttp.TraceHandler(s, "action", "/{action}", func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.Add("X-Response", "true")
		ctx.Response.Header.Add("X-Custom-Header-2", "response")
		ctx.Success("aaa/bbb", []byte("Ok response!"))
	})

	server := &fasthttp.Server{
		Handler: h,
	}

	ln := fasthttputil.NewInmemoryListener()

	go func() {
		if err := server.Serve(ln); err != nil {
			assert.NoError(t, err, "unexpected error: %v", err)
		}
	}()

	conn, err := ln.Dial()
	if err != nil {
		assert.NoError(t, err, "unexpected error: %v", err)
	}

	if _, err = conn.Write([]byte("GET /test?q=term HTTP/1.1\r\nHost: example.com\r\nAuthorization: Basic\r\nX-Custom-Header-1: request\r\n\r\n")); err != nil {
		assert.NoError(t, err, "unexpected error: %v", err)
	}

	br := bufio.NewReader(conn)
	resp := verifyResponse(t, br, fasthttp.StatusOK, "aaa/bbb", "Ok response!")

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
		Status: fasthttp.StatusOK,
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
	assert.Equal(t, instana.FormatID(span.TraceID), string(resp.Header.Peek(instana.FieldT)))
	assert.Equal(t, instana.FormatID(span.SpanID), string(resp.Header.Peek(instana.FieldS)))

	// w3c trace context
	traceparent := string(resp.Header.Peek(w3ctrace.TraceParentHeader))
	assert.Contains(t, traceparent, instana.FormatLongID(span.TraceIDHi, span.TraceID))
	assert.Contains(t, traceparent, instana.FormatID(span.SpanID))

	tracestate := string(resp.Header.Peek(w3ctrace.TraceStateHeader))
	assert.True(t, strings.HasPrefix(
		tracestate,
		"in="+instana.FormatID(span.TraceID)+";"+instana.FormatID(span.SpanID),
	), tracestate)

	if err := ln.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTracingHandlerFunc_InstanaFieldLPriorityOverTraceParentHeader(t *testing.T) {
	type testCase struct {
		headers                 map[string]string
		traceParentHeaderSuffix string
	}

	testCases := map[string]testCase{
		"traceparent is suppressed, x-instana-l is not suppressed": {
			headers: map[string]string{
				w3ctrace.TraceParentHeader: "00-00000000000000000000000000000001-0000000000000001-00",
				instana.FieldL:             "1",
			},
			traceParentHeaderSuffix: "-01",
		},
		"traceparent is suppressed, x-instana-l is absent (is not suppressed by default)": {
			headers: map[string]string{
				w3ctrace.TraceParentHeader: "00-00000000000000000000000000000001-0000000000000001-00",
			},
			traceParentHeaderSuffix: "-01",
		},
		"traceparent is not suppressed, x-instana-l is absent (tracing enabled by default)": {
			headers: map[string]string{
				w3ctrace.TraceParentHeader: "00-00000000000000000000000000000001-0000000000000001-01",
			},
			traceParentHeaderSuffix: "-01",
		},
		"traceparent is not suppressed, x-instana-l is not suppressed": {
			headers: map[string]string{
				w3ctrace.TraceParentHeader: "00-00000000000000000000000000000001-0000000000000001-01",
				instana.FieldL:             "1",
			},
			traceParentHeaderSuffix: "-01",
		},
		"traceparent is suppressed, x-instana-l is suppressed": {
			headers: map[string]string{
				w3ctrace.TraceParentHeader: "00-00000000000000000000000000000001-0000000000000001-00",
				instana.FieldL:             "0",
			},
			traceParentHeaderSuffix: "-00",
		},
		"traceparent is not suppressed, x-instana-l is suppressed": {
			headers: map[string]string{
				w3ctrace.TraceParentHeader: "00-00000000000000000000000000000001-0000000000000001-01",
				instana.FieldL:             "0",
			},
			traceParentHeaderSuffix: "-00",
		},
	}

	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{
		Service:     "go-sensor-test",
		AgentClient: alwaysReadyClient{},
	}, recorder))

	h := instafasthttp.TraceHandler(s, "action", "/{action}", func(ctx *fasthttp.RequestCtx) {
		ctx.Success("aaa/bbb", []byte("Ok response!"))
	})

	server := &fasthttp.Server{Handler: h}
	ln := fasthttputil.NewInmemoryListener()
	go func() {
		if err := server.Serve(ln); err != nil {
			assert.NoError(t, err, "unexpected error: %v", err)
		}
	}()

	for name, testCase := range testCases {

		conn, err := ln.Dial()
		if err != nil {
			assert.NoError(t, err, "unexpected error: %v", err)
		}

		url := "GET /test?q=term HTTP/1.1\r\nHost: example.com"
		for k, v := range testCase.headers {
			url = url + "\r\n" + k + ": " + v
		}
		url = url + "\r\n\r\n"

		_, err = conn.Write([]byte(url))
		if err != nil {
			assert.NoError(t, err, "unexpected error: %v", err)
		}

		br := bufio.NewReader(conn)

		resp := verifyResponse(t, br, fasthttp.StatusOK, "aaa/bbb", "Ok response!")

		assert.Equal(t, fasthttp.StatusOK, resp.StatusCode())
		assert.True(t, strings.HasSuffix(string(resp.Header.Peek(w3ctrace.TraceParentHeader)), testCase.traceParentHeaderSuffix), "case '"+name+"' failed")
	}

	if err := ln.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTracingHandlerFunc_WriteHeaders(t *testing.T) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder))

	h := instafasthttp.TraceHandler(s, "test", "", func(ctx *fasthttp.RequestCtx) {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
	})

	server := &fasthttp.Server{Handler: h}
	ln := fasthttputil.NewInmemoryListener()
	go func() {
		if err := server.Serve(ln); err != nil {
			assert.NoError(t, err, "unexpected error: %v", err)
		}
	}()

	conn, err := ln.Dial()
	if err != nil {
		assert.NoError(t, err, "unexpected error: %v", err)
	}

	if _, err = conn.Write([]byte("GET /test?q=term HTTP/1.1\r\nHost: example.com\r\n\r\n")); err != nil {
		assert.NoError(t, err, "unexpected error: %v", err)
	}

	br := bufio.NewReader(conn)

	resp := verifyResponse(t, br, fasthttp.StatusNotFound, "text/plain; charset=utf-8", "")

	assert.Equal(t, fasthttp.StatusNotFound, resp.StatusCode())

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
		Status:   fasthttp.StatusNotFound,
		Method:   "GET",
		Host:     "example.com",
		Path:     "/test",
		Params:   "q=term",
		RouteID:  "test",
		Protocol: "http",
	}, data.Tags)

	// check whether the trace context has been sent back to the client
	assert.Equal(t, instana.FormatID(span.TraceID), string(resp.Header.Peek(instana.FieldT)))
	assert.Equal(t, instana.FormatID(span.SpanID), string(resp.Header.Peek(instana.FieldS)))

	// w3c trace context
	traceparent := string(resp.Header.Peek(w3ctrace.TraceParentHeader))
	assert.Contains(t, traceparent, instana.FormatLongID(span.TraceIDHi, span.TraceID))
	assert.Contains(t, traceparent, instana.FormatID(span.SpanID))

	tracestate := string(resp.Header.Peek(w3ctrace.TraceStateHeader))
	assert.True(t, strings.HasPrefix(
		tracestate,
		"in="+instana.FormatID(span.TraceID)+";"+instana.FormatID(span.SpanID),
	), tracestate)

	if err := ln.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTracingHandlerFunc_W3CTraceContext(t *testing.T) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder))

	h := instafasthttp.TraceHandler(s, "test", "", func(ctx *fasthttp.RequestCtx) {
		ctx.Success("aaa/bbb", []byte("Ok response!"))
	})

	server := &fasthttp.Server{Handler: h}
	ln := fasthttputil.NewInmemoryListener()
	go func() {
		if err := server.Serve(ln); err != nil {
			assert.NoError(t, err, "unexpected error: %v", err)
		}
	}()

	conn, err := ln.Dial()
	if err != nil {
		assert.NoError(t, err, "unexpected error: %v", err)
	}

	url := "GET /test HTTP/1.1\r\nHost: example.com"
	// add trace parent header
	url = url + "\r\n" + w3ctrace.TraceParentHeader + ": " + "00-00000000000000010000000000000002-0000000000000003-01"
	// add trace state header
	url = url + "\r\n" + w3ctrace.TraceStateHeader + ": " + "in=1234;5678,rojo=00f067aa0ba902b7"
	url = url + "\r\n\r\n"

	if _, err = conn.Write([]byte(url)); err != nil {
		assert.NoError(t, err, "unexpected error: %v", err)
	}

	br := bufio.NewReader(conn)
	resp := verifyResponse(t, br, fasthttp.StatusOK, "aaa/bbb", "Ok response!")

	assert.Equal(t, fasthttp.StatusOK, resp.StatusCode())

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
		Status:   fasthttp.StatusOK,
		Method:   "GET",
		Path:     "/test",
		RouteID:  "test",
		Protocol: "http",
	}, data.Tags)

	// check whether the trace context has been sent back to the client
	assert.Equal(t, instana.FormatID(span.TraceID), string(resp.Header.Peek(instana.FieldT)))
	assert.Equal(t, instana.FormatID(span.SpanID), string(resp.Header.Peek(instana.FieldS)))

	// w3c trace context
	traceparent := string(resp.Header.Peek(w3ctrace.TraceParentHeader))
	assert.Contains(t, traceparent, instana.FormatLongID(span.TraceIDHi, span.TraceID))
	assert.Contains(t, traceparent, instana.FormatID(span.SpanID))

	tracestate := string(resp.Header.Peek(w3ctrace.TraceStateHeader))
	assert.True(t, strings.HasPrefix(
		tracestate,
		"in="+instana.FormatID(span.TraceID)+";"+instana.FormatID(span.SpanID),
	), tracestate)

	if err := ln.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTracingHandlerFunc_SecretsFiltering(t *testing.T) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder))

	h := instafasthttp.TraceHandler(s, "test", "/{action}", func(ctx *fasthttp.RequestCtx) {
		ctx.Success("aaa/bbb", []byte("Ok response!"))
	})

	server := &fasthttp.Server{Handler: h}
	ln := fasthttputil.NewInmemoryListener()
	go func() {
		if err := server.Serve(ln); err != nil {
			assert.NoError(t, err, "unexpected error: %v", err)
		}
	}()

	conn, err := ln.Dial()
	if err != nil {
		assert.NoError(t, err, "unexpected error: %v", err)
	}

	url := "GET /test?q=term&sensitive_key=s3cr3t&myPassword=qwerty&SECRET_VALUE=1 HTTP/1.1\r\nHost: example.com\r\n\r\n"

	if _, err = conn.Write([]byte(url)); err != nil {
		assert.NoError(t, err, "unexpected error: %v", err)
	}

	br := bufio.NewReader(conn)
	resp := verifyResponse(t, br, fasthttp.StatusOK, "aaa/bbb", "Ok response!")

	assert.Equal(t, fasthttp.StatusOK, resp.StatusCode())

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
		Status:       fasthttp.StatusOK,
		Method:       "GET",
		Path:         "/test",
		Params:       "SECRET_VALUE=%3Credacted%3E&myPassword=%3Credacted%3E&q=term&sensitive_key=%3Credacted%3E",
		PathTemplate: "/{action}",
		RouteID:      "test",
		Protocol:     "http",
	}, data.Tags)

	// check whether the trace context has been sent back to the client
	assert.Equal(t, instana.FormatID(span.TraceID), string(resp.Header.Peek(instana.FieldT)))
	assert.Equal(t, instana.FormatID(span.SpanID), string(resp.Header.Peek(instana.FieldS)))

	if err := ln.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTracingHandlerFunc_SyntheticCall(t *testing.T) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}, Tracer: instana.TracerOptions{CollectableHTTPHeaders: []string{instana.FieldSynthetic}}}, recorder))

	h := instafasthttp.TraceHandler(s, "test", "/{action}", func(ctx *fasthttp.RequestCtx) {
		ctx.Success("aaa/bbb", []byte("Ok response!"))
	})

	server := &fasthttp.Server{Handler: h}
	ln := fasthttputil.NewInmemoryListener()
	go func() {
		if err := server.Serve(ln); err != nil {
			assert.NoError(t, err, "unexpected error: %v", err)
		}
	}()

	conn, err := ln.Dial()
	if err != nil {
		assert.NoError(t, err, "unexpected error: %v", err)
	}

	url := "GET /test HTTP/1.1\r\nHost: example.com\r\n" + instana.FieldSynthetic + ": 1" + "\r\n\r\n"

	if _, err = conn.Write([]byte(url)); err != nil {
		assert.NoError(t, err, "unexpected error: %v", err)
	}

	br := bufio.NewReader(conn)
	resp := verifyResponse(t, br, fasthttp.StatusOK, "aaa/bbb", "Ok response!")

	assert.Equal(t, fasthttp.StatusOK, resp.StatusCode())

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span := spans[0]

	assert.True(t, span.Synthetic)

	if err := ln.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTracingHandlerFunc_EUMCall(t *testing.T) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}, Tracer: instana.TracerOptions{CollectableHTTPHeaders: []string{instana.FieldSynthetic}}}, recorder))

	h := instafasthttp.TraceHandler(s, "test", "/{action}", func(ctx *fasthttp.RequestCtx) {
		ctx.Success("aaa/bbb", []byte("Ok response!"))
	})

	server := &fasthttp.Server{Handler: h}
	ln := fasthttputil.NewInmemoryListener()
	go func() {
		if err := server.Serve(ln); err != nil {
			assert.NoError(t, err, "unexpected error: %v", err)
		}
	}()

	conn, err := ln.Dial()
	if err != nil {
		assert.NoError(t, err, "unexpected error: %v", err)
	}

	url := "GET /test HTTP/1.1\r\nHost: example.com\r\n" + instana.FieldL + ": 1,correlationType=web;correlationId=eum correlation id" + "\r\n\r\n"

	if _, err = conn.Write([]byte(url)); err != nil {
		assert.NoError(t, err, "unexpected error: %v", err)
	}

	br := bufio.NewReader(conn)
	resp := verifyResponse(t, br, fasthttp.StatusOK, "aaa/bbb", "Ok response!")

	assert.Equal(t, fasthttp.StatusOK, resp.StatusCode())

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)
	assert.Equal(t, "web", spans[0].CorrelationType)
	assert.Equal(t, "eum correlation id", spans[0].CorrelationID)

	if err := ln.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTracingHandlerFunc_PanicHandling(t *testing.T) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}, Tracer: instana.TracerOptions{CollectableHTTPHeaders: []string{instana.FieldSynthetic}}}, recorder))

	h := instafasthttp.TraceHandler(s, "test", "/{action}", func(ctx *fasthttp.RequestCtx) {
		panic("something went wrong")
	})

	c := &fasthttp.RequestCtx{}

	c.Request.Header.SetMethod(fasthttp.MethodGet)
	c.Request.Header.Set(instana.FieldL, "1,correlationType=web;correlationId=eum correlation id")
	c.URI().SetPath("/test")
	c.URI().SetQueryString("q=term")
	c.URI().SetHost("example.com")

	assert.Panics(t, func() {
		h(c)
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
		Status:       fasthttp.StatusInternalServerError,
		Method:       "GET",
		Host:         "example.com",
		Path:         "/test",
		Params:       "q=term",
		RouteID:      "test",
		Error:        "something went wrong",
		Protocol:     "http",
		PathTemplate: "/{action}",
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
		Service: "test-service",
		Tracer: instana.TracerOptions{
			CollectableHTTPHeaders: []string{"x-custom-header-1", "x-custom-header-2"},
		},
		AgentClient: alwaysReadyClient{},
	}
	tracer := instana.NewTracerWithEverything(opts, recorder)
	s := instana.NewSensorWithTracer(tracer)

	parentSpan := tracer.StartSpan("parent")
	ctx := instana.ContextWithSpan(context.Background(), parentSpan)

	server := &fasthttp.Server{
		Handler: func(ctx *fasthttp.RequestCtx) {
			ctx.Response.Header.Add("X-Response", "true")
			ctx.Response.Header.Add("X-Custom-Header-2", "response")
			ctx.Success("aaa/bbb", []byte("Ok response!"))
		},
	}

	ln := fasthttputil.NewInmemoryListener()

	go func() {
		if err := server.Serve(ln); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}()

	testT := func() fasthttp.RoundTripper {
		c, _ := ln.Dial()
		br := bufio.NewReader(c)
		bw := bufio.NewWriter(c)
		return &transportTest{br: br, bw: bw}
	}()

	hc := &fasthttp.HostClient{
		Transport: instafasthttp.RoundTripper(ctx, s, testT),
		Addr:      "example.com",
	}

	r := &fasthttp.Request{}
	r.Header.SetMethod(fasthttp.MethodGet)
	r.Header.Set("X-Custom-Header-1", "request")
	r.Header.Set("Authorization", "Basic blah")
	r.URI().SetPath("/hello")
	r.URI().SetQueryString("q=term&sensitive_key=s3cr3t&myPassword=qwerty&SECRET_VALUE=1")
	r.URI().SetHost("example.com")

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// Make the request
	err := hc.Do(r, resp)

	require.NoError(t, err)

	parentSpan.Finish()

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 2)

	cSpan, pSpan := spans[0], spans[1]
	assert.Equal(t, 0, cSpan.Ec)
	assert.EqualValues(t, instana.ExitSpanKind, cSpan.Kind)

	assert.Equal(t, pSpan.TraceID, cSpan.TraceID)
	assert.Equal(t, pSpan.SpanID, cSpan.ParentID)

	assert.Equal(t, instana.FormatID(cSpan.TraceID), testT.(*transportTest).traceIDHeader)
	assert.Equal(t, instana.FormatID(cSpan.SpanID), testT.(*transportTest).spanIDHeader)

	require.IsType(t, instana.HTTPSpanData{}, cSpan.Data)
	data := cSpan.Data.(instana.HTTPSpanData)

	assert.Equal(t, instana.HTTPSpanTags{
		Method: "GET",
		Status: fasthttp.StatusOK,
		URL:    "http://example.com/hello",
		Params: "SECRET_VALUE=%3Credacted%3E&myPassword=%3Credacted%3E&q=term&sensitive_key=%3Credacted%3E",
		Headers: map[string]string{
			"x-custom-header-1": "request",
			"x-custom-header-2": "response",
		},
	}, data.Tags)

	if err := ln.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRoundTripper_Error(t *testing.T) {

	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder))

	parentSpan := s.Tracer().StartSpan("parent")
	ctx := instana.ContextWithSpan(context.Background(), parentSpan)

	server := &fasthttp.Server{
		Handler: func(ctx *fasthttp.RequestCtx) {
			ctx.Response.Header.Add("X-Response", "true")
			ctx.Response.Header.Add("X-Custom-Header-2", "response")
			ctx.Success("aaa/bbb", []byte("Ok response!"))
		},
	}

	ln := fasthttputil.NewInmemoryListener()

	go func() {
		if err := server.Serve(ln); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}()

	testT := func() fasthttp.RoundTripper {
		c, _ := ln.Dial()
		br := bufio.NewReader(c)
		bw := bufio.NewWriter(c)
		return &transportTest{br: br, bw: bw, isErr: true}
	}()

	hc := &fasthttp.HostClient{
		Transport: instafasthttp.RoundTripper(ctx, s, testT),
		Addr:      "example.com",
	}

	r := &fasthttp.Request{}
	r.Header.SetMethod(fasthttp.MethodGet)
	r.Header.Set("Authorization", "Basic blah")
	r.URI().SetPath("/hello")
	r.URI().SetQueryString("q=term&key=s3cr3t")
	r.URI().SetHost("example.com")

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// Make the request
	err := hc.Do(r, resp)

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
		Message: `error.object: "something went wrong"`,
	}, logData.Tags)

	if err := ln.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRoundTripper_DefaultTransport(t *testing.T) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder))
	var numCalls int
	parentSpan := s.Tracer().StartSpan("parent")
	ctx := instana.ContextWithSpan(context.Background(), parentSpan)

	server := &fasthttp.Server{
		Handler: func(ctx *fasthttp.RequestCtx) {
			numCalls++
			// ctx.Response.Header.Add("X-Response", "true")
			// ctx.Response.Header.Add("X-Custom-Header-2", "response")
			ctx.Success("aaa/bbb", []byte("Ok response!"))
		},
	}

	ln := fasthttputil.NewInmemoryListener()

	go func() {
		if err := server.Serve(ln); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}()

	hc := &fasthttp.HostClient{
		Transport: instafasthttp.RoundTripper(ctx, s, nil),
		Addr:      "example.com",
		Dial:      func(addr string) (net.Conn, error) { return ln.Dial() },
	}

	r := &fasthttp.Request{}
	r.Header.SetMethod(fasthttp.MethodGet)
	r.Header.Set("Authorization", "Basic blah")
	r.URI().SetPath("/hello")
	r.URI().SetHost("example.com")

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// Make the request
	err := hc.Do(r, resp)

	require.NoError(t, err)
	assert.Equal(t, fasthttp.StatusOK, resp.StatusCode())

	assert.Equal(t, 1, numCalls)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	assert.Equal(t, 0, span.Ec)
	assert.EqualValues(t, instana.ExitSpanKind, span.Kind)

	require.IsType(t, instana.HTTPSpanData{}, span.Data)
	data := span.Data.(instana.HTTPSpanData)

	assert.Equal(t, instana.HTTPSpanTags{
		Status: fasthttp.StatusOK,
		Method: "GET",
		URL:    "http://example.com/hello",
	}, data.Tags)

	if err := ln.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func verifyResponse(t *testing.T, r *bufio.Reader, expectedStatusCode int, expectedContentType, expectedBody string) *fasthttp.Response {
	var resp fasthttp.Response
	if err := resp.Read(r); err != nil {
		t.Fatalf("Unexpected error when parsing response: %v", err)
	}

	if !bytes.Equal(resp.Body(), []byte(expectedBody)) {
		t.Fatalf("Unexpected body %q. Expected %q", resp.Body(), []byte(expectedBody))
	}
	verifyResponseHeader(t, &resp.Header, expectedStatusCode, len(resp.Body()), expectedContentType, "")
	return &resp
}

func verifyResponseHeader(t *testing.T, h *fasthttp.ResponseHeader, expectedStatusCode, expectedContentLength int, expectedContentType, expectedContentEncoding string) {
	if h.StatusCode() != expectedStatusCode {
		t.Fatalf("Unexpected status code %d. Expected %d", h.StatusCode(), expectedStatusCode)
	}
	if h.ContentLength() != expectedContentLength {
		t.Fatalf("Unexpected content length %d. Expected %d", h.ContentLength(), expectedContentLength)
	}
	if string(h.ContentType()) != expectedContentType {
		t.Fatalf("Unexpected content type %q. Expected %q", h.ContentType(), expectedContentType)
	}
	if string(h.ContentEncoding()) != expectedContentEncoding {
		t.Fatalf("Unexpected content encoding %q. Expected %q", h.ContentEncoding(), expectedContentEncoding)
	}
}

type alwaysReadyClient struct{}

func (alwaysReadyClient) Ready() bool                                       { return true }
func (alwaysReadyClient) SendMetrics(data acceptor.Metrics) error           { return nil }
func (alwaysReadyClient) SendEvent(event *instana.EventData) error          { return nil }
func (alwaysReadyClient) SendSpans(spans []instana.Span) error              { return nil }
func (alwaysReadyClient) SendProfiles(profiles []autoprofile.Profile) error { return nil }
func (alwaysReadyClient) Flush(context.Context) error                       { return nil }

type transportTest struct {
	// If the transport is expected to return an error
	isErr bool

	br *bufio.Reader
	bw *bufio.Writer

	// for extracting tracer headers from request
	traceIDHeader string
	spanIDHeader  string
}

func (t *transportTest) RoundTrip(hc *fasthttp.HostClient, req *fasthttp.Request, res *fasthttp.Response) (retry bool, err error) {
	if t.isErr {
		serverErr := errors.New("something went wrong")
		return false, serverErr
	}

	if err = req.Write(t.bw); err != nil {
		return false, err
	}
	if err = t.bw.Flush(); err != nil {
		return false, err
	}

	// extract tracer specific headers
	t.traceIDHeader = string(req.Header.Peek(instana.FieldT))
	t.spanIDHeader = string(req.Header.Peek(instana.FieldS))

	err = res.Read(t.br)
	return err != nil, err
}
