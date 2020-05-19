package instana

import (
	"context"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"

	"github.com/instana/go-sensor/w3ctrace"
	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
)

// TracingHandlerFunc is an HTTP middleware that captures the tracing data and ensures
// trace context propagation via OpenTracing headers. The wrapped handler will also propagate
// the W3C trace context (https://www.w3.org/TR/trace-context/) if found in request
func TracingHandlerFunc(sensor *Sensor, name string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		opts := []ot.StartSpanOption{
			ext.SpanKindRPCServer,
			ot.Tags{
				"http.host":     req.Host,
				"http.method":   req.Method,
				"http.protocol": req.URL.Scheme,
				"http.path":     req.URL.Path,
			},
		}

		tracer := sensor.Tracer()
		if ps, ok := SpanFromContext(req.Context()); ok {
			tracer = ps.Tracer()
			opts = append(opts, ot.ChildOf(ps.Context()))
		}

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

		if req.Header.Get(FieldSynthetic) == "1" {
			opts = append(opts, syntheticCall())
		}

		span := tracer.StartSpan("g.http", opts...)
		defer span.Finish()

		defer func() {
			// Be sure to capture any kind of panic / error
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

		wrapped := &statusCodeRecorder{ResponseWriter: w}
		tracer.Inject(span.Context(), ot.HTTPHeaders, ot.HTTPHeadersCarrier(wrapped.Header()))

		ctx = ContextWithSpan(ctx, span)
		w3ctrace.TracingHandlerFunc(handler)(wrapped, req.WithContext(ctx))

		if wrapped.Status > 0 {
			if wrapped.Status > http.StatusInternalServerError {
				span.SetTag("http.error", http.StatusText(wrapped.Status))
			}

			span.SetTag("http.status", wrapped.Status)
		}
	}
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

		resp, err := original.RoundTrip(req)
		if err != nil {
			span.SetTag("http.error", err.Error())
			span.LogFields(otlog.Error(err))
			return resp, err
		}

		span.SetTag(string(ext.HTTPStatusCode), resp.StatusCode)

		return resp, err
	})
}

// wrapper over http.ResponseWriter to spy the returned status code
type statusCodeRecorder struct {
	http.ResponseWriter
	Status int
}

func (rec *statusCodeRecorder) WriteHeader(status int) {
	rec.Status = status
	rec.ResponseWriter.WriteHeader(status)
}

func (rec *statusCodeRecorder) Write(b []byte) (int, error) {
	if rec.Status == 0 {
		rec.Status = http.StatusOK
	}

	return rec.ResponseWriter.Write(b)
}

type tracingRoundTripper func(*http.Request) (*http.Response, error)

func (rt tracingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return rt(req)
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
