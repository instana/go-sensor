// SPDX-FileCopyrightText: 2026 IBM Corp.
//
// SPDX-License-Identifier: MIT

package instafiber_test

import (
	"log"

	"github.com/gofiber/fiber/v3"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instafiber/v2"
)

func Example() {

	// Create a collector for instana instrumentation
	c := instana.InitCollector(&instana.Options{
		Service: "my-service",
	})
	defer instana.ShutdownCollector()

	app := fiber.New()

	// Use the instafiber.TraceHandler for instrumenting the handler
	app.Get("/greet", instafiber.TraceHandler(c, "greet", "/greet", hello))

	// Start server
	log.Fatal(app.Listen(":3000"))
}

func hello(c fiber.Ctx) error {
	return c.SendString("Hello world!")
}
