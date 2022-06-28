package main

import (
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instaecho"
	"github.com/labstack/echo/v4"
)

func main() {

	sensor := instana.NewSensor("test-echo")

	engine := instaecho.New(sensor)
	engine.GET("/projects/api/health", func(c echo.Context) error {
		return c.JSON(200, []byte("{}"))
	})

	engine.GET("/user/:id", func(c echo.Context) error {
		return c.JSON(200, "user with id")
	})

	engine.Start("0.0.0.0:9090")
}
