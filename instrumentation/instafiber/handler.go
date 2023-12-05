// (c) Copyright IBM Corp. 2023

//go:build go1.17
// +build go1.17

// Package instafiber provides Instana instrumentation for Fiber package.
package instafiber

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/valyala/fasthttp"
)

// TraceHandler adds Instana instrumentation to the route handler
func TraceHandler(sensor instana.TracerLogger, routeID, pathTemplate string, handler fiber.Handler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.UserContext()
		req := c.Request() // This is a fasthttp request and not a net/http request

		headers := collectAllHeaders(req)

		opts := initSpanOptions(req, routeID)

		tracer := sensor.Tracer()
		if ps, ok := instana.SpanFromContext(ctx); ok {
			tracer = ps.Tracer()
			opts = append(opts, ot.ChildOf(ps.Context()))
		}

		opts = append(opts, extractStartSpanOptionsFromFastHTTPHeaders(tracer, req, headers, sensor)...)

		if isSynthetic(req) {
			opts = append(opts, ot.Tag{Key: "synthetic_call", Value: true})
		}

		if isCustomPathTemplate(req, pathTemplate) {
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

		params = collectHTTPParams(req, tracer)

		collectableHTTPHeaders := configuredCollectableHeaders(tracer)
		collectRequestHeaders(headers, collectableHTTPHeaders, collectedHeaders)

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

		// Inject the span details to the headers
		h := make(ot.HTTPHeadersCarrier)
		tracer.Inject(span.Context(), ot.HTTPHeaders, h)
		for k, v := range h {
			c.Response().Header.Del(k)
			c.Set(k, strings.Join(v, ","))
		}

		c.SetUserContext(instana.ContextWithSpan(ctx, span))
		err := handler(c)

		collectResponseHeaders(c.Response(), collectableHTTPHeaders, collectedHeaders)
		processResponseStatus(c.Response().StatusCode(), span)

		return err
	}
}

func collectRequestHeaders(headers http.Header, collectableHTTPHeaders []string, collectedHeaders map[string]string) {
	for _, h := range collectableHTTPHeaders {

		if v := headers.Get(h); v != "" {
			collectedHeaders[h] = v
		}
	}
}

func extractStartSpanOptionsFromFastHTTPHeaders(tracer ot.Tracer,
	req *fasthttp.Request,
	headers map[string][]string,
	sensor instana.TracerLogger) []ot.StartSpanOption {
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

func collectHTTPParams(req *fasthttp.Request, tracer ot.Tracer) url.Values {
	var params url.Values

	if t, ok := tracer.(instana.Tracer); ok {
		opts := t.Options()
		params, _ = url.ParseQuery(string(req.URI().QueryString()))

		matcher := opts.Secrets
		for k := range params {
			if matcher.Match(k) {
				params[k] = []string{"<redacted>"}
			}
		}
	}

	return params
}

func collectResponseHeaders(response *fasthttp.Response, collectableHTTPHeaders []string, collectedHeaders map[string]string) {
	for _, h := range collectableHTTPHeaders {

		if value := response.Header.Peek(h); value != nil {
			headerCopy := make([]byte, len(value))
			copy(headerCopy, value)
			collectedHeaders[h] = string(headerCopy)
		}
	}

}

func processResponseStatus(statusCode int, span ot.Span) {
	if statusCode > 0 {
		if statusCode >= http.StatusInternalServerError {
			statusText := http.StatusText(statusCode)

			span.SetTag("http.error", statusText)
			span.LogFields(otlog.Object("error", statusText))
		}
		span.SetTag("http.status", statusCode)
	}
}

func initSpanOptions(req *fasthttp.Request, routeID string) []ot.StartSpanOption {
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

func isSynthetic(req *fasthttp.Request) bool { return nil != req.Header.Peek(instana.FieldSynthetic) }

func isCustomPathTemplate(req *fasthttp.Request, pathTemplate string) bool {
	return pathTemplate != "" && string(req.URI().Path()) != pathTemplate
}

func collectAllHeaders(req *fiber.Request) http.Header {
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

func configuredCollectableHeaders(tracer ot.Tracer) []string {
	var collectableHTTPHeaders []string
	if t, ok := tracer.(instana.Tracer); ok {
		opts := t.Options()
		collectableHTTPHeaders = opts.CollectableHTTPHeaders
	}

	return collectableHTTPHeaders
}
