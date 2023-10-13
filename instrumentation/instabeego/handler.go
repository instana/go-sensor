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
	beego.InsertFilter("*", beego.BeforeRouter, func(ctx *beecontext.Context) {
		instana.TracingHandlerFunc(sensor, ctx.Request.URL.Path, func(w http.ResponseWriter, r *http.Request) {
			ctx.Request = r
			// ctx.ResponseWriter.ResponseWriter = w
		})(ctx.ResponseWriter, ctx.Request)
	})
}
