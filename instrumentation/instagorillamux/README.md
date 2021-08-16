Instana instrumentation for gorilla mux
=============================================

This module contains middleware to instrument HTTP services written with [github.com/gorilla/mux](https://github.com/gorilla/mux).

[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)][godoc]


Installation
------------

To add the module to your `go.mod` file run the following command in your project directory:

```bash
$ go get github.com/instana/go-sensor/instrumentation/instagorillamux
```

Usage
-----

```go
// create router
r := mux.NewRouter()

// define handler
r.HandleFunc("/foo", func(writer http.ResponseWriter, request *http.Request) {})

// create a sensor
sensor := instana.NewSensor("gorillamux-sensor")

// add middleware
instagorillamux.AddMiddleware(sensor, r)
...
```
[Full example][fullExample]



[godoc]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instagorillamux
[fullExample]: https://github.com/instana/go-sensor/blob/master/example/gorillamux/main.go
