# OpenTracing → OpenTelemetry Migration Analysis

## Overview

The goal of this analysis is to understand how difficult it would be to migrate the Instana Go Sensor from OpenTracing to OpenTelemetry while keeping the existing API unchanged.

Overall, the migration is definitely possible because the current concepts already align closely with OpenTelemetry.

## Component Mapping

- tracer.go → trace.Tracer
- span.go → trace.Span
- span_context.go → trace.SpanContext
- propagation.go → TextMapPropagator
- recorder.go → SpanExporter
- instrumentation_http.go → OTel HTTP middleware
- context.go → OpenTelemetry Context APIs

## Key Findings

- Existing W3C Trace Context support already exists.
- Instana already supports 128-bit trace identifiers.
- propagation.go maps closely to OpenTelemetry's TextMapPropagator.
- recorder.go maps naturally to a SpanExporter.
- Context propagation already follows OpenTelemetry patterns.

## Main Challenges

### ID Mapping

Instana and OpenTelemetry use different SpanContext implementations. Although Instana already supports 128-bit trace IDs, there is still work needed to understand how the two models can be mapped to each other.
Risk: Medium

### Propagation

Current Instana headers (x-instana-t, x-instana-s, x-instana-l) map well to OpenTelemetry's propagation model.
Risk: Low

### Span Exporting

A custom SpanExporter will be required to convert OpenTelemetry spans into the format expected by the Instana agent.
Risk: Medium

### Span Suppression

The current implementation uses SpanContext.Suppressed and x-instana-l. OpenTelemetry does not provide an exact equivalent.
Risk: Medium

### Backward Compatibility

Preserving existing APIs such as InitCollector(), NewTracer(), and ContextWithSpan() is the most important requirement.
Risk: High

## Recommendation

Migrating gradually is recommended.

Suggested approach:

1. Introduce OpenTelemetry dependencies.
2. Implement a custom TextMapPropagator.
3. Implement a custom SpanExporter.
4. Create OpenTelemetry-based HTTP middleware.
5. Gradually migrate additional instrumentation packages.

Overall, the migration looks technically achievable, with backward compatibility representing the primary risk area.

## POC Approach

### otel_propagator.go

Implement a custom TextMapPropagator for:

- x-instana-t
- x-instana-s
- x-instana-l

### otel_exporter.go

Implement a SpanExporter that converts OpenTelemetry spans and forwards them to the existing Instana agent.

### otel_instrumentation_http.go

Implement an OpenTelemetry version of TracingHandlerFunc using OpenTelemetry propagation and tracing APIs.