// (c) Copyright IBM Corp. 2024

// Package instafasthttp provides Instana instrumentation for fasthttp package.
package instafasthttp

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"

	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/valyala/fasthttp"
)

// instanaUserContextKey defines the key name used to store context.Context in *fasthttp.RequestCtx
// as *fasthttp.RequestCtx does not provide any default option for storing context.Context.
// This can be utilised for trace propagation.
const instanaUserContextKey = "__instana_local_user_context__"

// UserContext returns a context implementation set by the user earlier.
// If not set, it returns a context.Context derived from *fasthttp.RequestCtx.
func UserContext(c *fasthttp.RequestCtx) context.Context {
	ctx, ok := c.UserValue(instanaUserContextKey).(context.Context)
	if !ok {
		// Since *fasthttp.RequestCtx satisfies context.Context,
		// we use it as the user context to ensure that all default values,
		// such as timeouts and other context-specific values, are retained.
		// This context can then be passed to other functions.
		ctx = c
		setUserContext(c, ctx)
	}

	return ctx
}

// SetUserContext sets a context implementation by user.
func setUserContext(c *fasthttp.RequestCtx, ctx context.Context) {
	c.SetUserValue(instanaUserContextKey, ctx)
}

// TraceHandler adds Instana instrumentation to the fasthttp.RequestHandler
func TraceHandler(sensor instana.TracerLogger, routeID, pathTemplate string, handler fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(c *fasthttp.RequestCtx) {
		var ctx context.Context = UserContext(c)

		req := &c.Request
		opts := initSpanOptionsFastHttp(req, routeID)

		tracer := sensor.Tracer()
		if ps, ok := instana.SpanFromContext(ctx); ok {
			tracer = ps.Tracer()
			opts = append(opts, ot.ChildOf(ps.Context()))
		}

		reqHeaders := collectAllHeaders(&req.Header)
		opts = append(opts, extractStartSpanOptionsFromHeadersFastHttp(tracer, req, reqHeaders, sensor)...)

		if string(req.Header.Peek(instana.FieldSynthetic)) == "1" {
			opts = append(opts, ot.Tag{Key: "synthetic_call", Value: true})
		}

		if pathTemplate != "" && string(req.URI().Path()) != pathTemplate {
			opts = append(opts, ot.Tag{Key: "http.path_tpl", Value: pathTemplate})
		}

		span := tracer.StartSpan("g.http", opts...)
		defer span.Finish()

		var params url.Values
		collectedHeaders := make(map[string]string)

		// ensure collected headers/params are sent in case of panic/error
		defer sendParamsAndHeaders(span, params, collectedHeaders)

		var collectableHTTPHeaders []string
		if t, ok := tracer.(instana.Tracer); ok {
			opts := t.Options()
			params = collectHTTPParamsFastHttp(req, opts.Secrets)
			collectableHTTPHeaders = opts.CollectableHTTPHeaders
		}

		collectHeadersFastHTTP(reqHeaders, collectableHTTPHeaders, collectedHeaders)

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

				span.SetTag(string(ext.HTTPStatusCode), fasthttp.StatusInternalServerError)

				// re-throw the panic
				panic(err)
			}
		}()

		// Inject the span details to the headers
		h := make(ot.HTTPHeadersCarrier)
		tracer.Inject(span.Context(), ot.HTTPHeaders, h)
		for k, v := range h {
			c.Response.Header.Del(k)
			c.Response.Header.Set(k, strings.Join(v, ","))
		}

		// setting context with span information for span propagation
		setUserContext(c, instana.ContextWithSpan(ctx, span))
		handler(c)

		resHeaders := collectAllHeaders(&c.Response.Header)
		collectHeadersFastHTTP(resHeaders, collectableHTTPHeaders, collectedHeaders)
		processResponseStatusFasthttp(&c.Response, span)
	}
}

func initSpanOptionsFastHttp(req *fasthttp.Request, routeID string) []ot.StartSpanOption {
	opts := []ot.StartSpanOption{
		ext.SpanKindRPCServer,
		ot.Tags{
			"http.host":     string(req.Host()),
			"http.method":   string(req.Header.Method()),
			"http.protocol": string(req.URI().Scheme()),
			"http.path":     string(req.URI().Path()),
			"http.route_id": routeID,
		},
	}
	return opts
}

// interface for req and res headers
// used to collect headers
type headerVisiter interface {
	VisitAll(f func(key, value []byte))
}

func collectAllHeaders(header headerVisiter) http.Header {
	headers := make(http.Header, 0)

	header.VisitAll(func(key, value []byte) {
		headerKey := make([]byte, len(key))
		copy(headerKey, key)

		headerVal := make([]byte, len(value))
		copy(headerVal, value)

		headers.Add(string(headerKey), string(headerVal))
	})

	return headers
}

func processResponseStatusFasthttp(response *fasthttp.Response, span ot.Span) {
	stCode := response.StatusCode()
	if stCode > 0 {
		if stCode >= fasthttp.StatusInternalServerError {
			statusText := fasthttp.StatusMessage(stCode)

			span.SetTag("http.error", statusText)
			span.LogFields(otlog.Object("error", statusText))
		}

		span.SetTag("http.status", stCode)
	}
}

func collectHeadersFastHTTP(headers http.Header, collectableHTTPHeaders []string, collectedHeaders map[string]string) {
	for _, h := range collectableHTTPHeaders {
		if v := headers.Get(h); v != "" {
			collectedHeaders[h] = v
		}
	}
}

func extractStartSpanOptionsFromHeadersFastHttp(tracer ot.Tracer, req *fasthttp.Request, headers map[string][]string, sensor instana.TracerLogger) []ot.StartSpanOption {
	var opts []ot.StartSpanOption
	wireContext, err := tracer.Extract(ot.HTTPHeaders, ot.HTTPHeadersCarrier(headers))
	switch {
	case err == nil:
		opts = append(opts, ext.RPCServerOption(wireContext))
	case errors.Is(err, ot.ErrSpanContextNotFound):
		sensor.Logger().Debug("no span context provided with ", string(req.Header.Method()), " ", string(req.URI().Path()))
	case errors.Is(err, ot.ErrUnsupportedFormat):
		sensor.Logger().Info("unsupported span context format provided with ", string(req.Header.Method()), " ", string(req.URI().Path()))
	default:
		sensor.Logger().Warn("failed to extract span context from the request:", err)
	}
	return opts
}

func collectHTTPParamsFastHttp(req *fasthttp.Request, matcher instana.Matcher) url.Values {
	params, _ := url.ParseQuery(string(req.URI().QueryString()))

	for k := range params {
		if matcher.Match(k) {
			params[k] = []string{"<redacted>"}
		}
	}

	return params
}

func sendParamsAndHeaders(span ot.Span, params url.Values, collectedHeaders map[string]string) {
	if len(collectedHeaders) > 0 {
		span.SetTag("http.header", collectedHeaders)
	}
	if len(params) > 0 {
		span.SetTag("http.params", params.Encode())
	}
}

type tracingRoundTripper func(*fasthttp.HostClient, *fasthttp.Request, *fasthttp.Response) (bool, error)

func (rt tracingRoundTripper) RoundTrip(hc *fasthttp.HostClient, req *fasthttp.Request, resp *fasthttp.Response) (retry bool, err error) {
	return rt(hc, req, resp)
}

// RoundTripper wraps an existing fasthttp.RoundTripper and injects the tracing headers into the outgoing request.
// If the original RoundTripper is nil, the fasthttp.DefaultTransport will be used.
func RoundTripper(ctx context.Context, sensor instana.TracerLogger, original fasthttp.RoundTripper) fasthttp.RoundTripper {
	if ctx == nil {
		ctx = context.Background()
	}
	if original == nil {
		original = fasthttp.DefaultTransport
	}
	return tracingRoundTripper(func(hc *fasthttp.HostClient, req *fasthttp.Request, resp *fasthttp.Response) (bool, error) {
		sanitizedURL := new(fasthttp.URI)
		req.URI().CopyTo(sanitizedURL)
		sanitizedURL.SetUsername("")
		sanitizedURL.SetPassword("")
		sanitizedURL.SetQueryString("")

		opts := []ot.StartSpanOption{
			ext.SpanKindRPCClient,
			ot.Tags{
				"http.url":    sanitizedURL.String(),
				"http.method": req.Header.Method(),
			},
		}

		tracer := sensor.Tracer()
		parentSpan, ok := instana.SpanFromContext(ctx)
		if ok {
			tracer = parentSpan.Tracer()
			opts = append(opts, ot.ChildOf(parentSpan.Context()))
		}

		span := tracer.StartSpan("http", opts...)
		defer span.Finish()

		// clone the request since the RoundTrip should not modify the original one
		reqClone := &fasthttp.Request{}
		req.CopyTo(reqClone)

		// Inject the span details to the headers
		h := make(ot.HTTPHeadersCarrier)
		tracer.Inject(span.Context(), ot.HTTPHeaders, h)
		for k, v := range h {
			reqClone.Header.Del(k)
			reqClone.Header.Set(k, strings.Join(v, ","))
		}

		var params url.Values
		collectedHeaders := make(map[string]string)

		// ensure collected headers/params are sent in case of panic/error
		defer sendParamsAndHeaders(span, params, collectedHeaders)

		var collectableHTTPHeaders []string
		if t, ok := tracer.(instana.Tracer); ok {
			opts := t.Options()
			params = collectHTTPParamsFastHttp(req, opts.Secrets)
			collectableHTTPHeaders = opts.CollectableHTTPHeaders
		}

		reqHeaders := collectAllHeaders(&req.Header)
		collectHeadersFastHTTP(reqHeaders, collectableHTTPHeaders, collectedHeaders)

		retry, err := original.RoundTrip(hc, reqClone, resp)
		if err != nil {
			span.SetTag("http.error", err.Error())
			span.LogFields(otlog.Error(err))
			return retry, err
		}

		resHeaders := collectAllHeaders(&resp.Header)
		collectHeadersFastHTTP(resHeaders, collectableHTTPHeaders, collectedHeaders)

		span.SetTag(string(ext.HTTPStatusCode), resp.StatusCode())

		return retry, err
	})
}
