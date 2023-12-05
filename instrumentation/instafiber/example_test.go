// (c) Copyright IBM Corp. 2023

//go:build go1.17
// +build go1.17

package instafiber_test

import (
	"log"

	"github.com/gofiber/fiber/v2"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instafiber"
)

func Example() {

	// Create a sensor for instana instrumentation
	sensor := instana.NewSensor("my-service")

	app := fiber.New()

	// Use the instafiber.TraceHandler for instrumenting the handler
	app.Get("/greet", instafiber.TraceHandler(sensor, "greet", "/greet", hello))

	// Start server
	log.Fatal(app.Listen(":3000"))
}

func hello(c *fiber.Ctx) error {
	return c.SendString("Hello world!")
}
