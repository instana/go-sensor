![golang banner 2017-07-11](https://disznc.s3.amazonaws.com/Instana-Go-2017-07-11-at-16.01.45.png)

# Instana Go Sensor
go-sensor requires Go version 1.9 or greater.

The Instana Go sensor consists of three parts:

* Metrics sensor
* [OpenTracing](http://opentracing.io) tracer
* AutoProfile™ continuous profiler

[![Build Status](https://circleci.com/gh/instana/go-sensor/tree/master.svg?style=svg)](https://circleci.com/gh/instana/go-sensor/tree/master)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/instana/go-sensor)][pkg.go.dev]
[![OpenTracing](https://img.shields.io/badge/OpenTracing-enabled-blue.svg)](http://opentracing.io)
[![Go Report Card](https://goreportcard.com/badge/github.com/instana/go-sensor)](https://goreportcard.com/report/github.com/instana/go-sensor)

## Table of Contents

* [Installation](#installation)
  * [Running in serverless environment](#running-in-serverless-environment)
  * [Using Instana to gather process metrics only](#using-instana-to-gather-process-metrics-only)
* [Common Operations](#common-operations)
  * [Setting the sensor log output](#setting-the-sensor-log-output)
  * [Trace Context Propagation](#trace-context-propagation)
  * [Secrets Filtering](#secrets-filtering)
  * [HTTP servers and clients](#http-servers-and-clients)
    * [Instrumenting HTTP request handling](#instrumenting-http-request-handling)
    * [Instrumenting HTTP request execution](#instrumenting-http-request-execution)
    * [Capturing custom HTTP headers](#capturing-custom-http-headers)
  * [Database Calls](#database-calls)
    * [Instrumenting sql\.Open()](#instrumenting-sqlopen)
    * [Instrumenting sql\.OpenDB()](#instrumenting-sqlopendb)
  * [GRPC servers and clients](#grpc-servers-and-clients)
  * [Kafka producers and consumers](#kafka-producers-and-consumers)
* [OpenTracing](#opentracing)
* [W3C Trace Context](#w3c-trace-context)
* [Events API](#events-api)
* [AutoProfile™](#autoprofile)
  * [Activation from within the application code](#activation-from-within-the-application-code)
  * [Activation without code changes](#activation-without-code-changes)
* [Examples](#examples)

## Installation

To add Instana Go sensor to your service run:

```bash
$ go get github.com/instana/go-sensor
```

To activate background metrics collection, add following line at the beginning of your service initialization (typically this would be the beginning of your `main()` function):

```go
func main() {
	instana.InitSensor(instana.DefaultOptions())

	// ...
}
```

The `instana.InitSensor()` function takes an [`*instana.Options`][instana.Options] to provide the initial configuration for the sensor.

**Note:** subsequent calls to `instana.InitSensor()` with different configuration won't have any effect.

Once initialized, the sensor performs a host agent lookup using following list of addresses (in order of priority):

1. The value of `INSTANA_AGENT_HOST` env variable
2. `localhost`
3. Default gateway

Once a host agent found listening on port `42699` (or the port specified in `INSTANA_AGENT_PORT` env variable) the sensor begins collecting in-app metrics and sending them to the host agent.

### Running in serverless environment

To use Instana Go sensor for monitoring a service running in a serverless environment, such as AWS Fargate or Google Cloud Run, make sure that you have `INSTANA_ENDPOINT_URL` and `INSTANA_AGENT_KEY` env variables set in your task definition. Note that the `INSTANA_AGENT_HOST` and `INSTANA_AGENT_PORT` env variables will be ignored in this case. Please refer to the respective section of Instana documentation for detailed explanation on how to do this:

* [Configuring AWS Fargate task definitions](https://www.instana.com/docs/ecosystem/aws-fargate/#configure-your-task-definition)
* [Configuring AWS Lambda functions](https://www.instana.com/docs/ecosystem/aws-lambda/go)
* [Configuring Google Cloud Run services](https://www.instana.com/docs/ecosystem/google-cloud-run/#configure-your-cloud-run-service)

Services running in serverless environments don't use host agent to send metrics and trace data to Instana backend, therefore the usual way of configuring the in-app sensor via [`configuration.yaml`](https://www.instana.com/docs/setup_and_manage/host_agent/configuration/#agent-configuration-file) file is not applicable. Instead, there is a set of environment variables that can optionally be configured in service task definition:

| Environment variable         | Default value                              | Description                                                                              |
|------------------------------|--------------------------------------------|------------------------------------------------------------------------------------------|
| `INSTANA_TIMEOUT`            | `500`                                      | The Instana backend connection timeout (in milliseconds)                                 |
| `INSTANA_SECRETS`            | `contains-ignore-case:secret,key,password` | The [secrets filter](#secrets-filtering) (also applied to process environment variables) |
| `INSTANA_EXTRA_HTTP_HEADERS` | none                                       | A semicolon-separated list of HTTP headers to collect from incoming requests             |
| `INSTANA_ENDPOINT_PROXY`     | none                                       | A proxy URL to use to connect to Instana backend                                         |
| `INSTANA_TAGS`               | none                                       | A comma-separated list of tags with optional values to associate with the ECS task       |
| `INSTANA_ZONE`               | `<Current AWS availability zone>`          | A custom Instana zone name for this service                                              |

Please refer to [Instana documentation](https://www.instana.com/docs/reference/environment_variables/#serverless-monitoring) for more detailed description of these variables and their value format.

### Using Instana to gather process metrics only

To use sensor without tracing ability, import the `instana` package and add the following line at the beginning of your `main()` function:

```go
instana.InitSensor(opt)
```

## Common Operations

The Instana Go sensor offers a set of quick features to support tracing of the most common operations like handling HTTP requests and executing HTTP requests.

To create an instance of the Instana sensor just request a new instance using the `instana.NewSensor` factory method and providing the name of the application. It is recommended to use a single instance only. The sensor implementation is fully thread-safe and can be shared by multiple threads.

```go
var sensor = instana.NewSensor("my-service")
```

A full example can be found under the examples folder in [example/http-database-greeter/](./example/http-database-greeter).

### Setting the sensor log output

The Go sensor uses a leveled logger to log internal errors and diagnostic information. The default `logger.Logger` uses `log.Logger`
configured with `log.Lstdflags` as a backend and writes messages to `os.Stderr`. By default, this logger only prints out the `ERROR` level
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
and enable the debug logging while configuring your logger:

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

### Secrets Filtering

Certain instrumentation modules provided by the Go sensor package, e.g. the [HTTP servers and clients](#http-servers-and-clients) wrappers, collect data that may contain sensitive information, such as passwords, keys and secrets. To avoid leaking these values the Go sensor replaces them with `<redacted>` before sending to the agent. The list of parameter name matchers is defined in `com.instana.secrets` section of the [Host Agent Configuration file](https://www.instana.com/docs/setup_and_manage/host_agent/configuration/#secrets) and will be sent to the in-app tracer during the announcement phase (requires agent Go trace plugin `com.instana.sensor-golang-trace` v1.3.0 and above).

The default setting for the secrets matcher is `contains-ignore-case` with following list of terms: `key`, `password`, `secret`. This would redact the value of a parameter which name _contains_ any of these strings ignoring the case.

### HTTP servers and clients

The Go sensor module provides instrumentation for clients and servers that use `net/http` package. Once activated (see below) this
instrumentation automatically collects information about incoming and outgoing requests and sends it to the Instana agent. See the [instana.HTTPSpanTags][instana.HTTPSpanTags] documentation to learn which call details are collected.

#### Instrumenting HTTP request handling

With support to wrap a `http.HandlerFunc`, Instana quickly adds the possibility to trace requests and collect child spans, executed in the context of the request span.

Minimal changes are required for Instana to be able to capture the necessary information. By simply wrapping the currently existing `http.HandlerFunc` Instana collects and injects necessary information automatically.

That said, a simple handler function like the following will simple be wrapped and registered like normal.

The following example code demonstrates how to instrument an HTTP handler using `instana.TracingHandlerFunc()`:

```go
sensor := instana.NewSensor("my-http-server")

http.HandleFunc("/", instana.TracingHandlerFunc(sensor, "/", func(w http.ResponseWriter, req *http.Request) {
	// Extract the parent span and use its tracer to initialize any child spans to trace the calls
	// inside the handler, e.g. database queries, 3rd-party API requests, etc.
	if parent, ok := instana.SpanFromContext(req.Context()); ok {
		sp := parent.Tracer().StartSpan("index")
		defer sp.Finish()
	}

	// ...
}))
```

In case your handler is implemented as a `http.Handler`, pass its `ServeHTTP` method instead:

```go
h := http.FileServer(http.Dir("./"))
http.HandleFunc("/files", instana.TracingHandlerFunc(sensor, "index", h.ServeHTTP))
```

#### Instrumenting HTTP request execution

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

#### Capturing custom HTTP headers

The HTTP instrumentation wrappers are capable of collecting HTTP headers and sending them along with the incoming/outgoing request spans. The list of case-insensitive header names can be provided both within `(instana.Options).Tracer.CollectableHTTPHeaders` field of the options object passed to `instana.InitSensor()` and in the [Host Agent Configuration file](https://www.instana.com/docs/setup_and_manage/host_agent/configuration/#capture-custom-http-headers). The latter setting takes precedence and requires agent Go trace plugin `com.instana.sensor-golang-trace` v1.3.0 and above:

```go
instana.InitSensor(&instana.Options{
	// ...
	Tracer: instana.TracerOptions{
		// ...
		CollectableHTTPHeaders: []string{"x-request-id", "x-loadtest-id"},
	},
})
```

This configuration is an equivalent of following settings in the [Host Agent Configuration file](https://www.instana.com/docs/setup_and_manage/host_agent/configuration/#capture-custom-http-headers):

```
com.instana.tracing:
  extra-http-headers:
    - 'x-request-id'
    - 'x-loadtest-id'
```

By default, the HTTP instrumentation does not collect any headers.

### Database Calls

The Go sensor provides `instana.InstrumentSQLDriver()` and `instana.WrapSQLConnector()` (since Go v1.10+) to instrument SQL database calls made with `database/sql`. The tracer will then automatically capture the `Query` and `Exec` calls, gather information about the query, such as statement, execution time, etc. and forward them to be displayed as a part of the trace.

#### Instrumenting `sql.Open()`

To instrument a database driver, register it using `instana.InstrumentSQLDriver()` first and replace the call to `sql.Open()` with `instana.SQLOpen()`. Here is an example on how to do this for `github.com/lib/pq` PostgreSQL driver:

```go
// Create a new instana.Sensor instance
sensor := instana.NewSensor("my-database-app")

// Instrument the driver
instana.InstrumentSQLDriver(sensor, "postgres", &pq.Driver{})

// Create an instance of *sql.DB to use for database queries
db, err := instana.SQLOpen("postgres", "postgres://...")
```

You can find the complete example in the [Examples section][godoc] of package documentation on [pkg.go.dev][pkg.go.dev].

The instrumented driver is registered with the name `<original_name>_with_instana`, e.g. in the example above the name would be `postgres_with_instana`.

#### Instrumenting `sql.OpenDB()`

Starting from Go v1.10 `database/sql` provides a new way to initialize `*sql.DB` that does not require the use of global driver registry. If the database driver library provides a type that satisfies the `database/sql/driver.Connector` interface, it can be used to create a database connection.

To instrument a `driver.Connector` instance, wrap it using `instana.WrapSQLConnector()`. Here is an example on how this can be done for `github.com/go-sql-driver/mysql/` MySQL driver:

```go
// Create a new instana.Sensor instance
sensor := instana.NewSensor("my-database-app")

// Initialize a new connector
connector, err := mysql.NewConnector(cfg)
// ...

// Wrap the connector before passing it to sql.OpenDB()
db, err := sql.OpenDB(instana.WrapSQLConnector(sensor, "mysql://...", connector))
```

You can find the complete example in the [Examples section][godoc] of package documentation on [pkg.go.dev][pkg.go.dev].

### GRPC servers and clients

[`github.com/instana/go-sensor/instrumentation/instagrpc`](./instrumentation/instagrpc) provides both unary and stream interceptors to instrument GRPC servers and clients that use `google.golang.org/grpc`.

### Kafka producers and consumers

[`github.com/instana/go-sensor/instrumentation/instasarama`](./instrumentation/instasarama) provides both unary and stream interceptors to instrument Kafka producers and consumers built on top of `github.com/Shopify/sarama`.

## OpenTracing

Instana tracer provides an interface compatible with [`github.com/opentracing/opentracing-go`](https://github.com/opentracing/opentracing-go) and thus can be used as a global tracer. However, the recommended approach is to use the Instana wrapper packages/functions [provided](./instrumentation) in the library. They set up a lot of semantic information which helps Instana get the best picture of the application possible. Sending proper tags is especially important when it comes to correlating calls to infrastructure and since they are strings mostly, there is a large room for making a mistake.

In case you want to integrate Instana into an app that is already instrumented with OpenTracing, register an instance of Instana tracer as a global tracer at the beginning of your `main()` function:

```go
import (
	instana "github.com/instana/go-sensor"
	opentracing "github.com/opentracing/opentracing-go"
)

func main() {
	opentracing.InitGlobalTracer(instana.NewTracerWithOptions(instana.DefaultOptions())
	// ...
}
```

This will automatically initialize the sensor and thus also activate the metrics stream. The tracer takes the same options that the sensor takes for initialization, described above.

The tracer is able to protocol and piggyback OpenTracing baggage, tags and logs. Only text mapping is implemented yet, binary is not supported. Also, the tracer tries to map the OpenTracing spans to the Instana model based on OpenTracing recommended tags. See [the Instana OpenTracing integration example](./example/opentracing) for details on how recommended tags are used.

The Instana tracer will remap OpenTracing HTTP headers into Instana headers, so parallel use with some other OpenTracing model is not possible. The Instana tracer is based on the OpenTracing Go basictracer with necessary modifications to map to the Instana tracing model. Also, sampling isn't implemented yet and will be focus of future work.

## W3C Trace Context

The Go sensor library fully supports [the W3C Trace Context standard](https://www.w3.org/TR/trace-context/):

* An [instrumented `http.Client`][instana.RoundTripper] sends the `traceparent` and `tracestate` headers, updating them with the exit span ID and flags.
* Any `http.Handler` instrumented with [`instana.TracingHandlerFunc()`][instana.TracingHandlerFunc] picks up the trace context passed in the `traceparent` header, potentially restoring the trace from `tracestate` even if the upstream service is not instrumented with Instana.

## Events API

The sensor, be it instantiated explicitly or implicitly through the tracer, provides a simple wrapper API to send events to Instana as described in [its documentation](https://www.instana.com/docs/api/agent/#event-sdk-web-service).

To learn more, see the [Events API](./EventAPI.md) document in this repository.

## AutoProfile™

AutoProfile™ generates and reports process profiles to Instana. Unlike development-time and on-demand profilers, where a user must manually initiate profiling, AutoProfile™ automatically schedules and continuously performs profiling appropriate for critical production environments.

### Activation from within the application code

To enable continuous profiling for your service provide `EnableAutoProfile: true` while initializing the sensor:

```go
func main() {
	instana.InitSensor(&instana.Options{
		EnableAutoProfile: true,
		// ...other options
	})

	// ...
}
```

To temporarily turn AutoProfile™ on and off from your code, call `autoprofile.Enable()` and `autoprofile.Disable()`.

### Activation without code changes

To enable AutoProfile™ for an app without code changes, set `INSTANA_AUTO_PROFILE=true` env variable. Note that this value takes precedence and overrides any attempt to disable profiling from inside the application code.

## Examples

Following examples are included in the `example` folder:

* [Greeter](./example/http-database-greeter) - an instrumented HTTP server that queries a database
* [Doubler](./example/kafka-producer-consumer) - an instrumented Kafka processor, that consumes and produces messages
* [Event](./example/event) - Demonstrates usage of the Events API
* [Autoprofile](./example/autoprofile) - Demonstrates usage of the AutoProfile™
* [OpenTracing](./example/opentracing) - an example of usage of Instana tracer in an app instrumented with OpenTracing
* [gRPC](./example/grpc-client-server) - an example of usage of Instana tracer in an app instrumented with gRPC
* [Gin](./example/gin) - an example of usage of Instana tracer instrumenting a [`Gin`](github.com/gin-gonic/gin) application
* [Gorilla mux](./example/gorillamux) - an example of usage of Instana tracer instrumenting the [`github.com/gorilla/mux`](https://github.com/gorilla/mux) router
* [httprouter](./example/httprouter) - an example of usage of Instana tracer instrumenting a [`github.com/julienschmidt/httprouter`](https://github.com/julienschmidt/httprouter) router

For more examples please consult the [godoc][godoc].

[godoc]: https://pkg.go.dev/github.com/instana/go-sensor/?tab=doc#pkg-examples
[pkg.go.dev]: https://pkg.go.dev/github.com/instana/go-sensor
[instana.TracingHandlerFunc]: https://pkg.go.dev/github.com/instana/go-sensor/?tab=doc#TracingHandlerFunc
[instana.RoundTripper]: https://pkg.go.dev/github.com/instana/go-sensor/?tab=doc#RoundTripper
[instana.HTTPSpanTags]: https://pkg.go.dev/github.com/instana/go-sensor/?tab=doc#HTTPSpanTags
[instana.Options]: https://pkg.go.dev/github.com/instana/go-sensor/?tab=doc#Options
[instana.TracerOptions]: https://pkg.go.dev/github.com/instana/go-sensor/?tab=doc#TracerOptions
