package main

import (
	"net/http"

	beego "github.com/beego/beego/v2/server/web"

	beecontext "github.com/beego/beego/v2/server/web/context"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instabeego"
)

type UserJson struct {
	ID   int    `json:"ID"`
	Name string `json:"Name"`
}

type UserController struct {
	beego.Controller
}

func (u *UserController) GetUser() {
	u.Ctx.Output.Status = http.StatusOK
	u.Ctx.JSONResp(UserJson{
		ID:   1,
		Name: "abcd",
	})
}

func main() {
	// create a sensor
	sensor := instana.NewSensor("beego-server")
	// instrument the server with instabeego
	instabeego.InstrumentWebServer(sensor)

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
