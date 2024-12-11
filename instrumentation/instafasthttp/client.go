// (c) Copyright IBM Corp. 2024

package instafasthttp

import (
	"context"
	"time"

	instana "github.com/instana/go-sensor"
	"github.com/valyala/fasthttp"
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
