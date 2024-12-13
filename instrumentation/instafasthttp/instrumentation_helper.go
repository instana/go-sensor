// (c) Copyright IBM Corp. 2024

package instafasthttp

import (
	"context"
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

type clientFuncType int

const (
	doFunc clientFuncType = iota
	doWithTimeoutFunc
	doWithDeadlineFunc
	doWithRedirectsFunc

	doRoundTripFunc
)

type clientFuncParams struct {
	sensor instana.TracerLogger

	hc *fasthttp.HostClient
	rt fasthttp.RoundTripper

	ic                *instaClient
	clientFuncType    clientFuncType
	timeout           time.Duration
	deadline          time.Time
	maxRedirectsCount int
}

func instrumentClient(ctx context.Context, req *fasthttp.Request, resp *fasthttp.Response, cfp *clientFuncParams) (bool, error) {
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

	tracer := cfp.sensor.Tracer()
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

	var collectableHTTPHeaders []string
	if t, ok := tracer.(instana.Tracer); ok {
		opts := t.Options()
		params = collectHTTPParamsFastHttp(req, opts.Secrets)
		collectableHTTPHeaders = opts.CollectableHTTPHeaders
	}

	collectedHeaders := make(map[string]string, len(collectableHTTPHeaders))

	// ensure collected headers/params are sent in case of panic/error
	defer setHeadersAndParamsToSpan(span, collectedHeaders, params)

	reqHeaders := collectAllHeaders(&req.Header)
	collectHeadersFastHTTP(reqHeaders, collectableHTTPHeaders, collectedHeaders)

	var err error
	var retry bool

	switch cfp.clientFuncType {
	case doWithRedirectsFunc:
		err = cfp.ic.Client.DoRedirects(reqClone, resp, cfp.maxRedirectsCount)
	case doWithDeadlineFunc:
		err = cfp.ic.Client.DoDeadline(reqClone, resp, cfp.deadline)
	case doWithTimeoutFunc:
		err = cfp.ic.Client.DoTimeout(reqClone, resp, cfp.timeout)
	case doFunc:
		err = cfp.ic.Client.Do(reqClone, resp)
	case doRoundTripFunc:
		retry, err = cfp.rt.RoundTrip(cfp.hc, reqClone, resp)
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

func collectHeadersFastHTTP(headers http.Header, collectableHTTPHeaders []string, collectedHeaders map[string]string) {
	for _, h := range collectableHTTPHeaders {
		if v := headers.Get(h); v != "" {
			collectedHeaders[h] = v
		}
	}
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
