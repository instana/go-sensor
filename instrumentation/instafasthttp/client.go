// (c) Copyright IBM Corp. 2024

package instafasthttp

import (
	"context"
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

func GetClient(sensor instana.TracerLogger, orgClient *fasthttp.Client) Client {
	return &instaClient{
		Client: orgClient,
		sensor: sensor,
	}
}

type Client interface {
	// methods from original *fasthttp.Client
	// no need to implement this
	Get(dst []byte, url string) (statusCode int, body []byte, err error)
	GetTimeout(dst []byte, url string, timeout time.Duration) (statusCode int, body []byte, err error)
	GetDeadline(dst []byte, url string, deadline time.Time) (statusCode int, body []byte, err error)
	Post(dst []byte, url string, postArgs *fasthttp.Args) (statusCode int, body []byte, err error)
	CloseIdleConnections()

	// new methods
	// used by instana instrumentation
	DoTimeout(ctx context.Context, req *fasthttp.Request, resp *fasthttp.Response, timeout time.Duration) error
	DoDeadline(ctx context.Context, req *fasthttp.Request, resp *fasthttp.Response, deadline time.Time) error
	DoRedirects(ctx context.Context, req *fasthttp.Request, resp *fasthttp.Response, maxRedirectsCount int) error
	Do(ctx context.Context, req *fasthttp.Request, resp *fasthttp.Response) error

	// new method
	// used to return the original *fasthttp.Client
	GetOriginal() *fasthttp.Client
}

type instaClient struct {
	*fasthttp.Client
	sensor instana.TracerLogger
}

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

func (ic *instaClient) GetOriginal() *fasthttp.Client {
	return ic.Client
}

func (ic *instaClient) DoTimeout(ctx context.Context, req *fasthttp.Request, resp *fasthttp.Response, timeout time.Duration) error {
	dp := &doParams{
		sensor:  ic.sensor,
		ic:      ic,
		doType:  doFuncWithTimeout,
		timeout: timeout,
	}
	_, err := instrumentedDo(ctx, req, resp, dp)
	return err
}

func (ic *instaClient) DoDeadline(ctx context.Context, req *fasthttp.Request, resp *fasthttp.Response, deadline time.Time) error {
	dp := &doParams{
		sensor:   ic.sensor,
		ic:       ic,
		doType:   doFuncWithDeadline,
		deadline: deadline,
	}
	_, err := instrumentedDo(ctx, req, resp, dp)
	return err
}

func (ic *instaClient) DoRedirects(ctx context.Context, req *fasthttp.Request, resp *fasthttp.Response, maxRedirectsCount int) error {
	dp := &doParams{
		sensor:            ic.sensor,
		ic:                ic,
		doType:            doFuncWithRedirects,
		maxRedirectsCount: maxRedirectsCount,
	}
	_, err := instrumentedDo(ctx, req, resp, dp)
	return err
}

func (ic *instaClient) Do(ctx context.Context, req *fasthttp.Request, resp *fasthttp.Response) error {
	dp := &doParams{
		sensor: ic.sensor,
		ic:     ic,
		doType: doFunc,
	}
	_, err := instrumentedDo(ctx, req, resp, dp)
	return err
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
