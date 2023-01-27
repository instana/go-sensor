// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instana

import (
	"bufio"
	"context"
	"mime/multipart"
	"net"
	"net/http"
	"net/textproto"
	"net/url"

	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
)

// TracingHandlerFunc is an HTTP middleware that captures the tracing data and ensures
// trace context propagation via OpenTracing headers. The pathTemplate parameter, when provided,
// will be added to the span as a template string used to match the route containing variables, regular
// expressions, etc.
//
// The wrapped handler will also propagate the W3C trace context (https://www.w3.org/TR/trace-context/)
// if found in request.
func TracingHandlerFunc(sensor *Sensor, pathTemplate string, handler http.HandlerFunc) http.HandlerFunc {
	return TracingNamedHandlerFunc(sensor, "", pathTemplate, handler)
}

// TracingNamedHandlerFunc is an HTTP middleware that similarly to instana.TracingHandlerFunc() captures the tracing data,
// while allowing to provide a unique route indetifier to be associated with each request.
func TracingNamedHandlerFunc(sensor *Sensor, routeID, pathTemplate string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		opts := initSpanOptions(req, routeID)

		tracer := sensor.Tracer()
		if ps, ok := SpanFromContext(req.Context()); ok {
			tracer = ps.Tracer()
			opts = append(opts, ot.ChildOf(ps.Context()))
		}

		opts = append(opts, extractStartSpanOptionsFromHeaders(tracer, req, sensor)...)

		if req.Header.Get(FieldSynthetic) == "1" {
			opts = append(opts, syntheticCall())
		}

		if pathTemplate != "" && req.URL.Path != pathTemplate {
			opts = append(opts, ot.Tag{Key: "http.path_tpl", Value: pathTemplate})
		}

		span := tracer.StartSpan("g.http", opts...)
		defer span.Finish()

		var collectableHTTPHeaders []string
		if t, ok := tracer.(Tracer); ok {
			opts := t.Options()
			collectableHTTPHeaders = opts.CollectableHTTPHeaders

			params := collectHTTPParams(req, opts.Secrets)
			if len(params) > 0 {
				span.SetTag("http.params", params.Encode())
			}
		}

		collectedHeaders := make(map[string]string)
		// make sure collected headers are sent in case of panic/error
		defer func() {
			if len(collectedHeaders) > 0 {
				span.SetTag("http.header", collectedHeaders)
			}
		}()

		collectRequestHeaders(req, collectableHTTPHeaders, collectedHeaders)

		defer func() {
			// Be sure to capture any kind of panic/error
			if err := recover(); err != nil {
				if e, ok := err.(error); ok {
					span.SetTag("http.error", e.Error())
					span.LogFields(otlog.Error(e))
				} else {
					span.SetTag("http.error", err)
					span.LogFields(otlog.Object("error", err))
				}

				span.SetTag(string(ext.HTTPStatusCode), http.StatusInternalServerError)

				// re-throw the panic
				panic(err)
			}
		}()

		wrapped := wrapResponseWriter(w)
		tracer.Inject(span.Context(), ot.HTTPHeaders, ot.HTTPHeadersCarrier(wrapped.Header()))

		handler(wrapped, req.WithContext(ContextWithSpan(ctx, span)))

		collectResponseHeaders(wrapped, collectableHTTPHeaders, collectedHeaders)
		processResponseStatus(wrapped, span)
	}
}

func initSpanOptions(req *http.Request, routeID string) []ot.StartSpanOption {
	opts := []ot.StartSpanOption{
		ext.SpanKindRPCServer,
		ot.Tags{
			"http.host":     req.Host,
			"http.method":   req.Method,
			"http.protocol": req.URL.Scheme,
			"http.path":     req.URL.Path,
			"http.route_id": routeID,
		},
	}
	return opts
}

func processResponseStatus(response wrappedResponseWriter, span ot.Span) {
	if response.Status() > 0 {
		if response.Status() >= http.StatusInternalServerError {
			statusText := http.StatusText(response.Status())

			span.SetTag("http.error", statusText)
			span.LogFields(otlog.Object("error", statusText))
		}

		// GraphQL queries received by a webserver are HTTP incoming requests that must be "merged" into one single entry
		// span, as we cannot have two entry spans (http and graphql).
		// Our UI will render a span as HTTP if the http.status tag is present, so in the case of a GraphQL span, we must
		// ensure that this status is not provided.
		// Due to a design limitation, the graphql instrumentation doesn't have access to the instana.spanS struct. Plus,
		// we are unable to control the span finishing from the graphql instrumentation, which leads us to add this workaround
		// here.
		if spS, ok := span.(*spanS); ok {
			if spS.Operation != "graphql.server" {
				span.SetTag("http.status", response.Status())
			}
		}
	}
}

func collectResponseHeaders(response wrappedResponseWriter, collectableHTTPHeaders []string, collectedHeaders map[string]string) {
	for _, h := range collectableHTTPHeaders {
		if v := response.Header().Get(h); v != "" {
			collectedHeaders[h] = v
		}
	}
}

func collectRequestHeaders(req *http.Request, collectableHTTPHeaders []string, collectedHeaders map[string]string) {
	for _, h := range collectableHTTPHeaders {
		if v := req.Header.Get(h); v != "" {
			collectedHeaders[h] = v
		}
	}
}

func extractStartSpanOptionsFromHeaders(tracer ot.Tracer, req *http.Request, sensor *Sensor) []ot.StartSpanOption {
	var opts []ot.StartSpanOption
	wireContext, err := tracer.Extract(ot.HTTPHeaders, ot.HTTPHeadersCarrier(req.Header))
	switch err {
	case nil:
		opts = append(opts, ext.RPCServerOption(wireContext))
	case ot.ErrSpanContextNotFound:
		sensor.Logger().Debug("no span context provided with ", req.Method, " ", req.URL.Path)
	case ot.ErrUnsupportedFormat:
		sensor.Logger().Info("unsupported span context format provided with ", req.Method, " ", req.URL.Path)
	default:
		sensor.Logger().Warn("failed to extract span context from the request:", err)
	}
	return opts
}

// RoundTripper wraps an existing http.RoundTripper and injects the tracing headers into the outgoing request.
// If the original RoundTripper is nil, the http.DefaultTransport will be used.
func RoundTripper(sensor *Sensor, original http.RoundTripper) http.RoundTripper {
	return tracingRoundTripper(func(req *http.Request) (*http.Response, error) {
		if original == nil {
			original = http.DefaultTransport
		}

		ctx := req.Context()
		parentSpan, ok := SpanFromContext(ctx)
		if !ok {
			// don't trace the exit call if there was no entry span provided
			return original.RoundTrip(req)
		}

		sanitizedURL := cloneURL(req.URL)
		sanitizedURL.RawQuery = ""
		sanitizedURL.User = nil

		span := sensor.Tracer().StartSpan("http",
			ext.SpanKindRPCClient,
			ot.ChildOf(parentSpan.Context()),
			ot.Tags{
				"http.url":    sanitizedURL.String(),
				"http.method": req.Method,
			})
		defer span.Finish()

		// clone the request since the RoundTrip should not modify the original one
		req = cloneRequest(ContextWithSpan(ctx, span), req)
		sensor.Tracer().Inject(span.Context(), ot.HTTPHeaders, ot.HTTPHeadersCarrier(req.Header))

		var collectableHTTPHeaders []string
		if t, ok := sensor.Tracer().(Tracer); ok {
			opts := t.Options()
			collectableHTTPHeaders = opts.CollectableHTTPHeaders

			params := collectHTTPParams(req, opts.Secrets)
			if len(params) > 0 {
				span.SetTag("http.params", params.Encode())
			}
		}

		collectedHeaders := make(map[string]string)
		// make sure collected headers are sent in case of panic/error
		defer func() {
			if len(collectedHeaders) > 0 {
				span.SetTag("http.header", collectedHeaders)
			}
		}()

		// collect request headers
		for _, h := range collectableHTTPHeaders {
			if v := req.Header.Get(h); v != "" {
				collectedHeaders[h] = v
			}
		}

		resp, err := original.RoundTrip(req)
		if err != nil {
			span.SetTag("http.error", err.Error())
			span.LogFields(otlog.Error(err))
			return resp, err
		}

		// collect response headers
		for _, h := range collectableHTTPHeaders {
			if v := resp.Header.Get(h); v != "" {
				collectedHeaders[h] = v
			}
		}

		span.SetTag(string(ext.HTTPStatusCode), resp.StatusCode)

		return resp, err
	})
}

type wrappedResponseWriter interface {
	http.ResponseWriter
	Status() int
}

func wrapResponseWriter(w http.ResponseWriter) wrappedResponseWriter {
	if _, ok := w.(http.Hijacker); ok {
		return &statusCodeRecorderHTTP10{
			ResponseWriter: w,
		}
	}

	return &statusCodeRecorder{
		ResponseWriter: w,
	}
}

// statusCodeRecorder is a wrapper over http.ResponseWriter to spy the returned status code
type statusCodeRecorder struct {
	http.ResponseWriter
	status int
}

func (rec *statusCodeRecorder) SetStatus(status int) {
	rec.status = status
}

func (rec *statusCodeRecorder) WriteHeader(status int) {
	rec.SetStatus(status)
	rec.ResponseWriter.WriteHeader(status)
}

func (rec *statusCodeRecorder) Write(b []byte) (int, error) {
	if rec.status == 0 {
		rec.SetStatus(http.StatusOK)
	}

	return rec.ResponseWriter.Write(b)
}

func (rec *statusCodeRecorder) Status() int {
	return rec.status
}

// statusCodeRecorderHTTP10 is a wrapper over http.ResponseWriter similar to statusCodeRecorder, but
// also implementing http.Hijaker
type statusCodeRecorderHTTP10 = statusCodeRecorder

func (rec *statusCodeRecorderHTTP10) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return rec.ResponseWriter.(http.Hijacker).Hijack()
}

type tracingRoundTripper func(*http.Request) (*http.Response, error)

func (rt tracingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return rt(req)
}

func collectHTTPParams(req *http.Request, matcher Matcher) url.Values {
	params := cloneURLValues(req.URL.Query())

	for k := range params {
		if matcher.Match(k) {
			params[k] = []string{"<redacted>"}
		}
	}

	return params
}

// The following code is ported from $GOROOT/src/net/http/clone.go with minor changes
// for compatibility with Go versions prior to 1.13
//
// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

func cloneRequest(ctx context.Context, r *http.Request) *http.Request {
	r2 := new(http.Request)
	*r2 = *r
	r2 = r2.WithContext(ctx)

	r2.URL = cloneURL(r.URL)
	if r.Header != nil {
		r2.Header = cloneHeader(r.Header)
	}

	if r.Trailer != nil {
		r2.Trailer = cloneHeader(r.Trailer)
	}

	if s := r.TransferEncoding; s != nil {
		s2 := make([]string, len(s))
		copy(s2, s)
		r2.TransferEncoding = s
	}

	r2.Form = cloneURLValues(r.Form)
	r2.PostForm = cloneURLValues(r.PostForm)
	r2.MultipartForm = cloneMultipartForm(r.MultipartForm)

	return r2
}

func cloneURLValues(v url.Values) url.Values {
	if v == nil {
		return nil
	}

	// http.Header and url.Values have the same representation, so temporarily
	// treat it like http.Header, which does have a clone:

	return url.Values(cloneHeader(http.Header(v)))
}

func cloneURL(u *url.URL) *url.URL {
	if u == nil {
		return nil
	}

	u2 := new(url.URL)
	*u2 = *u

	if u.User != nil {
		u2.User = new(url.Userinfo)
		*u2.User = *u.User
	}

	return u2
}

func cloneMultipartForm(f *multipart.Form) *multipart.Form {
	if f == nil {
		return nil
	}

	f2 := &multipart.Form{
		Value: (map[string][]string)(cloneHeader(http.Header(f.Value))),
	}

	if f.File != nil {
		m := make(map[string][]*multipart.FileHeader)
		for k, vv := range f.File {
			vv2 := make([]*multipart.FileHeader, len(vv))
			for i, v := range vv {
				vv2[i] = cloneMultipartFileHeader(v)
			}
			m[k] = vv2

		}

		f2.File = m
	}

	return f2
}

func cloneMultipartFileHeader(fh *multipart.FileHeader) *multipart.FileHeader {
	if fh == nil {
		return nil
	}

	fh2 := new(multipart.FileHeader)
	*fh2 = *fh

	fh2.Header = textproto.MIMEHeader(cloneHeader(http.Header(fh.Header)))

	return fh2
}

// The following code is ported from $GOROOT/src/net/http/header.go with minor changes
// for compatibility with Go versions prior to 1.13
//
// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

func cloneHeader(h http.Header) http.Header {
	if h == nil {
		return nil
	}

	// Find total number of values.
	nv := 0
	for _, vv := range h {
		nv += len(vv)
	}
	sv := make([]string, nv) // shared backing array for headers' values
	h2 := make(http.Header, len(h))
	for k, vv := range h {
		n := copy(sv, vv)
		h2[k] = sv[:n:n]
		sv = sv[n:]
	}
	return h2
}
