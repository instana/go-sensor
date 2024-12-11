// (c) Copyright IBM Corp. 2024

package instafasthttp

import (
	"context"

	instana "github.com/instana/go-sensor"
	"github.com/valyala/fasthttp"
)

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
		cfp := &clientFuncParams{
			sensor:         sensor,
			hc:             hc,
			rt:             original,
			clientFuncType: doRoundTripFunc,
		}
		retry, err := instrumentClient(ctx, req, resp, cfp)
		return retry, err
	})
}
