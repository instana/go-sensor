// (c) Copyright IBM Corp. 2023

//go:build go1.18
// +build go1.18

package instabeego_test

import (
	beego "github.com/beego/beego/v2/server/web"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instabeego"
)

// UserController represents the router for user APIs in the given example
type UserController struct {
	beego.Controller
}

// GetUserById is a sample handler function for the example
func (u *UserController) GetUserById() {
	u.Ctx.WriteString("GetUserById")
}

// This example shows how to instrument a beego web server.
func Example_serverInstrumentation() {
	t := instana.InitCollector(&instana.Options{
		Service:           "beego-server",
		EnableAutoProfile: true,
	})
	// This will add instana.TracingHandlerFunc() function as middleware for every API.
	instabeego.InstrumentWebServer(t)

	beego.CtrlGet("api/user/:id", (*UserController).GetUserById)
	beego.Run()
}
