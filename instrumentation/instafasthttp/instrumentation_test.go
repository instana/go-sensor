// (c) Copyright IBM Corp. 2024

package instafasthttp_test

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net/http"
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

type alwaysReadyClient struct{}

func (alwaysReadyClient) Ready() bool                                       { return true }
func (alwaysReadyClient) SendMetrics(data acceptor.Metrics) error           { return nil }
func (alwaysReadyClient) SendEvent(event *instana.EventData) error          { return nil }
func (alwaysReadyClient) SendSpans(spans []instana.Span) error              { return nil }
func (alwaysReadyClient) SendProfiles(profiles []autoprofile.Profile) error { return nil }
func (alwaysReadyClient) Flush(context.Context) error                       { return nil }

func BenchmarkTracingHandlerFunc(b *testing.B) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	s := instana.NewSensorWithTracer(tracer)
	// defer instana.ShutdownSensor()

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
	// defer instana.ShutdownSensor()

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
	// defer instana.ShutdownSensor()

	h := instafasthttp.TraceHandler(s, "action", "/{action}", func(ctx *fasthttp.RequestCtx) {
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
