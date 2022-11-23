Instana instrumentation for Microsoft Azure Functions
=====================================================

This module contains instrumentation code for Microsoft Azure functions written in Go that uses the custom runtime.

[![PkgGoDev](https://pkg.go.dev/badge/github.com/instana/go-sensor/instrumentation/instaazurefuntion)](https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instaazurefunction)

Installation
------------

To add `github.com/instana/go-sensor/instrumentation/instaazurefunction` to your `go.mod` file, from your project directory
run:

```bash
$ go get github.com/instana/go-sensor/instrumentation/instaazurefunction
```

Usage
-----

For detailed usage example see [the documentation][godoc] or [`./example/customhandler.go`](./example/customhandler.go).

### Instrumenting a custom handler

To instrument a custom handler, wrap it with [`instaazurefunction.WrapFunctionHandler()`][instaazurefunction.WrapFunctionHandler] before passing 
it to the http router. 

```go
func handlerFn(w http.ResponseWriter, r *http.Request) {
	// ...
}

func main() {
// Initialize a new sensor.
sensor := instana.NewSensor("my-azf-sensor")

// Instrument your handler before passing it to the http router.
http.HandleFunc("/api/azf-test", azf.WrapFunctionHandler(sensor, handlerFn))
}
```

[godoc]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instaazurefunction
[instaazurefunction.WrapFunctionHandler]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instaazurefunction#WrapFunctionHandler
