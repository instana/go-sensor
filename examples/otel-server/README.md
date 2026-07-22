# OpenTelemetry POC Server

This example demonstrates the small OpenTelemetry proof-of-concept created during the OpenTracing to OpenTelemetry migration investigation.

The example uses the proposed `OTelTracingHandlerFunc()` middleware to create an OpenTelemetry span for incoming HTTP requests.

## Run

From the repository root:

```bash
go run examples/otel-server/main.go
```

## Test

Open another terminal and run:

```bash
curl http://localhost:8080
```

Expected response:

```text
Hello from the OpenTelemetry POC
```

## What This Demonstrates

This example shows how the existing HTTP instrumentation model could be represented using OpenTelemetry APIs.

The proof-of-concept includes:

- `otel_propagator.go`
- `otel_exporter.go`
- `otel_instrumentation_http.go`

The goal is to demonstrate the migration approach rather than provide a production-ready implementation.

## Notes

This example was created as part of the OpenTracing → OpenTelemetry migration investigation for the Instana Go Sensor.