// (c) Copyright IBM Corp. 2023

//go:build go1.18
// +build go1.18

package instabeego_test

import (
	beego "github.com/beego/beego/v2/server/web"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instabeego"
)

type UserController struct {
	beego.Controller
}

func (u *UserController) GetUserById() {
	u.Ctx.WriteString("GetUserById")
}

// This example shows how to instrument a beego web server.
func Example_server_instrument() {
	sensor := instana.NewSensor("my-web-server")
	// This will add instana.TracingHandlerFunc() function as middleware for every API.
	instabeego.InstrumentWebServer(sensor)

	beego.CtrlGet("api/user/:id", (*UserController).GetUserById)
	beego.Run()
}
