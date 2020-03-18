Instana instrumentation for go-grpc library
===========================================

This module contains instrumentation code for GRPC servers and clients that use `google.golang.org/grpc` library.

[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)][godoc]

Installation
------------

Unlike the Instana Go sensor, GRPC instrumentation module requires Go v1.9+ which is the minimal version for `google.golang.org/grpc`.

```bash
$ go get github.com/instana/go-sensor/instrumentation/instagrpc
```

Usage
-----

For detailed usage example see [the documentation][godoc] or [`example_test.go`](./example_test.go).

This instrumentation requires an `opentracing.Tracer` to initialize spans and handle the trace context propagation.
You can create a new instance of Instana tracer using `instana.NewTracer()`.

### Instrumenting a server

To instrument your GRPC server instance include `instagrpc.UnaryServerInterceptor()` and `instagrpc.StreamServerInterceptor()`
into the list of server options passed to `grpc.NewServer()`. These interceptors will use the provided `instana.Sensor` to
handle the OpenTracing headers, start a new span for each incoming request and inject it into the handler:

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

The parent span can be than retrieved inside the handler using `instana.ContextFromSpan()`:

```go
func (s MyServer) SampleCall(ctx context.Context, req *MyRequest) (*MyResponse, error) {
	parentSpan, ok := instana.SpanFromContext(ctx)
	// ...
}
```

### Instrumenting a client

Similar to the server instrumentation, to instrument a GRPC client add `instagrpc.UnaryClientInterceptor()` and
`instagrpc.StreamClientInterceptor()` into the list of dial options passed to the `grpc.Dial()` call. The interceptor
will inject the trace context into each outgoing request made with this connection:

```go
conn, err := grpc.Dial(
	serverAddr,
	grpc.WithUnaryInterceptor(instagrpc.UnaryClientInterceptor(sensor)),
	grpc.WithStreamInterceptor(instagrpc.StreamClientInterceptor(sensor)),
	// ...
)
```

If the context contains an active span stored using `instana.ContextWithSpan()`, the tracer of this span will be used instead.

[godoc]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instagrpc
