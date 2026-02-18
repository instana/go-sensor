Instana instrumentation for Echo v5 framework
=============================================

This module contains middleware to instrument HTTP services written with [`Echo v5`](https://github.com/labstack/echo).

[![PkgGoDev](https://pkg.go.dev/badge/github.com/instana/go-sensor/instrumentation/instaecho/v2)][godoc]

Installation
------------

To add the module to your `go.mod` file run the following command in your project directory:

```bash
$ go get github.com/instana/go-sensor/instrumentation/instaecho/v2
```

Features
--------

- **Automatic tracing**: Creates entry spans for all incoming HTTP requests
- **Trace context propagation**: Propagates trace context to downstream handlers and services
- **Route information capture**: Records route names, path templates, and HTTP methods
- **Custom header collection**: Supports collecting custom HTTP headers for tracing

Usage
-----

### Basic Setup

The simplest way to use the instrumentation is with the `New()` function, which returns a fully instrumented Echo v5 instance:

```go
// create an instana collector
collector := instana.InitCollector(&instana.Options{
    Service: "echo-app",
    Tracer:  instana.DefaultTracerOptions(),
})

// init instrumented Echo v5
e := instaechov2.New(collector)

// define API
e.GET("/foo", func(c *echo.Context) error { /* ... */ })
// ...
```

### Using Middleware Separately

If you need more control over your Echo instance configuration, you can use the `Middleware()` function directly:

```go
collector := instana.InitCollector(&instana.Options{
    Service: "echo-app",
    Tracer:  instana.DefaultTracerOptions(),
})

e := echo.New()

// Add Instana middleware before defining routes
e.Use(instaechov2.Middleware(collector))

// define API
e.GET("/foo", func(c *echo.Context) error { /* ... */ })
```

**Important**: The middleware should be added before defining routes to ensure all handlers are instrumented.

### Collecting Custom Headers

You can configure the tracer to collect specific HTTP headers:

```go
collector := instana.InitCollector(&instana.Options{
    Service: "echo-app",
    Tracer: instana.TracerOptions{
        CollectableHTTPHeaders: []string{"x-custom-header-1", "x-custom-header-2"},
    },
})

e := instaechov2.New(collector)
```

Example
-------

For a complete working example demonstrating Echo v5 instrumentation with database integration and context propagation, see the [Echo v5 example application](../../../example/echo/README.md).

The example showcases:
- Complete application setup with Instana instrumentation
- Database query instrumentation with SQLite
- Context propagation from HTTP requests to database operations
- RESTful API endpoints with automatic tracing

[godoc]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instaecho/v2