// (c) Copyright IBM Corp. 2024

package instafasthttp_test

import (
	"bufio"
	"context"
	"errors"
	"net"
	"testing"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"
	"github.com/instana/go-sensor/instrumentation/instafasthttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
)

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
