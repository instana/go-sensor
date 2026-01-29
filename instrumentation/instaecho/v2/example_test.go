// SPDX-FileCopyrightText: 2026 IBM Corp.
// SPDX-FileCopyrightText: 2026 Instana Inc.
//
// SPDX-License-Identifier: MIT

//go:build go1.25
// +build go1.25

package instaechov2_test

import (
	"log"

	instana "github.com/instana/go-sensor"
	instaechov2 "github.com/instana/go-sensor/instrumentation/instaecho/v2"
	"github.com/labstack/echo/v5"
)

// This example shows how to instrument an HTTP server that uses github.com/labstack/echo/v5 with Instana
func Example() {
	c := instana.InitCollector(&instana.Options{
		Service: "test_service",
	})

	// Use instaechov2.New() to create a new instance of Echo v5. The returned instance is instrumented
	// with Instana and will create an entry HTTP span for each incoming request.
	engine := instaechov2.New(c)

	// Use the instrumented instance as usual
	engine.GET("/myendpoint", func(c *echo.Context) error {
		return c.JSON(200, map[string]string{
			"message": "pong",
		})
	})

	log.Fatalln(engine.Start(":0"))
}
