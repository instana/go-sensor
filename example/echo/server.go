// SPDX-FileCopyrightText: 2026 IBM Corp.
// SPDX-FileCopyrightText: 2026 Instana Inc.
//
// SPDX-License-Identifier: MIT

package main

import (
	"log"

	instana "github.com/instana/go-sensor"
	instaechov2 "github.com/instana/go-sensor/instrumentation/instaecho/v2"
	"github.com/labstack/echo/v5"
)

func main() {
	// Initialize Instana collector
	collector := instana.InitCollector(&instana.Options{
		Service: "echo-v5-example",
	})

	// Create instrumented Echo v5 instance
	e := instaechov2.New(collector)

	// Define routes - they will be automatically instrumented
	e.GET("/myendpoint", func(c *echo.Context) error {
		return c.JSON(200, map[string]string{
			"message": "pong",
		})
	})

	// Start server
	log.Fatal(e.Start(":8080"))
}
