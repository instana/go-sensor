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

// InstrumentRequest wrap the original BeegoHTTPRequest transport with instana.RoundTripper().
func InstrumentRequest(sensor *instana.Sensor, req *httplib.BeegoHTTPRequest) {
	req.AddFilters(func(next httplib.Filter) httplib.Filter {
		return func(ctx context.Context, req *httplib.BeegoHTTPRequest) (*http.Response, error) {
			req.SetTransport(instana.RoundTripper(sensor, nil))
			return next(ctx, req)
		}
	})
}
