Instana instrumentation for fiber
=============================================

This module provides Instana instrumentation for [`fiber`](https://github.com/gofiber/fiber) library.

[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)][godoc]


Installation
------------

To add the module to your `go.mod` file run the following command in your project directory:

```bash
$ go get github.com/instana/go-sensor/instrumentation/instafiber
```

Usage
-----

```go
// Create a collector for instana instrumentation
collector := instana.InitCollector(&instana.Options{
  Service: "fiber-app",
  Tracer:  instana.DefaultTracerOptions(),
})

app := fiber.New()

// Use the instafiber.TraceHandler for instrumenting the handler
app.Get("/greet", instafiber.TraceHandler(collector, "greet", "/greet", func(c *fiber.Ctx) error {
return c.SendString("Hello world!")
}))
```

Refer to [`instafiber`](https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instafiber) package documentation for more details.

[godoc]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instafiber

