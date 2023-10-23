// (c) Copyright IBM Corp. 2023
// (c) Copyright Instana Inc. 2023

//go:build go1.21
// +build go1.21

package instabeego

import (
	"net/http"

	beego "github.com/beego/beego/v2/server/web"
	beecontext "github.com/beego/beego/v2/server/web/context"
	instana "github.com/instana/go-sensor"
)

func New(sensor *instana.Sensor) {
	beego.InsertFilterChain("*", func(next beego.FilterFunc) beego.FilterFunc {
		return func(ctx *beecontext.Context) {
			instana.TracingHandlerFunc(sensor, ctx.Request.URL.Path, func(w http.ResponseWriter, r *http.Request) {
				ctx.Request = r
				ctx.ResponseWriter = &beecontext.Response{
					ResponseWriter: w,
					Started:        ctx.ResponseWriter.Started,
					Status:         ctx.ResponseWriter.Status,
					Elapsed:        ctx.ResponseWriter.Elapsed,
				}
				next(ctx)
			})(ctx.ResponseWriter, ctx.Request)
		}
	})
}
