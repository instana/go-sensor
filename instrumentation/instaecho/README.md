Instana instrumentation for echo framework
=============================================

This module contains middleware to instrument HTTP services written with [github.com/labstack/echo](https://github.com/labstack/echo).

[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)][godoc]


Installation
------------

To add the module to your `go.mod` file run the following command in your project directory:

```bash
$ go get github.com/instana/go-sensor/instrumentation/instaecho
```

Usage
-----

```go
// init Echo
e := echo.New()

// create a sensor
sensor := instana.NewSensor("echo-sensor")

// add middleware to the Echo's handlers
instaecho.AddMiddleware(sensor, e)

// define API
e.GET("/foo", func(c echo.Context) error {...})
...
```
[Full example][fullExample]



[godoc]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instaecho
[fullExample]: https://github.com/instana/go-sensor/blob/master/example/echo/main.go
