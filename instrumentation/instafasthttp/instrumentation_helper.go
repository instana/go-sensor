// (c) Copyright IBM Corp. 2024

package instafasthttp

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/valyala/fasthttp"
)

type doType int

const (
	doFunc doType = iota
	doFuncWithTimeout
	doFuncWithDeadline
	doFuncWithRedirects

	doRoundTrip
)

type doParams struct {
	sensor instana.TracerLogger

	hc *fasthttp.HostClient
	rt fasthttp.RoundTripper

	ic                *instaClient
	doType            doType
	timeout           time.Duration
	deadline          time.Time
	maxRedirectsCount int
}

func instrumentedDo(ctx context.Context, req *fasthttp.Request, resp *fasthttp.Response, dp *doParams) (bool, error) {
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

	tracer := dp.sensor.Tracer()
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
	var collectableHTTPHeaders []string
	if t, ok := tracer.(instana.Tracer); ok {
		opts := t.Options()
		params = collectHTTPParamsFastHttp(req, opts.Secrets)
		collectableHTTPHeaders = opts.CollectableHTTPHeaders
	}

	// ensure collected headers/params are sent in case of panic/error
	defer setHeadersAndParamsToSpan(span, collectedHeaders, params)

	reqHeaders := collectAllHeaders(&req.Header)
	collectHeadersFastHTTP(reqHeaders, collectableHTTPHeaders, collectedHeaders)

	var err error
	var retry bool

	switch dp.doType {
	case doFuncWithRedirects:
		err = dp.ic.Client.DoRedirects(reqClone, resp, dp.maxRedirectsCount)
	case doFuncWithDeadline:
		err = dp.ic.Client.DoDeadline(reqClone, resp, dp.deadline)
	case doFuncWithTimeout:
		err = dp.ic.Client.DoTimeout(reqClone, resp, dp.timeout)
	case doFunc:
		err = dp.ic.Client.Do(reqClone, resp)
	case doRoundTrip:
		retry, err = dp.rt.RoundTrip(dp.hc, reqClone, resp)
	}

	if err != nil {
		span.SetTag("http.error", err.Error())
		span.LogFields(otlog.Error(err))
		return retry, err
	}

	resHeaders := collectAllHeaders(&resp.Header)
	collectHeadersFastHTTP(resHeaders, collectableHTTPHeaders, collectedHeaders)

	span.SetTag(string(ext.HTTPStatusCode), resp.StatusCode())

	return retry, err
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

func setHeadersAndParamsToSpan(span ot.Span, headers map[string]string, params url.Values) {
	if len(headers) > 0 {
		span.SetTag("http.header", headers)
	}
	if len(params) > 0 {
		span.SetTag("http.params", params.Encode())
	}
}
