Instana instrumentation for Echo framework
=============================================

This module contains middleware to instrument HTTP services written with [`github.com/labstack/echo`](https://github.com/labstack/echo).

[![PkgGoDev](https://pkg.go.dev/badge/github.com/instana/go-sensor/instrumentation/instaecho)][godoc]

Installation
------------

To add the module to your `go.mod` file run the following command in your project directory:

```bash
$ go get github.com/instana/go-sensor/instrumentation/instaecho
```

Usage
-----

```go
// create an instana collector
collector := instana.InitCollector(&instana.Options{
    Service: "echo-app",
    Tracer:  instana.DefaultTracerOptions(),
})

// init instrumented Echo
e := instaecho.New(collector)

// define API
e.GET("/foo", func(c echo.Context) error { /* ... */ })
// ...
```
[Full example][fullExample]


[godoc]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instaecho
[fullExample]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instaecho#example-package

