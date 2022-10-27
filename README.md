# Instana Go Collector

![Instana, an IBM company](https://user-images.githubusercontent.com/203793/135623131-0babc5b4-7599-4511-8bf0-ce05922de8a3.png)

[![Build Status](https://circleci.com/gh/instana/go-sensor/tree/master.svg?style=svg)](https://circleci.com/gh/instana/go-sensor/tree/master)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/instana/go-sensor)][pkg.go.dev]
[![OpenTracing](https://img.shields.io/badge/OpenTracing-enabled-blue.svg)](http://opentracing.io)
[![Go Report Card](https://goreportcard.com/badge/github.com/instana/go-sensor)](https://goreportcard.com/report/github.com/instana/go-sensor)

The Go Collector is a runtime metrics collector, code execution tracer and profiler for applications and services written in Go. This module
is a part of [Instana](https://instana.com) APM solution.

Since version 1.47 the Go Collector requires Go version 1.13 or later.

## Installation

To add Instana Go Collector to your service run:

```bash
$ go get github.com/instana/go-sensor
```

You might also consider installing [supplemental modules](https://www.ibm.com/docs/en/obi/current?topic=technologies-monitoring-go#supported-frameworks-and-libraries)
that provide instrumentation for most popular 3rd-party packages.

Please refer to [Instana Go Collector documentation][docs.installation] for further details on how to activate Go Collector and use it to
instrument your application code.

## Configuration

The Go Collector accepts both configuration from within the application code and via environment variables. The values provided via enironment
variables take precedence. In case is no specific configuration provided, the values returned by
[instana.DefaultOptions()][instana.DefaultOptions] will be used.

Please refer to [Go Collector Configuration page][docs.configuration] for detailed instructions. There is also the
[Go Collector How To page][docs.howto.configuration] that covers the most common configuration use cases.

## Usage

In order to trace the code execution, a few minor changes to your app's source code is needed. Please check the [examples section](#examples)
and the [Go Collector How To guide][docs.howto.instrumentation] to learn about common instrumentation patterns.

## Features

### Runtime metrics collection

Once [initialized](https://www.ibm.com/docs/en/obi/current?topic=go-collector-common-operations#how-to-initialize-go-collector), the Go Collector starts automatically
collecting and sending the following runtime metrics to Instana in background:

* Memory usage
* Heap usage
* GC activity
* Goroutines

### Code execution tracing

Instana Go Collector provides an API to [instrument][docs.howto.instrumentation] function and method calls from within the application code
to trace its execution.

The core `github.com/instana/go-sensor` package is shipped with instrumentation wrappers for the standard library, including HTTP client and
server, as well as SQL database drivers compatible with `database/sql`. There are also supplemental
[instrumentation modules](https://www.ibm.com/docs/en/obi/current?topic=technologies-monitoring-go#supported-frameworks-and-libraries) provide code wrappers to instrument
the most popular 3rd-party libraries.

Please check the [examples section](#examples) and the [Go Collector How To guide][docs.howto.instrumentation] to learn about common
instrumentation patterns.

#### OpenTracing

Instana Go Collector provides an interface compatible with [`github.com/opentracing/opentracing-go`](https://github.com/opentracing/opentracing-go) and thus can be used as a global tracer. However, the recommended approach is to use the Instana wrapper packages/functions [provided](./instrumentation) in the library. They set up a lot of semantic information which helps Instana get the best picture of the application possible. Sending proper tags is especially important when it comes to correlating calls to infrastructure and since they are strings mostly, there is a large room for making a mistake.

The Go Collector will remap OpenTracing HTTP headers into Instana headers, so parallel use with some other OpenTracing model is not possible. The Instana tracer is based on the OpenTracing Go basictracer with necessary modifications to map to the Instana tracing model.

### Trace continuation and propagation

Instana Go Collector ensures that application trace context will continued and propagated beyond the service boundaries using various
methods, depending on the technology being used. Alongside with Instana-specific HTTP headers or message attributes, a number of open
standards are supported, such as W3C Trace Context and OpenTelemetry.

#### W3C Trace Context & OpenTelemetry

The instrumentation wrappers provided with Go Collector automatically inject and extract trace context provided via W3C Trace Context HTTP
headers.

### Continuous profiling

[Instana AutoProfile™][docs.autoprofile] generates and reports process profiles to Instana. Unlike development-time and on-demand profilers,
where a user must manually initiate profiling, AutoProfile™ automatically schedules and continuously performs profiling appropriate for
critical production environments.

Please refer to the [Instana Go Collector docs](https://www.ibm.com/docs/en/obi/current?topic=go-collector-common-operations#instana-autoprofile%E2%84%A2) to learn how to activate and
use continuous profiling for your applications and services.

### Sending custom events

The Go Collector, be it instantiated explicitly or implicitly through the tracer, provides a simple wrapper API to send events to Instana as described in [its documentation](https://www.ibm.com/docs/en/obi/current?topic=integrations-sdks-apis).

To learn more, see the [Events API](./EventAPI.md) document in this repository.

## Examples

Following examples are included in the `example` folder:

* [Greeter](./example/http-database-greeter) - an instrumented HTTP server that queries a database
* [Doubler](./example/kafka-producer-consumer) - an instrumented Kafka processor, that consumes and produces messages
* [Event](./example/event) - Demonstrates usage of the Events API
* [Autoprofile](./example/autoprofile) - Demonstrates usage of the AutoProfile™
* [OpenTracing](./example/opentracing) - an example of usage of Instana tracer in an app instrumented with OpenTracing
* [gRPC](./example/grpc-client-server) - an example of usage of Instana tracer in an app instrumented with gRPC
* [Gin](./example/gin) - an example of usage of Instana tracer instrumenting a [`Gin`](github.com/gin-gonic/gin) application
* [httprouter](./example/httprouter) - an example of usage of Instana tracer instrumenting a [`github.com/julienschmidt/httprouter`](https://github.com/julienschmidt/httprouter) router

For more examples please consult the [godoc][godoc] and the [Go Collector How To page](https://www.ibm.com/docs/en/obi/current?topic=go-collector-common-operations).

## Filing Issues

If something is not working as expected or you have a question, instead of opening an issue in this repository, please open a ticket at [Instana Support portal](https://support.instana.com/hc/requests/new) instead.

<!-- Links section -->

[godoc]: https://pkg.go.dev/github.com/instana/go-sensor/?tab=doc#pkg-examples
[pkg.go.dev]: https://pkg.go.dev/github.com/instana/go-sensor
[docs.autoprofile]: https://www.ibm.com/docs/en/obi/current?topic=technologies-monitoring-go#instana-autoprofile%E2%84%A2
[docs.configuration]: https://www.ibm.com/docs/en/obi/current?topic=go-collector-configuration
[docs.installation]: https://www.ibm.com/docs/en/obi/current?topic=go-collector-installation
[docs.howto.configuration]: https://www.ibm.com/docs/en/obi/current?topic=go-collector-common-operations#configuration
[docs.howto.instrumentation]: https://www.ibm.com/docs/en/obi/current?topic=go-collector-common-operations#instrumentation
[instana.DefaultOptions]: https://pkg.go.dev/github.com/instana/go-sensor#DefaultOptions
