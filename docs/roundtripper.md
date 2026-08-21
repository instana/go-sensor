## Tracing HTTP Outgoing Requests

The tracer is able to collect data from the Go standard library for outgoing HTTP requests.
That is, when one creates an [http.Client](https://pkg.go.dev/net/http@go1.21.3#Client) instance to make calls to HTTP servers.

To achieve this data collection we provide a [RoundTripper](https://pkg.go.dev/net/http@go1.21.3#RoundTripper) wrapper to be used as the http.Client Transport.
Additionally, as HTTP outgoing requests are exit spans, the HTTP request must be attached to a context containing a entry span.

### Usage

Let's assume the following code snippet as an example to be traced:

```go
client := &http.Client{}

req, err := http.NewRequest(http.MethodGet, "https://www.instana.com", nil)

_, err = client.Do(req)

```

The first thing we need is to add the collector to the project:

```go
col := instana.InitCollector(&instana.Options{
  Service: "my-http-client",
  Tracer:  instana.DefaultTracerOptions(),
})
```

Then, we need to wrap the current Transport (if any) with `instana.RoundTripper`.
If no Transport is provided, simply pass `nil` as the second argument.
The `instana.RoundTripper` wrapper will intercept the relevant information from the HTTP request, such as URL, methods and so on. It will then collect them and send to the Agent periodically.

```go
// Wrap the original http.Client transport with instana.RoundTripper().
// The http.DefaultTransport will be used if there was no transport provided.
client := &http.Client{
  Transport: instana.RoundTripper(col, nil),
}
```

Usually, your application will have an entry span already, received via HTTP or other type of incoming request.
This span should be contained into the context, which needs to be passed ahead to your client HTTP request.
Also, make sure to finish the span in order to send it to the Agent.

```go
// Inject the parent span into request context
ctx := instana.ContextWithSpan(context.Background(), entrySpan)

// Use your instrumented http.Client to propagate tracing context with the request
_, err = client.Do(req.WithContext(ctx))
```

If you do not have an entry span as explained above or you are not sure, it's possible to manually create an entry span and attach it to a context to be passed forward to your HTTP client:

```go
// Every call should start with an entry span:
// https://www.ibm.com/docs/en/instana-observability/current?topic=tracing-best-practices#start-new-traces-with-entry-spans
// Normally this would be your HTTP/GRPC/message queue request span, but here we need to create it explicitly,
// since an HTTP client call is an exit span. And all exit spans must have a parent entry span.
entrySpan := col.Tracer().StartSpan("client-call")
entrySpan.SetTag(string(ext.SpanKind), "entry")

...

// Inject the parent span into request context
ctx := instana.ContextWithSpan(context.Background(), entrySpan)

// Use your instrumented http.Client to propagate tracing context with the request
_, err = client.Do(req.WithContext(ctx))

...

// Remember to always finish spans that were created manually to make sure it's propagated to the Agent.
// In this case, we want to make sure that the entry span is finished after the HTTP request is completed.
// Optionally, we could use defer right after the span is created.
entrySpan.Finish()
```
You can learn more about manually instrumenting your code [here]().

#### Complete Example

```go
package main

import (
  "context"
  "log"
  "net/http"

  instana "github.com/instana/go-sensor"
  "github.com/opentracing/opentracing-go/ext"
)

func main() {
  col := instana.InitCollector(&instana.Options{
    Service: "my-http-client",
    Tracer:  instana.DefaultTracerOptions(),
  })

  client := &http.Client{
    Transport: instana.RoundTripper(col, nil),
  }

  entrySpan := col.Tracer().StartSpan("client-call")
  entrySpan.SetTag(string(ext.SpanKind), "entry")

  req, err := http.NewRequest(http.MethodGet, "https://www.instana.com", nil)
  if err != nil {
    log.Fatalf("failed to create request: %s", err)
  }

  ctx := instana.ContextWithSpan(context.Background(), entrySpan)

  _, err = client.Do(req.WithContext(ctx))
  if err != nil {
    log.Fatalf("failed to GET https://www.instana.com: %s", err)
  }

  entrySpan.Finish()
}
```

### Error Handling

The `RoundTripper` marks an exit span as an error (`span.ec=1`) in the following situations:

**Transport errors** — when the request never receives a response (network failure, DNS error, timeout):
the span is marked as an error and `span.data.http.error` is populated with the Go error message.

**5xx responses** — always treated as errors, regardless of any configuration:

| Status | `span.ec` | `span.data.http.error` |
|---|---|---|
| 500 Internal Server Error | 1 | `"Internal Server Error"` |
| 503 Service Unavailable | 1 | `"Service Unavailable"` |

**4xx responses** — not treated as errors by default. Opt in via configuration:

| Configuration | Effect |
|---|---|
| Default (nothing set) | All 4xx → `ec=0`, no error |
| `classify-all-4xx-as-errors: true` | All 4xx → `ec=1` |
| `classify-as-errors: [401, 403]` | Only listed codes → `ec=1`; others stay `ec=0` |

When a 4xx code is classified as an error, `span.data.http.error` is set to `"<code> <text>"`, for example `"401 Unauthorized"`.

See [HTTP 4xx error opt-in](http-4xx-error-opt-in.md) for the full configuration reference (environment variables, config file, agent config, and in-code options).

> **Entry spans are never affected.** The `RoundTripper` only controls exit (outbound) spans. The server-side entry span always follows its own rules: 5xx → error, 4xx → never an error.

-----
[README](../README.md) |
[Tracer Options](options.md) |
[Tracing SQL Driver Databases](sql.md) |
[Tracing Other Go Packages](other_packages.md) |
[Instrumenting Code Manually](manual_instrumentation.md)
