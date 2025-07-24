Instana instrumentation for AWS Lambda
======================================

This module contains instrumentation code for AWS Lambda functions written in Go that use
[`github.com/aws/aws-lambda-go`](https://github.com/aws/aws-lambda-go) as a runtime.

[![PkgGoDev](https://pkg.go.dev/badge/github.com/instana/go-sensor/instrumentation/instalambda)](https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instalambda)

Installation
------------

To add `github.com/instana/go-sensor/instrumentation/instalambda` to your `go.mod` file, from your project directory
run:

```bash
$ go get github.com/instana/go-sensor/instrumentation/instalambda
```

Usage
-----

For detailed usage example see [the documentation][godoc] or [`example_test.go`](./example_test.go).

### Instrumenting a `lambda.Handler`

To instrument a `lambda.Handler` wrap it with [`instalambda.WrapHandler()`][instalambda.WrapHandler] before passing it
to `labmda.StartHandler()`:

```go
type Handler struct {
	// ...
}

func (Handler) Invoke(ctx context.Context, payload []byte) ([]byte, error) {
	// ...
}

func main() {
	// Initialize a new collector
	collector := instana.InitCollector(&instana.Options{
		Service: "go-lambda",
		Tracer:  instana.DefaultTracerOptions(),
	})

	h := &Handler{
		// ...
	}

	// Instrument your handler before passing it to lambda.StartHandler()
	lambda.StartHandler(instalambda.WrapHandler(h, collector))
}
```

### Instrumenting a handler function

To instrument a handler function passed to `lambda.Start()` or `lambda.StartWithContext()` first create an instrumented
handler from it using [`instalambda.NewHandler()`][instalambda.NewHandler] and then pass it to `lambda.StartHandler()`:

```go
func handle() {
	return "Hello, ƛ!", nil
}

func main() {
	// Initialize a new collector
	collector := instana.InitCollector(&instana.Options{
		Service: "graphql-app",
		Tracer:  instana.DefaultTracerOptions(),
	})

	// Create a new instrumented lambda.Handler from your handle function
	h := instalambda.NewHandler(func() (string, error) {

	}, collector)

	// Pass the instrumented handler to lambda.StartHandler()
	lambda.StartHandler(h)
}
```

### Trace context propagation

Whenever a handler function accepts `context.Context` as a first argument (and `(lambda.Handler).Invoke()` always does), `instalambda`
instrumentation injects the entry span for this Lambda invokation into it. This span can be retireved with
[`instana.SpanFromContext()`][instana.SpanFromContext] and used as a parent to create any intermediate or exit spans within the handler function:

```go
func MyHandler(ctx context.Context) error {
	// Pass the handler context to a subcall to trace its execution
	subCall(ctx)

	// ...

	// Propagate the trace context within an HTTP request to another service monitored with Instana
	// using an instrumented http.Client
	req, err := http.NewRequest("GET", url, nil)
    client := &http.Client{
	    Transport: instana.RoundTripper(collector, nil),
	}

	client.Do(req.WithContext(ctx))

	// ...
}

func subCall(ctx context.Context) {
	if parent, ok := instana.SpanFromContext(ctx); ok {
		// start a new span, using the Lambda entry span as a parent
		sp = parent.Tracer().StartSpan(/* ... */)
		defer sp.Finish()
	}

	// ...
}
```

[godoc]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instalambda
[instalambda.NewHandler]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instalambda#NewHandler
[instalambda.WrapHandler]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instalambda#WrapHandler
[instana.SpanFromContext]: https://pkg.go.dev/github.com/instana/go-sensor#SpanFromContext
