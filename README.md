![golang banner 2017-07-11](https://disznc.s3.amazonaws.com/Instana-Go-2017-07-11-at-16.01.45.png)

# Instana Go Sensor
go-sensor requires Go version 1.8 or greater.

The Instana Go sensor consists of two parts:

* metrics sensor
* [OpenTracing](http://opentracing.io) tracer

[![Build Status](https://travis-ci.org/instana/go-sensor.svg?branch=master)](https://travis-ci.org/instana/go-sensor)
[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)](https://pkg.go.dev/github.com/instana/go-sensor)
[![OpenTracing Badge](https://img.shields.io/badge/OpenTracing-enabled-blue.svg)](http://opentracing.io)

## Common Operations

The Instana Go sensor offers a set of quick features to support tracing of the most common operations like handling HTTP requests and executing HTTP requests.

To create an instance of the Instana sensor just request a new instance using the `instana.NewSensor` factory method and providing the name of the application. It is recommended to use a single instance only. The sensor implementation is fully thread-safe and can be shared by multiple threads.

```go
var sensor = instana.NewSensor("my-service")
```

A full example can be found under the examples folder in [example/webserver/instana/http.go](./example/webserver/instana/http.go).

### Log Output

The Go sensor uses a leveled logger to log internal errors and diagnostic information. The default `logger.Logger` uses `log.Logger`
configured with `log.Lstdflags` as a backend and writes messages to `os.Stderr`. By default this logger only prints out the `ERROR` level
messages unless the environment variable `INSTANA_DEBUG` is set.

To change the min log level in runtime it is recommended to configure and inject an instance of `instana.LeveledLogger` instead of using
the deprecated `instana.SetLogLevel()` method:

```go
l := logger.New(log.New(os.Stderr, "", os.Lstdflags))
instana.SetLogger(l)

// ...

l.SetLevel(logger.WarnLevel)
```

The `logger.LeveledLogger` interface is implemented by such popular logging libraries as [`github.com/sirupsen/logrus`](https://github.com/sirupsen/logrus) and [`go.uber.org/zap`](https://go.uber.org/zap), so they can be used as a replacement.

**Note**: the value of `INSTANA_DEBUG` environment variable does not affect custom loggers. You'd need to explicitly check whether it's set
and enable the debug logging while onfiguring your logger:

```go
import (
	instana "github.com/instana/go-sensor"
	"github.com/sirupsen/logrus"
)

func main() {	
	// initialize Instana sensor
	instana.InitSensor(&instana.Options{Service: SERVICE})

	// initialize and configure the logger
	logger := logrus.New()
	logger.Level = logrus.InfoLevel

	// check if INSTANA_DEBUG is set and set the log level to DEBUG if needed
	if _, ok := os.LookupEnv("INSTANA_DEBUG"); ok {
		logger.Level = logrus.DebugLevel
	}

	// use logrus to log the Instana Go sensor messages
	instana.SetLogger(logger)

	// ...
}
```

The Go sensor [AutoProfile™](#autoprofile) by default uses the same logger as the sensor itself, however it can be configured to
use its own, for example to write messages with different tags/prefix or use a different logging level. The following snippet
demonstrates how to configure the custom logger for autoprofiler:

```
autoprofile.SetLogger(autoprofileLogger)
```

### Trace Context Propagation

Instana Go sensor provides an API to propagate the trace context throughout the call chain:

```go
func MyFunc(ctx context.Context) {
	var spanOpts []ot.StartSpanOption

	// retrieve parent span from context and reference it in the new one
	if parent, ok := instana.SpanFromContext(ctx); ok {
	    spanOpts = append(spanOpts, ot.ChildOf(parent.Context()))
	}

	// start a new span
	span := tracer.StartSpan("my-func", spanOpts...)
	defer span.Finish()

	// and use it as a new parent inside the context
	SubCall(instana.ContextWithSpan(ctx, span))
}
```

### HTTP Server Handlers

With support to wrap a `http.HandlerFunc`, Instana quickly adds the possibility to trace requests and collect child spans, executed in the context of the request span.

Minimal changes are required for Instana to be able to capture the necessary information. By simply wrapping the currently existing `http.HandlerFunc` Instana collects and injects necessary information automatically.

That said, a simple handler function like the following will simple be wrapped and registered like normal.

For your own preference registering the handler and wrapping it can be two separate steps or a single one. The following example code shows both versions, starting with two steps.

```go
import (
	"net/http"

	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
)

// Doing registration and wrapping
func main() {
	http.HandleFunc(
		"/path/to/handler",
		sensor.TracingHandlerFunc("myHandler", myHandler),
	)
}

// Accessing the parent request inside a handler
func myHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	parent, _ := instana.SpanFromContext(ctx))
    
	tracer := parent.Tracer()
	spanCtx := parent.Context().(instana.SpanContext)
	traceID := spanCtx.TraceID // use this with EumSnippet
}
```

### Executing HTTP Requests

Requesting data or information from other, often external systems, is commonly implemented through HTTP requests. To make sure traces contain all spans, especially over all the different systems, certain span information have to be injected into the HTTP request headers before sending it out. Instana's Go sensor provides support to automate this process as much as possible.

To have Instana inject information into the request headers, create the `http.Client`, wrap its `Transport` with `instana.RoundTripper()` and use it as in the following example. 

```go
req, err := http.NewRequest("GET", url, nil)
client := &http.Client{
	Transport: instana.RoundTripper(sensor, nil),
}

ctx := instana.ContextWithSpan(context.Background(), parentSpan)
resp, err := client.Do(req.WithContext(ctx))
```

The provided `parentSpan` is the incoming request from the request handler (see above) and provides the necessary tracing and span information to create a child span and inject it into the request.

### GRPC servers and clients

[`github.com/instana/go-sensor/instrumentation/instagrpc`](./instrumentation/instagrpc) provides both unary and stream interceptors to instrument GRPC servers and clients that use `google.golang.org/grpc`.

### Kafka producers and consumers

[`github.com/instana/go-sensor/instrumentation/instasarama`](./instrumentation/instasarama) provides both unary and stream interceptors to instrument Kafka producers and consumers built on top of `github.com/Shopify/sarama`.

## Sensor

To use sensor only without tracing ability, import the `instana` package and run

```go
instana.InitSensor(opt)
```

in your main function. The init function takes an `Options` object with the following optional fields:

* **Service** - global service name that will be used to identify the program in the Instana backend
* **AgentHost**, **AgentPort** - default to `localhost:42699`, set the coordinates of the Instana proxy agent
* **LogLevel** - one of `Error`, `Warn`, `Info` or `Debug`
* **EnableAutoProfile** - enables automatic continuous process profiling when `true`

Once initialized, the sensor will try to connect to the given Instana agent and in case of connection success will send metrics and snapshot information through the agent to the backend.

## OpenTracing

In case you want to use the OpenTracing tracer, it will automatically initialize the sensor and thus also activate the metrics stream. To activate the global tracer, run for example

```go
ot.InitGlobalTracer(instana.NewTracerWithOptions(&instana.Options{
	Service:  SERVICE,
	LogLevel: instana.DEBUG,
}))
```

in your main function. The tracer takes the same options that the sensor takes for initialization, described above.

The tracer is able to protocol and piggyback OpenTracing baggage, tags and logs. Only text mapping is implemented yet, binary is not supported. Also, the tracer tries to map the OpenTracing spans to the Instana model based on OpenTracing recommended tags. See [simple.go](./example/ot-simple/simple.go) example for details on how recommended tags are used.

The Instana tracer will remap OpenTracing HTTP headers into Instana Headers, so parallel use with some other OpenTracing model is not possible. The Instana tracer is based on the OpenTracing Go basictracer with necessary modifications to map to the Instana tracing model. Also, sampling isn't implemented yet and will be focus of future work.

## W3C Trace Context

The Go sensor library fully supports [the W3C Trace Context standard](https://www.w3.org/TR/trace-context/):

* An [instrumented `http.Client`][instana.RoundTripper] sends the `traceparent` and `tracestate` headers, updating them with the exit span ID and flags.
* Any `http.Handler` instrumented with [`instana.TracingHandlerFunc()`][instana.TracingHandlerFunc] picks up the trace context passed in the `traceparent` header, potentially restoring the trace from `tracestate` even if the upstream service is not instrumented with Instana.

## Events API

The sensor, be it instantiated explicitly or implicitly through the tracer, provides a simple wrapper API to send events to Instana as described in [its documentation](https://docs.instana.io/quick_start/api/#event-sdk-rest-web-service).

To learn more, see the [Events API](https://github.com/instana/go-sensor/blob/master/EventAPI.md) document in this repository.

## AutoProfile™ 

AutoProfile™ generates and reports process profiles to Instana. Unlike development-time and on-demand profilers, where a user must manually initiate profiling, AutoProfile™ automatically schedules and continuously performs profiling appropriate for critical production environments.


## Examples

Following examples are included in the `example` folder:

* [ot-simple/simple.go](./example/ot-simple/simple.go) - Demonstrates basic usage of the tracer
* [webserver/http.go](./example/webserver/http.go) - Demonstrates how http server and client should be instrumented
* [rpc/rpc.go](./example/rpc/rpc.go) - Demonstrates a basic RPC service
* [event/](./example/event/) - Demonstrates usage of the Events API
* [database/elasticsearch.go](./example/database/elasticsearch.go) - Demonstrates how to instrument a database client (Elasticsearch in this case)
* [httpclient/multi_request.go](./example/httpclient/multi_request.go) - Demonstrates the instrumentation of an HTTP client
* [many.go](./example/many.go) - Demonstrates how to create nested spans within the same execution context

[instana.TracingHandlerFunc]: https://pkg.go.dev/github.com/instana/go-sensor/?tab=doc#TracingHandlerFunc
[instana.RoundTripper]: https://pkg.go.dev/github.com/instana/go-sensor/?tab=doc#RoundTripper
