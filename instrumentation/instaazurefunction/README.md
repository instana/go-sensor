Instana instrumentation for Microsoft Azure Functions
=====================================================

This module contains the instrumentation code for Microsoft Azure functions written in Go that uses the custom runtime.

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
http.HandleFunc("/api/azf-test", instaazurefunction.WrapFunctionHandler(sensor, handlerFn))
}
```

Refer the [`instaazurefunction`](https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instaazurefunction) package documentation for more details.

[godoc]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instaazurefunction
[instaazurefunction.WrapFunctionHandler]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instaazurefunction#WrapFunctionHandler

Limitations
-----------
- The instrumentation supports only HTTP and Queue trigger types.
- The instrumentation cannot support HTTP triggers if `enableForwardingHttpRequest` is set to `true` in the `host.json` file.