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

## Issues Encountered

### Issue 1: SpanContext Differences

**Problem**

Instana and OpenTelemetry don't represent trace information in exactly the same way, so there would need to be a way to move between the two models.

**Proposed Solution**

Introduce a conversion layer between Instana's SpanContext and OpenTelemetry's SpanContext.

One positive finding from the analysis is that Instana already supports 128-bit trace IDs, which should make this easier than starting from scratch.

---

### Issue 2: Existing Propagation Format

**Problem**

The current implementation relies on the existing Instana headers:

- x-instana-t
- x-instana-s
- x-instana-l

Changing these could create compatibility issues for services already using the SDK.

**Proposed Solution**

Keep the existing headers and implement a custom OpenTelemetry propagator around them.

This approach was tested in the POC through `otel_propagator.go`, which injects and extracts trace information using the current header format.

---

### Issue 3: Span Exporting

**Problem**

OpenTelemetry spans can't be sent directly to the Instana agent because the agent expects Instana span structures.

**Proposed Solution**

Add a custom SpanExporter that receives OpenTelemetry spans, converts them into Instana spans, and then forwards them to the existing agent.

This would allow the current export pipeline to remain largely unchanged.

---

### Issue 4: Backward Compatibility

**Problem**

Many users are already relying on the current SDK APIs and behaviour.

Examples include:

- `InitCollector()`
- `NewTracer()`
- `ContextWithSpan()`
- Existing propagation headers
- Existing instrumentation packages

Changing these would likely require application changes, which would increase migration risk.

**Proposed Solution**

My preference would be to keep the current public APIs and introduce OpenTelemetry behind the scenes.

The idea would be that existing applications continue working as they do today, while the underlying implementation gradually moves to OpenTelemetry.

For example:

- Existing APIs would remain available.
- Existing propagation headers would continue to work.
- Existing instrumentation would continue to function.
- OpenTelemetry would be introduced internally rather than exposed immediately to users.

This is one of the main reasons I'd favour a gradual migration over a full rewrite.

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