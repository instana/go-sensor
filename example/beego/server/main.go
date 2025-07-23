// (c) Copyright IBM Corp. 2023

//go:build go1.18
// +build go1.18

package main

import (
	"net/http"

	beego "github.com/beego/beego/v2/server/web"

	beecontext "github.com/beego/beego/v2/server/web/context"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instabeego"
)

// UserJson defines the response for the APIs in the given example
type UserJson struct {
	ID   int    `json:"ID"`
	Name string `json:"Name"`
}

// UserController handles the user API requests
type UserController struct {
	beego.Controller
}

// GetUser is a handler function for the sample API in the given example
func (u *UserController) GetUser() {
	u.Ctx.Output.Status = http.StatusOK
	u.Ctx.JSONResp(UserJson{
		ID:   1,
		Name: "abcd",
	})
}

func main() {
	// create an Instana collector
	collector := instana.InitCollector(&instana.Options{
		Service: "beego-server",
		Tracer:  instana.DefaultTracerOptions(),
	})

	// instrument the server with instabeego
	instabeego.InstrumentWebServer(collector)

	// Controller Style Router
	beego.CtrlGet("/controller/user/:id", (*UserController).GetUser)

	// Functional Style Router
	beego.Get("/functional/user/:id", func(ctx *beecontext.Context) {
		user := UserJson{
			ID:   1,
			Name: "abcd",
		}

		ctx.Output.SetStatus(http.StatusOK)
		ctx.JSONResp(user)
	})

	beego.Run("localhost:8080")
}
