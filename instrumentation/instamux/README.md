Instana instrumentation for gorilla mux
=============================================

This module contains middleware to instrument HTTP services written with [`github.com/gorilla/mux`](https://github.com/gorilla/mux).

[![PkgGoDev](https://pkg.go.dev/badge/github.com/instana/go-sensor/instrumentation/instamux)][godoc]

Installation
------------

To add the module to your `go.mod` file run the following command in your project directory:

```bash
$ go get github.com/instana/go-sensor/instrumentation/instamux
```

Usage
-----

```go
// Create a sensor
sensor := instana.NewSensor("my-web-server")

// Create router
r := mux.NewRouter()

// Instrument your router by adding a middleware
instamux.AddMiddleware(sensor, r)

// Define handlers
r.HandleFunc("/foo", func(writer http.ResponseWriter, request *http.Request) {})
...
```
[Full example][fullExample]



[godoc]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instamux
[fullExample]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instamux#example-package

<!---
Mandatory comment section for CI/CD !!
target-pkg-url: github.com/gorilla/mux
current-version: v1.8.0
--->
