Instana instrumentation for gin framework
=============================================

This module contains middleware to instrument HTTP services written with [github.com/gin-gonic/gin](https://github.com/gin-gonic/gin).

[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)][godoc]

Compatibility
-------------

 * Starting at version 1.6.0 of instagin, [gin v1.9.0](https://github.com/gin-gonic/gin/releases/tag/v1.9.0) is used, which requires Go v1.18 or higher


Installation
------------

To add the module to your `go.mod` file run the following command in your project directory:

```bash
$ go get github.com/instana/go-sensor/instrumentation/instagin
```

Usage
-----

```go
// init gin engine
engine := gin.Default()

// create a collector
collector := instana.InitCollector(&instana.Options{
    Service: "rabbitmq-client",
    Tracer:  instana.DefaultTracerOptions(),
})

// add middleware to the gin handlers
instagin.AddMiddleware(collector, engine)

// define API
engine.GET("/api", func(c *gin.Context) {}
...
```
[Full example][fullExample]



[godoc]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instagin
[fullExample]: https://github.com/instana/go-sensor/blob/main/example/gin/main.go
