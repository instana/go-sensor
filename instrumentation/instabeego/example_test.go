// (c) Copyright IBM Corp. 2023
// (c) Copyright Instana Inc. 2023

//go:build go1.21
// +build go1.21

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

func main() {
	sensor := instana.NewSensor("my-web-server")
	instabeego.New(sensor)

	beego.CtrlGet("api/user/:id", (*UserController).GetUserById)
	beego.Run()
}
