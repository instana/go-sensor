# IBM Instana Go Tracer

[![Build Status](https://circleci.com/gh/instana/go-sensor/tree/main.svg?style=svg)](https://circleci.com/gh/instana/go-sensor/tree/main)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/instana/go-sensor)][pkg.go.dev]
[![OpenTracing](https://img.shields.io/badge/OpenTracing-enabled-blue.svg)](http://opentracing.io)
[![Go Report Card](https://goreportcard.com/badge/github.com/instana/go-sensor)](https://goreportcard.com/report/github.com/instana/go-sensor)

The IBM Instana Go Tracer is an SDK that collects traces, metrics, logs and provides profiling for Go applications. The tracer is part of the [IBM Instana Observability](https://www.ibm.com/products/instana) tool set.

## Compatibility

### Supported Runtimes
-----
- Go Collector **v1.70** or later supports Go **1.24** and **1.25**, and maintains compatibility with *Go 1.23 (EOL)*.

> [!NOTE]
> Make sure to always use the latest version of the tracer, as it provides new features, improvements, security updates and fixes.

## Installation

To add the tracer to your project, run:

```bash
go get -u github.com/instana/go-sensor@latest
```

> [!NOTE]
> As a good practice, add this command to your CI pipeline or your automated tool before building the application to keep the tracer up to date.

## Usage

### Initial Setup

Once the tracer is added to the project, import the package into the entrypoint file of your application:

```go
import (
  ...
  instana "github.com/instana/go-sensor"
)
```

Create a reference to the collector and initialize it with a service name:

```go
var (
  ...
  col instana.TracerLogger
)

func init() {
  ...
  col = instana.InitCollector(&instana.Options{
    Service: "My app",
    Tracer:  instana.DefaultTracerOptions(),
  })
}
```

> [!NOTE]
> The tracer expects the Instana Agent to be up and running in the default port 42699. You can change the port with the environment variable ``INSTANA_AGENT_PORT``.

> [!NOTE]
> For non default options, like the Agent host and port, the tracer can be configured either via SDK options, environment variables or Agent options.

### Collecting Metrics

Once the collector has been initialized with `instana.InitCollector`, application metrics such as memory, CPU consumption, active goroutine count etc will be automatically collected and reported to the Agent without further actions or configurations to the SDK.
This data is then already available in the dashboard.

### Tracing Calls

Let's collect traces of calls received by an HTTP server.

Before any changes, your code should look something like this:

```go
// endpointHandler is the standard http.Handler function
http.HandleFunc("/endpoint", endpointHandler)

log.Fatal(http.ListenAndServe(":9090", nil))
```

Wrap the `endpointHandler` function with `instana.TracingHandlerFunc`. Now your code should look like this:

```go
// endpointHandler is now wrapped by `instana.TracingHandlerFunc`
http.HandleFunc("/endpoint", instana.TracingHandlerFunc(col, "/endpoint", endpointHandler))

log.Fatal(http.ListenAndServe(":9090", nil))
```

When running the application, every time `/endpoint` is called, the tracer will collect this data and send it to the Instana Agent.
You can monitor traces to this endpoint in the Instana UI.

### Profiling

Unlike metrics, profiling needs to be enabled with the `EnableAutoProfile` option, as seen here:

```go
col = instana.InitCollector(&instana.Options{
  Service: "My app",
  EnableAutoProfile: true,
  Tracer:  instana.DefaultTracerOptions(),
})
```

You should be able to see your application profiling in the Instana UI under Analytics/Profiles.

### Logging

In terms of logging, the SDK provides two distinct logging features:

1. Traditional logging, that is, logs reported to the standard output, usually used for debugging purposes
1. Instana logs, a feature that allows customers to report logs to the dashboard under Analytics/Logs

#### Traditional Logging

Many logs are provided by the SDK, usually prefixed with "INSTANA" and are useful to understand what the tracer is doing underneath. It can also be used for debugging and troubleshoot reasons.
Customers can also provide logs by calling one of the following: [Collector.Info()](https://pkg.go.dev/github.com/instana/go-sensor#Collector.Info), [Collector.Warn()](https://pkg.go.dev/github.com/instana/go-sensor#Collector.Warn), [Collector.Error()](https://pkg.go.dev/github.com/instana/go-sensor#Collector.Error), [Collector.Debug()](https://pkg.go.dev/github.com/instana/go-sensor#Collector.Debug). You can setup the log level via options or the `INSTANA_LOG_LEVEL` environment variable.

You can find detailed information in the [Instana documentation](https://www.ibm.com/docs/en/instana-observability/current?topic=technologies-monitoring-go#tracers-logs).

#### Instana Logs

Instana Logs are spans of the type `log.go` that are rendered in a special format in the dashboard.
You can create logs and report them to the agent or attach them as children of an existing span.

The code snippet below shows how to create logs and send them to the agent:

```go
col := instana.InitCollector(&instana.Options{
  Service: "My Go App",
  Tracer:  instana.DefaultTracerOptions(),
})

col.StartSpan("log.go", []ot.StartSpanOption{
  ot.Tags{
    "log.level":   "error", // available levels: info, warn, error, debug
    "log.message": "error from log.go span",
  },
}...).Finish() // make sure to "finish" the span, so it's sent to the agent
```

This log can then be visualized in the dashboard under Analytics/Logs. You can add a filter by service name. In our example, the service name is "My Go App".

### Opt-in Exit Spans

 Go tracer support the opt-in feature for the exit spans. When enabled, the collector can start capturing exit spans, even without an entry span. This capability is particularly useful for scenarios like cronjobs and other background tasks, enabling the users to tailor the tracing according to their specific requirements. By setting the `INSTANA_ALLOW_ROOT_EXIT_SPAN` variable, users can choose whether the tracer should start a trace with an exit span or not. The environment variable can have 2 values. (1: Tracer should record exit spans for the outgoing calls, when it has no active entry span. 0 or any other values: Tracer should not start a trace with an exit span).

 ```bash
export INSTANA_ALLOW_ROOT_EXIT_SPAN=1
 ```

### Complete Example

[Basic Usage](./example/basic_usage/main.go)
```go
package main

import (
  "log"
  "net/http"

  instana "github.com/instana/go-sensor"
)

func main() {
  col := instana.InitCollector(&instana.Options{
    Service:           "Basic Usage",
    EnableAutoProfile: true,
    Tracer:  instana.DefaultTracerOptions(),
  })

  http.HandleFunc("/endpoint", instana.TracingHandlerFunc(col, "/endpoint", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
  }))

  log.Fatal(http.ListenAndServe(":7070", nil))
}
```

### Wrapping up

Let's quickly summarize what we have seen so far:

1. We learned how to install, import and initialize the Instana Go Tracer.
1. Once the tracer is initialized, application metrics are collected out of the box.
1. Application profiling can be enabled via the `EnableAutoProfile` option.
1. Tracing incoming HTTP requests by wrapping the Go standard library `http.Handler` with `instana.TracingHandlerFunc`.

With this knowledge it's already possible to make your Go application traceable by our SDK.
But there is much more you can do to enhance tracing for your application.

The basic functionality covers tracing for the following standard Go features:

1. HTTP incoming requests
1. HTTP outgoing requests
1. SQL drivers

As we already covered HTTP incoming requests, we suggest that you understand how to collect data from HTTP outgoing requests and SQL driver databases.

Another interesting feature is the usage of additional packages located under [instrumentation](./instrumentation/). Each of these packages provide tracing for specific Go packages like the AWS SDK, Gorm and Fiber.

## What's Next

1. [Tracer Options](docs/options.md)
1. [Tracing HTTP Outgoing Requests](docs/roundtripper.md)
1. [Tracing SQL Driver Databases](docs/sql.md)
1. [Tracing an application running on Azure Container Apps](docs/azure_container_apps.md)
1. [Tracing Other Go Packages](docs/other_packages.md)
1. [Instrumenting Code Manually](docs/manual_instrumentation.md)
1. [Disabling Spans by Category](docs/disabling_spans.md)
1. [Generic Serverless Agent](/docs/generic_serverless_agent.md)

<!-- Links section -->

[godoc]: https://pkg.go.dev/github.com/instana/go-sensor/?tab=doc#pkg-examples
[pkg.go.dev]: https://pkg.go.dev/github.com/instana/go-sensor
[docs.autoprofile]: https://www.ibm.com/docs/en/obi/current?topic=technologies-monitoring-go#instana-autoprofile%E2%84%A2
[docs.configuration]: https://www.ibm.com/docs/en/obi/current?topic=go-collector-configuration
[docs.installation]: https://www.ibm.com/docs/en/obi/current?topic=go-collector-installation
[docs.howto.configuration]: https://www.ibm.com/docs/en/obi/current?topic=go-collector-common-operations#configuration
[docs.howto.instrumentation]: https://www.ibm.com/docs/en/obi/current?topic=go-collector-common-operations#instrumentation
[instana.DefaultOptions]: https://pkg.go.dev/github.com/instana/go-sensor#DefaultOptions
