# HTTP Client Instrumentation Example

This example demonstrates how to instrument outgoing HTTP requests using Instana Go sensor's [`RoundTripper`](../../docs/roundtripper.md) wrapper.

## Overview

The example shows how to:
- Initialize the Instana collector
- Wrap an HTTP client with `instana.RoundTripper()` for automatic tracing
- Create an entry span for the HTTP client call
- Propagate trace context through the request

## Prerequisites

- Go 1.23 or later
- Instana Agent running (or configured backend)

## Running the Example

1. **Install dependencies:**
   ```bash
   go mod download
   ```

2. **Run the application:**
   ```bash
   go run main.go
   ```

   By default, the application makes a request to `https://example.com`.

3. **Specify a custom server URL:**
   ```bash
   go run main.go -s https://www.instana.com
   ```

4. **Stop the application:**
   Press `Ctrl+C` to gracefully shutdown.


## Documentation

For detailed information about tracing HTTP outgoing requests, see the [RoundTripper documentation](../../docs/roundtripper.md).

## Key Code Snippet

```go
// Initialize collector
collector := instana.InitCollector(&instana.Options{
    Service: "http-client",
    Tracer:  instana.DefaultTracerOptions(),
})

// Wrap HTTP client
client := &http.Client{
    Transport: instana.RoundTripper(collector, nil),
}

// Create entry span
span := collector.Tracer().StartSpan("http-client-call")
span.SetTag(string(ext.SpanKind), "entry")

// Attach span to context and execute request
ctx := instana.ContextWithSpan(context.Background(), span)
resp, err := client.Do(req.WithContext(ctx))

// Finish span
span.Finish()