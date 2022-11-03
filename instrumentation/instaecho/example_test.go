// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2016

//go:build go1.16
// +build go1.16

package instaecho_test

import (
	"log"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instaecho"
	"github.com/labstack/echo/v4"
)

// This example shows how to instrument an HTTP server that uses github.com/labstack/echo with Instana
func Example() {
	sensor := instana.NewSensor("my-web-server")

	// Use instaecho.New() to create a new instance of Echo. The returned instance is instrumented
	// with Instana and will create an entry HTTP span for each incoming request.
	engine := instaecho.New(sensor)

	// Use the instrumented instance as usual
	engine.GET("/myendpoint", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"message": "pong",
		})
	})

	log.Fatalln(engine.Start(":0"))
}
