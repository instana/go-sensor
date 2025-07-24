## Instrumenting Code Manually

The IBM Instana Go Tracer is built on top of the [Opentracing SDK](https://github.com/opentracing/opentracing-go).
In practical terms this means that we provide concrete implementations for Opentracing's tracer, span interfaces and other required implementations to fulfill the SDK.

All these concrete implementations are publicly available and can be used by anyone.

The main difference between customers creating their own traces and using our SDK is that we encapsulate all the boilerplate code and data collection logic within the tracer and additional packages instrumentation for your convenience.

However, customers can extend it to create custom spans or provide extra tags with our SDK.

This section is dedicated to explore the creation of custom spans to be sent to the Instana Agent.


### Understanding Entry Spans

A trace should always start with an entry span, which should be the parent span of subsequent spans.

If a span is created with the type `intermediate` or `exit` and it is sent to the Agent, the UI will provide a hollow entry span automatically called `Internal Trigger`.
However, our tracer limits its own spans (which is, not custom spans) to always require an entry span.

A common use case is the instrumentation of outgoing HTTP requests with our SDK.
These requests are `exit` spans, as they are "exiting" your application and require an entry span.
Often this entry span will be an incoming HTTP request or a receiving queue message. But if this is not the case, an entry span must be manually created and attached as the parent span of the outgoing HTTP request (exit) span.

### Creating and Reporting Spans

You can create an entry span with the `StartSpan` method and by providing `ext.SpanKindRPCServer` as one of the options:

```go
package main

import (
  instana "github.com/instana/go-sensor"
  ot "github.com/opentracing/opentracing-go"
  "github.com/opentracing/opentracing-go/ext"
)

func main() {
  col := instana.InitCollector(&instana.Options{
    Service: "My Service",
    Tracer:  instana.DefaultTracerOptions(),
  })

  ps := col.StartSpan("my-entry-span", []ot.StartSpanOption{
    ext.SpanKindRPCServer,
  }...)

  // Do some work

  // Always make sure to call Finish to send the span to the Agent.
  ps.Finish()
}
```

The `StartSpan` method is compliant to the Opentracing interfaces, so you can rely on the Opentracing elements, such as `ot.StartSpanOption` for the span options or `ot.Tags` as part of the span options.
You can also notice the usage of Opentracing's `ext.SpanKindRPCServer` to define the span as an entry span.

> [!NOTE]
> You can use Opentracing's `ext.SpanKindRPCClient` to define an exit span. If no kind is provided, the span is assumed to be an `intermediate` span.


### Correlating Spans

If you wish to define a relation for a span with respect to another span, so that multiple spans can be correlated to each other to form a more elaborate trace, you can use the method `ot.ChildOf()` as one of the options for child spans.

```go
package main

import (
  instana "github.com/instana/go-sensor"
  ot "github.com/opentracing/opentracing-go"
  "github.com/opentracing/opentracing-go/ext"
)

func main() {
  col := instana.InitCollector(&instana.Options{
    Service: "My Service",
    Tracer:  instana.DefaultTracerOptions(),
  })

  ps := col.StartSpan("my-parent-entry-span", []ot.StartSpanOption{
    ext.SpanKindRPCServer,
  }...)

  // Do some work

  ps.Finish()

  exs := col.StartSpan("my-child-exit-span", []ot.StartSpanOption{
    ext.SpanKindRPCClient,

    // Make sure to provide the parent span context to ot.ChildOf in order to correlate these spans
    ot.ChildOf(ps.Context()),
  }...)

  // Do some work

  exs.Finish()
}
```

-----
[README](../README.md) |
[Tracer Options](options.md) |
[Tracing HTTP Outgoing Requests](roundtripper.md) |
[Tracing SQL Driver Databases](sql.md) |
[Tracing Other Go Packages](other_packages.md)
