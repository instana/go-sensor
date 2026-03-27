Instana instrumentation for fiber v3
=============================================

This module provides Instana instrumentation for [`fiber/v3`](https://pkg.go.dev/github.com/gofiber/fiber/v3) library.

[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)][godoc]


Installation
------------

To add the module to your `go.mod` file run the following command in your project directory:

```bash
$ go get github.com/instana/go-sensor/instrumentation/instafiber/v2
```

Usage
-----

```go
// Create a collector for instana instrumentation
collector := instana.InitCollector(&instana.Options{
  Service: "fiber-app",
})

app := fiber.New()

// Use the instafiber.TraceHandler for instrumenting the handler
// Parameters: collector, routeID, pathTemplate, handler
app.Get("/greet", instafiber.TraceHandler(collector, "greet", "/greet", func(c fiber.Ctx) error {
  return c.SendString("Hello world!")
}))
```

Refer to [`instafiber`](https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instafiber/v2) package documentation for more details.

[godoc]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instafiber/v2

