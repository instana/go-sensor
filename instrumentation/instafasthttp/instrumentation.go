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

// instanaUserContextKey define the key name for storing context.Context in *fasthttp.RequestCtx
// as *fasthttp.RequestCtx doesnt give any default option to store context.Context
// can be used for trace propagation
const instanaUserContextKey = "__instana_local_user_context__"

func TraceHandler(sensor instana.TracerLogger, pathTemplate string, handler fasthttp.RequestHandler) fasthttp.RequestHandler {
	return tracingNamedHandlerFuncFastHttp(sensor, "", pathTemplate, handler)
}

func tracingNamedHandlerFuncFastHttp(sensor instana.TracerLogger, routeID, pathTemplate string, handler fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(c *fasthttp.RequestCtx) {
		var ctx context.Context = UserContext(c)
		req := &c.Request

		opts := initSpanOptionsFastHttp(req, routeID)

		tracer := sensor.Tracer()
		if ps, ok := instana.SpanFromContext(ctx); ok {
			tracer = ps.Tracer()
			opts = append(opts, ot.ChildOf(ps.Context()))
		}

		headers := collectAllHeaders(req)

		opts = append(opts, extractStartSpanOptionsFromHeadersFastHttp(tracer, req, headers, sensor)...)

		if string(req.Header.Peek(instana.FieldSynthetic)) == "1" {
			opts = append(opts, ot.Tag{Key: instana.FieldSynthetic, Value: true})
		}

		if pathTemplate != "" && string(req.URI().Path()) != pathTemplate {
			opts = append(opts, ot.Tag{Key: "http.path_tpl", Value: pathTemplate})
		}

		span := tracer.StartSpan("g.http", opts...)
		defer span.Finish()

		var params url.Values
		collectedHeaders := make(map[string]string)

		// ensure collected headers/params are sent in case of panic/error
		defer func() {
			if len(collectedHeaders) > 0 {
				span.SetTag("http.header", collectedHeaders)
			}
			if len(params) > 0 {
				span.SetTag("http.params", params.Encode())
			}
		}()

		var collectableHTTPHeaders []string
		if t, ok := tracer.(instana.Tracer); ok {
			opts := t.Options()
			params = collectHTTPParamsFastHttp(req, opts.Secrets)
			collectableHTTPHeaders = opts.CollectableHTTPHeaders
		}

		collectRequestHeadersFastHTTP(headers, collectableHTTPHeaders, collectedHeaders)

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

		setUserContext(c, instana.ContextWithSpan(ctx, span))
		handler(c)

		collectResponseHeadersFasthttp(&c.Response, collectableHTTPHeaders, collectedHeaders)
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

func collectAllHeaders(req *fasthttp.Request) http.Header {
	headers := make(http.Header, 0)

	req.Header.VisitAll(func(key, value []byte) {
		headerKey := make([]byte, len(key))
		copy(headerKey, key)

		headerVal := make([]byte, len(value))
		copy(headerVal, value)

		headers.Add(string(headerKey), string(headerVal))
	})

	return headers
}

// UserContext returns a context implementation that was set by
// user earlier or returns a non-nil, empty context,if it was not set earlier.
func UserContext(c *fasthttp.RequestCtx) context.Context {
	ctx, ok := c.UserValue(instanaUserContextKey).(context.Context)
	if !ok {
		// as *fasthttp.RequestCtx satisfies context.Context
		// we are taking the same as the user context so that all the
		// default values like timeout and other context specific values
		// will be copied and this context can be used to pass to other functions.
		ctx = c
		setUserContext(c, ctx)
	}

	return ctx
}

// SetUserContext sets a context implementation by user.
func setUserContext(c *fasthttp.RequestCtx, ctx context.Context) {
	c.SetUserValue(instanaUserContextKey, ctx)
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

func collectResponseHeadersFasthttp(response *fasthttp.Response, collectableHTTPHeaders []string, collectedHeaders map[string]string) {
	for _, h := range collectableHTTPHeaders {

		if value := response.Header.Peek(h); value != nil {
			headerCopy := make([]byte, len(value))
			copy(headerCopy, value)
			collectedHeaders[h] = string(headerCopy)
		}
	}

}

func collectRequestHeadersFastHTTP(headers http.Header, collectableHTTPHeaders []string, collectedHeaders map[string]string) {
	for _, h := range collectableHTTPHeaders {
		if v := headers.Get(h); v != "" {
			collectedHeaders[h] = v
		}
	}
}

func extractStartSpanOptionsFromHeadersFastHttp(tracer ot.Tracer, req *fasthttp.Request, headers map[string][]string, sensor TracerLogger) []ot.StartSpanOption {
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
