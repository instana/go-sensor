package instafasthttp_test

import (
	"context"
	"net"
	"testing"
	"time"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instafasthttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
)

func TestClient_Do(t *testing.T) {
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

	var fieldTFrmHeader, fieldSFrmHeader string

	server := &fasthttp.Server{
		Handler: func(ctx *fasthttp.RequestCtx) {
			// get the header span and trace id from request header
			fieldTFrmHeader = string(ctx.Request.Header.Peek(instana.FieldT))
			fieldSFrmHeader = string(ctx.Request.Header.Peek(instana.FieldS))
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

	c := &fasthttp.Client{
		Dial: func(addr string) (net.Conn, error) { return ln.Dial() },
	}
	ic := instafasthttp.GetInstrumentedClient(s, c)

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
	err := ic.Do(ctx, r, resp)

	require.NoError(t, err)

	parentSpan.Finish()

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 2)

	cSpan, pSpan := spans[0], spans[1]
	assert.Equal(t, 0, cSpan.Ec)
	assert.EqualValues(t, instana.ExitSpanKind, cSpan.Kind)

	assert.Equal(t, pSpan.TraceID, cSpan.TraceID)
	assert.Equal(t, pSpan.SpanID, cSpan.ParentID)

	assert.Equal(t, instana.FormatID(cSpan.TraceID), fieldTFrmHeader)
	assert.Equal(t, instana.FormatID(cSpan.SpanID), fieldSFrmHeader)

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

func TestClient_Do_Error(t *testing.T) {

	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder))

	parentSpan := s.Tracer().StartSpan("parent")
	ctx := instana.ContextWithSpan(context.Background(), parentSpan)

	server := &fasthttp.Server{
		Handler: func(ctx *fasthttp.RequestCtx) {
			ctx.Success("aaa/bbb", []byte("Ok response!"))
		},
	}

	ln := fasthttputil.NewInmemoryListener()

	go func() {
		if err := server.Serve(ln); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}()

	ln.Close()

	c := &fasthttp.Client{
		Dial: func(addr string) (net.Conn, error) { return ln.Dial() },
	}
	ic := instafasthttp.GetInstrumentedClient(s, c)

	r := &fasthttp.Request{}
	r.Header.SetMethod(fasthttp.MethodGet)
	r.Header.Set("Authorization", "Basic blah")
	r.URI().SetPath("/hello")
	r.URI().SetQueryString("q=term&key=s3cr3t")
	r.URI().SetHost("example.com")

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// Make the request
	err := ic.Do(ctx, r, resp)

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
		Error:  "InmemoryListener is already closed: use of closed network connection",
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
		Message: `error.object: "InmemoryListener is already closed: use of closed network connection"`,
	}, logData.Tags)
}

func TestClient_DoTimeout(t *testing.T) {
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

	var fieldTFrmHeader, fieldSFrmHeader string

	server := &fasthttp.Server{
		Handler: func(ctx *fasthttp.RequestCtx) {
			// get the header span and trace id from request header
			fieldTFrmHeader = string(ctx.Request.Header.Peek(instana.FieldT))
			fieldSFrmHeader = string(ctx.Request.Header.Peek(instana.FieldS))
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

	c := &fasthttp.Client{
		Dial: func(addr string) (net.Conn, error) { return ln.Dial() },
	}
	ic := instafasthttp.GetInstrumentedClient(s, c)

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
	err := ic.DoTimeout(ctx, r, resp, time.Minute*10)

	require.NoError(t, err)

	parentSpan.Finish()

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 2)

	cSpan, pSpan := spans[0], spans[1]
	assert.Equal(t, 0, cSpan.Ec)
	assert.EqualValues(t, instana.ExitSpanKind, cSpan.Kind)

	assert.Equal(t, pSpan.TraceID, cSpan.TraceID)
	assert.Equal(t, pSpan.SpanID, cSpan.ParentID)

	assert.Equal(t, instana.FormatID(cSpan.TraceID), fieldTFrmHeader)
	assert.Equal(t, instana.FormatID(cSpan.SpanID), fieldSFrmHeader)

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

func TestClient_DoTimeout_Error(t *testing.T) {

	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder))

	parentSpan := s.Tracer().StartSpan("parent")
	ctx := instana.ContextWithSpan(context.Background(), parentSpan)

	server := &fasthttp.Server{
		Handler: func(ctx *fasthttp.RequestCtx) {
			ctx.Success("aaa/bbb", []byte("Ok response!"))
		},
	}

	ln := fasthttputil.NewInmemoryListener()

	go func() {
		if err := server.Serve(ln); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}()

	ln.Close()

	c := &fasthttp.Client{
		Dial: func(addr string) (net.Conn, error) { return ln.Dial() },
	}
	ic := instafasthttp.GetInstrumentedClient(s, c)

	r := &fasthttp.Request{}
	r.Header.SetMethod(fasthttp.MethodGet)
	r.Header.Set("Authorization", "Basic blah")
	r.URI().SetPath("/hello")
	r.URI().SetQueryString("q=term&key=s3cr3t")
	r.URI().SetHost("example.com")

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// Make the request
	err := ic.DoTimeout(ctx, r, resp, time.Minute*10)

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
		Error:  "InmemoryListener is already closed: use of closed network connection",
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
		Message: `error.object: "InmemoryListener is already closed: use of closed network connection"`,
	}, logData.Tags)
}

func TestClient_DoDeadline(t *testing.T) {
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

	var fieldTFrmHeader, fieldSFrmHeader string

	server := &fasthttp.Server{
		Handler: func(ctx *fasthttp.RequestCtx) {
			// get the header span and trace id from request header
			fieldTFrmHeader = string(ctx.Request.Header.Peek(instana.FieldT))
			fieldSFrmHeader = string(ctx.Request.Header.Peek(instana.FieldS))
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

	c := &fasthttp.Client{
		Dial: func(addr string) (net.Conn, error) { return ln.Dial() },
	}
	ic := instafasthttp.GetInstrumentedClient(s, c)

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
	err := ic.DoDeadline(ctx, r, resp, time.Now().Add(time.Minute*10))

	require.NoError(t, err)

	parentSpan.Finish()

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 2)

	cSpan, pSpan := spans[0], spans[1]
	assert.Equal(t, 0, cSpan.Ec)
	assert.EqualValues(t, instana.ExitSpanKind, cSpan.Kind)

	assert.Equal(t, pSpan.TraceID, cSpan.TraceID)
	assert.Equal(t, pSpan.SpanID, cSpan.ParentID)

	assert.Equal(t, instana.FormatID(cSpan.TraceID), fieldTFrmHeader)
	assert.Equal(t, instana.FormatID(cSpan.SpanID), fieldSFrmHeader)

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

func TestClient_DoDeadline_Error(t *testing.T) {

	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder))

	parentSpan := s.Tracer().StartSpan("parent")
	ctx := instana.ContextWithSpan(context.Background(), parentSpan)

	server := &fasthttp.Server{
		Handler: func(ctx *fasthttp.RequestCtx) {
			ctx.Success("aaa/bbb", []byte("Ok response!"))
		},
	}

	ln := fasthttputil.NewInmemoryListener()

	go func() {
		if err := server.Serve(ln); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}()

	ln.Close()

	c := &fasthttp.Client{
		Dial: func(addr string) (net.Conn, error) { return ln.Dial() },
	}
	ic := instafasthttp.GetInstrumentedClient(s, c)

	r := &fasthttp.Request{}
	r.Header.SetMethod(fasthttp.MethodGet)
	r.Header.Set("Authorization", "Basic blah")
	r.URI().SetPath("/hello")
	r.URI().SetQueryString("q=term&key=s3cr3t")
	r.URI().SetHost("example.com")

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// Make the request
	err := ic.DoDeadline(ctx, r, resp, time.Now().Add(time.Minute*10))

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
		Error:  "InmemoryListener is already closed: use of closed network connection",
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
		Message: `error.object: "InmemoryListener is already closed: use of closed network connection"`,
	}, logData.Tags)
}

func TestClient_DoRedirects(t *testing.T) {
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

	var fieldTFrmHeader, fieldSFrmHeader string

	server := &fasthttp.Server{
		Handler: func(ctx *fasthttp.RequestCtx) {
			// get the header span and trace id from request header
			fieldTFrmHeader = string(ctx.Request.Header.Peek(instana.FieldT))
			fieldSFrmHeader = string(ctx.Request.Header.Peek(instana.FieldS))
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

	c := &fasthttp.Client{
		Dial: func(addr string) (net.Conn, error) { return ln.Dial() },
	}
	ic := instafasthttp.GetInstrumentedClient(s, c)

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
	err := ic.DoRedirects(ctx, r, resp, 2)

	require.NoError(t, err)

	parentSpan.Finish()

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 2)

	cSpan, pSpan := spans[0], spans[1]
	assert.Equal(t, 0, cSpan.Ec)
	assert.EqualValues(t, instana.ExitSpanKind, cSpan.Kind)

	assert.Equal(t, pSpan.TraceID, cSpan.TraceID)
	assert.Equal(t, pSpan.SpanID, cSpan.ParentID)

	assert.Equal(t, instana.FormatID(cSpan.TraceID), fieldTFrmHeader)
	assert.Equal(t, instana.FormatID(cSpan.SpanID), fieldSFrmHeader)

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

func TestClient_DoRedirects_Error(t *testing.T) {

	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder))

	parentSpan := s.Tracer().StartSpan("parent")
	ctx := instana.ContextWithSpan(context.Background(), parentSpan)

	server := &fasthttp.Server{
		Handler: func(ctx *fasthttp.RequestCtx) {
			ctx.Success("aaa/bbb", []byte("Ok response!"))
		},
	}

	ln := fasthttputil.NewInmemoryListener()

	go func() {
		if err := server.Serve(ln); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}()

	ln.Close()

	c := &fasthttp.Client{
		Dial: func(addr string) (net.Conn, error) { return ln.Dial() },
	}
	ic := instafasthttp.GetInstrumentedClient(s, c)

	r := &fasthttp.Request{}
	r.Header.SetMethod(fasthttp.MethodGet)
	r.Header.Set("Authorization", "Basic blah")
	r.URI().SetPath("/hello")
	r.URI().SetQueryString("q=term&key=s3cr3t")
	r.URI().SetHost("example.com")

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// Make the request
	err := ic.DoRedirects(ctx, r, resp, 2)

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
		Error:  "InmemoryListener is already closed: use of closed network connection",
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
		Message: `error.object: "InmemoryListener is already closed: use of closed network connection"`,
	}, logData.Tags)
}

func Test_Client_Unwrap(t *testing.T) {
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

	ln := fasthttputil.NewInmemoryListener()
	c := &fasthttp.Client{
		Dial: func(addr string) (net.Conn, error) { return ln.Dial() },
	}
	ic := instafasthttp.GetInstrumentedClient(s, c)

	org := ic.Unwrap()

	assert.IsType(t, &fasthttp.Client{}, org)
	assert.NotNil(t, org)
}
