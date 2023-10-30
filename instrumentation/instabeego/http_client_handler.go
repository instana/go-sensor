// (c) Copyright IBM Corp. 2023

//go:build go1.18
// +build go1.18

package instabeego

import (
	"context"
	"net/http"

	"github.com/beego/beego/v2/client/httplib"
	instana "github.com/instana/go-sensor"
)

type FilterChainBuilder struct {
	Sensor *instana.Sensor
}

// FilterChain wrap the original BeegoHTTPRequest transport with instana.RoundTripper().
func (builder *FilterChainBuilder) FilterChain(next httplib.Filter) httplib.Filter {
	return func(ctx context.Context, req *httplib.BeegoHTTPRequest) (*http.Response, error) {
		req.SetTransport(instana.RoundTripper(builder.Sensor, nil))
		return next(ctx, req)
	}
}
