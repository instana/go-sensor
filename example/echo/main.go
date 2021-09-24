// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2016

// +build go1.13

package main

import (
	"flag"
	"os"

	"github.com/instana/go-sensor/instrumentation/instaecho"

	"github.com/labstack/echo/v4"

	instana "github.com/instana/go-sensor"
)

var listenAddr string

func main() {
	flag.StringVar(&listenAddr, "l", os.Getenv("LISTEN_ADDR"), "Server listen address")
	flag.Parse()

	if listenAddr == "" {
		flag.Usage()
		os.Exit(2)
	}
	// create a sensor
	sensor := instana.NewSensor("sensor")
	// create an instrumented Echo instance
	engine := instaecho.New(sensor)

	engine.GET("/myendpoint", func(c echo.Context) error {
		return c.JSON(200, []byte(`{"message": "pong"}`))
	})

	// use group: v1
	v1 := engine.Group("/v1")

	v1.GET("/myendpoint", func(c echo.Context) error {
		return c.JSON(200, []byte(`{"message": "ping"}`))
	})

	engine.Start(listenAddr)
}
