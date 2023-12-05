Instana instrumentation for github.com/julienschmidt/httprouter
===============================================================

This module contains middleware to instrument HTTP services written with [github.com/julienschmidt/httprouter](https://github.com/julienschmidt/httprouter).

[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)][godoc]


Installation
------------

To add the module to your `go.mod` file run the following command in your project directory:

```bash
$ go get github.com/instana/go-sensor/instrumentation/instahttprouter
```

Usage
-----

```go
// Create a sensor
sensor := instana.NewSensor("my-web-server")

// Create router and wrap it with Instana
r := instahttprouter.Wrap(httprouter.New(), sensor)

// Define handlers
r.GET("/foo", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {})
r.Handle(http.MethodPost, "/foo/:id", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {})

// There is no need to additionally instrument your handlers with instana.TracingHandlerFunc(), since
// the instrumented router takes care of this during the registration process.
r.HandlerFunc(http.MethodDelete, "/foo/:id", func(writer http.ResponseWriter, request *http.Request) {})

// ...
```

[Full example](../../example/httprouter/main.go)

[godoc]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instahttprouter

<!---
Mandatory comment section for CI/CD !!
target-pkg-url: github.com/julienschmidt/httprouter
current-version: v1.3.0
--->
