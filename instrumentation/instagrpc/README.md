Instana instrumentation for go-grpc library
===========================================

This module contains instrumentation code for GRPC servers and clients that use `google.golang.org/grpc` library.

[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)][godoc]

Installation
------------

To add the module to your `go.mod` file, run the following command in your project directory:

```bash
$ go get github.com/instana/go-sensor/instrumentation/instagrpc
```

Note
----
From `instagrpc` v1.11.0, the minimum version of `grpc` package required has been changed to v1.55.0. For working with older versions of 
`grpc`, one can use `instagrpc` v1.10.0.

Usage
-----

For detailed usage example see [the documentation][godoc] or [`example_test.go`](./example_test.go).

This instrumentation requires an `instana.Sensor` to initialize spans and handle the trace context propagation.
You can create a new instance of Instana tracer using [`instana.NewSensor()`][NewSensor].

### Instrumenting a server

To instrument your GRPC server instance include [`instagrpc.UnaryServerInterceptor()`][UnaryServerInterceptor] and
[`instagrpc.StreamServerInterceptor()`][StreamServerInterceptor] into the list of server options passed to `grpc.NewServer()`.
These interceptors will use the provided [`instana.Sensor`][Sensor] to handle the OpenTracing headers, start a new span for each incoming
request and inject it into the handler:

```go
// initialize a new tracer instance
sensor := instana.NewSensor("my-server")

// instrument the server
srv := grpc.NewServer(
	grpc.UnaryInterceptor(instagrpc.UnaryServerInterceptor(sensor)),
	grpc.StreamInterceptor(instagrpc.StreamServerInterceptor(sensor)),
	// ...
)
```

The parent span can be than retrieved inside the handler using [`instana.SpanFromContext()`][SpanFromContext]:

```go
func (s MyServer) SampleCall(ctx context.Context, req *MyRequest) (*MyResponse, error) {
	parentSpan, ok := instana.SpanFromContext(ctx)
	// ...
}
```

### Instrumenting a client

Similar to the server instrumentation, to instrument a GRPC client add [`instagrpc.UnaryClientInterceptor()`][UnaryClientInterceptor] and
[`instagrpc.StreamClientInterceptor()`][StreamClientInterceptor] to the list of dial options passed to the `grpc.Dial()` call. The interceptor
will inject the trace context into each outgoing request made with this connection:

```go
conn, err := grpc.Dial(
	serverAddr,
	grpc.WithUnaryInterceptor(instagrpc.UnaryClientInterceptor(sensor)),
	grpc.WithStreamInterceptor(instagrpc.StreamClientInterceptor(sensor)),
	// ...
)
```

If the context contains an active span stored using [`instana.ContextWithSpan()`][ContextWithSpan], the tracer of this span will be used instead.

[godoc]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instagrpc
[NewSensor]: https://pkg.go.dev/github.com/instana/go-sensor?tab=doc#NewSensor
[StreamClientInterceptor]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instagrpc?tab=doc#StreamClientInterceptor
[StreamServerInterceptor]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instagrpc?tab=doc#StreamServerInterceptor
[UnaryClientInterceptor]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instagrpc?tab=doc#UnaryClientInterceptor
[UnaryServerInterceptor]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instagrpc?tab=doc#UnaryServerInterceptor
[Sensor]: https://pkg.go.dev/github.com/instana/go-sensor/?tab=doc#Sensor
[SpanFromContext]: https://pkg.go.dev/github.com/instana/go-sensor/?tab=doc#SpanFromContext
[ContextWithSpan]: https://pkg.go.dev/github.com/instana/go-sensor/?tab=doc#ContextWithSpan

<!---
Mandatory comment section for CI/CD !!
target-pkg-url: google.golang.org/grpc
current-version: v1.59.0
--->
