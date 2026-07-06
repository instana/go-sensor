# Agent guide for IBM Instana Go Tracer (`go-sensor`)

> This file is written for AI coding assistants. It describes the repository precisely enough to make high-quality contributions without speculation.

---

## 1. Repository Overview

- **What it is**: IBM Instana's Go tracing SDK. Collects traces, metrics, logs, and profiling data for Go applications.
- **Foundation**: Built on [OpenTracing](https://github.com/opentracing/opentracing-go) (`v1.2.0`). This is intentional and permanent — do **not** migrate to OpenTelemetry unless explicitly requested.
- **Instrumentation model**: Manual only. There is no automatic/bytecode instrumentation. Every instrumented library is a separate opt-in package under `instrumentation/`.
- **Non-goals**: automatic instrumentation, OpenTelemetry migration, bytecode manipulation.
- **Module path**: `github.com/instana/go-sensor`
- **Current version**: `1.73.x` (see [`version.go`](version.go))
- **Minimum Go**: `1.23` (core module `go.mod`); instrumentation packages may require newer versions via build tags.
- **Supported runtime targets**: Go 1.25 and 1.26 (actively tested); Go 1.24/1.23 compatibility maintained.

---

## 2. Architecture

### Core Packages (repo root)

| File / Package | Responsibility |
|---|---|
| [`sensor.go`](sensor.go) | Global `sensorS` singleton; initializes agent client, metrics, and auto-profiling. |
| [`collector.go`](collector.go) | `Collector` type and `InitCollector()` — the primary public entry point for users. |
| [`tracer.go`](tracer.go) | `tracerS` — implements `opentracing.Tracer`; creates and starts spans. |
| [`span.go`](span.go) | `spanS` — implements `opentracing.Span`; handles `Finish`, log records, error counting. |
| [`span_context.go`](span_context.go) | `SpanContext` — holds trace/span IDs, baggage, W3C context, EUM correlation. |
| [`propagation.go`](propagation.go) | Inject/extract of Instana headers (`x-instana-t/s/l/b-*`) and W3C `traceparent`/`tracestate`. |
| [`context.go`](context.go) | `ContextWithSpan` / `SpanFromContext` — store/retrieve active span in `context.Context`. |
| [`recorder.go`](recorder.go) | `Recorder` / `SpanRecorder` — buffers finished spans, flushes to agent on a 1-second ticker. |
| [`registered_span.go`](registered_span.go) | Constants for all known `RegisteredSpanType` values (e.g. `g.http`, `kafka`, `rpc-server`). |
| [`options.go`](options.go) | `Options` struct and `applyConfiguration()` — merges env vars, in-code, and agent config. |
| [`instrumentation_http.go`](instrumentation_http.go) | `TracingHandlerFunc` / `TracingNamedHandlerFunc` — HTTP server middleware. |
| [`instrumentation_sql.go`](instrumentation_sql.go) | SQL driver wrapping via `database/sql` hooks. |
| [`tags.go`](tags.go) | Tag constants and helper `Tags` interface. |
| [`adapters.go`](adapters.go) | `Tracer` interface (extends `opentracing.Tracer`), `Sensor` type (legacy adapter). |
| [`log.go`](log.go) | `LeveledLogger` interface; `SetLogger()`. |
| `agent.go`, `fsm.go` | Host agent discovery, heartbeat, FSM-based connection state machine. |
| `fargate_agent.go`, `lambda_agent.go`, `gcr_agent.go`, `azure_agent.go`, `generic_serverless_agent.go` | Serverless-specific agent implementations. |
| `w3ctrace/` | W3C Trace Context parsing/formatting. |
| `acceptor/` | HTTP acceptor (spans/metrics transport). |
| `autoprofile/` | Continuous profiling subsystem. |
| `logger/` | Default logger implementation. |
| `secrets/` | Secrets-matcher for query param scrubbing. |
| `process/` | Process metadata collection. |

### Key Interfaces

```go
// Primary user-facing interface
type TracerLogger interface {
    Tracer           // opentracing.Tracer + Options() + Flush() + StartSpanWithOptions()
    LeveledLogger    // Debug/Info/Warn/Error
    LegacySensor() *Sensor
    SensorLogger     // Tracer() ot.Tracer / Logger() / SetLogger()
}
```

### Span Lifecycle

1. User calls `col.StartSpan("operation", opts...)` on `Collector` (implements `TracerLogger`).
2. `tracerS.StartSpanWithOptions` creates a `spanS`, assigns a new `SpanContext` (or inherits from parent via `ot.ChildOf`).
3. User calls `span.Finish()` → `spanS.FinishWithOptions` → `sendSpanToAgent()` checks suppress flag and root-exit-span policy → `recorder.RecordSpan(span)`.
4. `Recorder` buffers spans; flushes to `AgentClient.SendSpans()` every second (or when `ForceTransmissionStartingAt` is reached).
5. If agent is not yet ready, spans go into a `delayed` queue and are replayed once the agent announces.

### Context Propagation Flow

- **HTTP inbound**: `TracingHandlerFunc` calls `tracer.Extract(ot.HTTPHeaders, ot.HTTPHeadersCarrier(req.Header))` → creates child span → stores span in `context.Context` via `instana.ContextWithSpan`.
- **HTTP outbound**: `RoundTripper` wraps `http.RoundTripper`; reads parent span from context via `instana.SpanFromContext`; calls `tracer.Inject` into outgoing headers.
- **Messaging**: e.g. `instasarama.ProducerMessageWithSpan()` injects trace context into Kafka record headers; consumer-side calls `tracer.Extract` from record headers.
- **gRPC**: `UnaryServerInterceptor` / `UnaryClientInterceptor` extract/inject via gRPC metadata using `ot.HTTPHeaders` format.
- Instana headers: `x-instana-t` (trace ID), `x-instana-s` (span ID), `x-instana-l` (level/suppression), `x-instana-b-*` (baggage).
- W3C `traceparent`/`tracestate` are also propagated automatically.

### Configuration Precedence

`ENV vars > in-code Options > agent config (configuration.yaml) > defaults`

Key env vars: `INSTANA_SERVICE_NAME`, `INSTANA_AGENT_HOST`, `INSTANA_AGENT_PORT`, `INSTANA_ENDPOINT_URL` (serverless), `INSTANA_AGENT_KEY`, `INSTANA_LOG_LEVEL`, `INSTANA_SECRETS`, `INSTANA_EXTRA_HTTP_HEADERS`, `INSTANA_TRACING_DISABLE`, `INSTANA_CONFIG_PATH`, `INSTANA_ALLOW_ROOT_EXIT_SPAN`, `INSTANA_DISABLE_W3C_TRACE_CORRELATION`, `INSTANA_AUTO_PROFILE`, `INSTANA_TIMEOUT`.

---

## 3. Coding Standards

- **Go style**: Standard `gofmt` + `goimports`. CI enforces both (`make fmtcheck`, `make importcheck`).
- **Linting**: `golangci-lint` with `deadcode`, `gosimple`, `govet`, `ineffassign`, `staticcheck`, `structcheck`, `typecheck`, `unused`, `varcheck`.
- **Copyright header**: Every `.go` file **must** start with `// (c) Copyright IBM Corp. <year>`. The `make legal` target enforces this.
- **Backward compatibility**: Never remove or change the signature of exported functions, types, or constants. Deprecate with `// Deprecated:` godoc instead.
- **Minimal dependencies**: Core module has only 4 direct dependencies (`opentracing-go`, `google/pprof`, `google/uuid`, `looplab/fsm`, `stretchr/testify`, `gopkg.in/yaml.v3`). Do not add new dependencies without strong justification.
- **No global mutable state** beyond the guarded `sensor` singleton and `once`-initialized `c` collector.
- **Concurrency**: All shared state must be protected. `sensorS` uses `sync.RWMutex`; `spanS` uses `sync.Mutex`. Do not introduce unguarded global variables.
- **Error handling**: Instrumentation must never panic or return errors that affect application logic. Log errors via `sensor.Logger()`, degrade gracefully.
- **Build tags**: Use `//go:build go1.XX` + `// +build go1.XX` pair for Go version-gated code.

---

## 4. Instrumentation Design Principles

- **Follow existing patterns**: Look at [`instrumentation/instagrpc`](instrumentation/instagrpc/), [`instrumentation/instagin`](instrumentation/instagin/), [`instrumentation/instasarama`](instrumentation/instasarama/) before writing new instrumentation.
- **Accept `instana.TracerLogger`**: All instrumentation entry points take `sensor instana.TracerLogger`, not `*Sensor` or `*Collector`.
- **Wrapper / middleware pattern**: Wrap the upstream type (struct embedding or middleware function) rather than patching or replacing it.
- **Preserve upstream behavior**: The instrumented version must behave identically to the original from the caller's perspective. Never change return values, error semantics, or side effects.
- **Never change application semantics**: Tracing is additive. Do not alter business logic, request routing, or data flow.
- **Span kinds**: Use `ext.SpanKindRPCServer` for entry spans, `ext.SpanKindRPCClient` for exit spans. Absence means intermediate.
- **Registered span types**: Use the constants in [`registered_span.go`](registered_span.go) (e.g. `instana.HTTPServerSpanType`, `instana.KafkaSpanType`). If a new span type is needed, define it in the core module first.
- **Context propagation**: Always store the active span with `instana.ContextWithSpan(ctx, span)` and pass the resulting context downstream. Retrieve with `instana.SpanFromContext(ctx)`.
- **Avoid duplicate spans**: Check `instana.SpanFromContext(ctx)` before starting a new span to avoid double-wrapping.
- **Suppress tracing**: Respect `instana.SuppressTracing()` tag and `SpanContext.Suppressed` flag.
- **Error recording**: Call `span.SetTag("error", true)` and `span.LogFields(otlog.Error(err))` on errors. Re-raise panics after recording.
- **Panic safety**: Use `defer recover()` in instrumentation interceptors to record panics, then re-panic.

---

## 5. Performance Rules

- **No unnecessary allocations** in hot paths (span creation, header injection/extraction).
- **No reflection**: Use type switches or direct type assertions instead.
- **No unnecessary goroutines**: Do not launch goroutines inside span creation or injection paths.
- **Minimal locking**: Use read locks for reads. Avoid holding locks across I/O.
- **Instrumentation overhead must be negligible**: Span creation and header operations should complete in microseconds.
- **Buffer reuse**: Prefer `bytes.Buffer` reuse over `fmt.Sprintf` for string building in hot paths.

---

## 6. Error Handling

- **Never panic in instrumentation code** (unless re-panicking after recording).
- **Never let instrumentation errors surface to the application**. Log with `sensor.Logger().Warn/Error()` and continue.
- **Preserve original errors**: Return upstream errors unchanged after recording them on the span.
- **Fail safely**: If the sensor/agent is not ready, skip span recording silently — do not block or error.
- **No noisy logging**: Use `Debug` for expected absent-context cases, `Warn` for unexpected-but-recoverable, `Error` for failures. Avoid `Info` spam.
- **Example pattern**:
  ```go
  switch sc, err := tracer.Extract(ot.HTTPHeaders, carrier); err {
  case nil:
      opts = append(opts, ot.ChildOf(sc))
  case ot.ErrSpanContextNotFound:
      sensor.Logger().Debug("no tracing context found, starting new trace")
  case ot.ErrUnsupportedFormat:
      sensor.Logger().Warn("unsupported format")
  default:
      sensor.Logger().Error("failed to extract context: ", err)
  }
  ```

---

## 7. Testing Requirements

### Test types

- **Unit tests**: Required for all new code. Use `instana.NewTestRecorder()` to capture spans without sending to an agent.
- **Race safety**: Run `go test -race ./...`. CI enforces this.
- **Integration tests**: Live-service tests use build tags (`-tags=integration`). See `Makefile` for tag names per package.
- **Benchmarks**: Add when a performance-sensitive path is changed or introduced.

### Test patterns

- Use `github.com/stretchr/testify` (`assert`, `require`) — it is the only test assertion library used.
- The `NewTestRecorder()` collects spans in memory; inspect via `recorder.GetQueuedSpans()`.
- Instrument test servers/clients with real `instana.TracerLogger` using `instana.InitCollector(...)` or `instana.NewSensorWithTracer(instana.NewTracerWithOptions(...))`.
- Tests must not rely on a live Instana agent.
- Each instrumentation package has its own `go.mod` and test suite. Run tests per-package with `cd instrumentation/<pkg> && go test ./...`.
- The core module tests live at the repo root: `go test ./...`.

### CI checks (must all pass)

1. `make fmtcheck` — gofmt compliance
2. `make importcheck` — goimports compliance
3. `make legal` — copyright header presence
4. `make test` (with `RUN_LINTER=yes`) — unit tests + golangci-lint
5. `make integration` — integration tests (CircleCI, requires service sidecars)

---

## 8. Pull Request Expectations

- **Branch**: PR against `main`.
- **Commit messages**: Follow [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) (`feat:`, `fix:`, `chore:`, `refactor:`, `ci:`, `docs:`).
- **Changelog**: Update [`CHANGELOG.md`](CHANGELOG.md) under an unreleased header with a concise bullet.
- **Versioning**: Semver. `minor` for new features or instrumentation packages; `patch` for fixes. Core and instrumentation packages are versioned independently.
- **CI must pass**: All `fmtcheck`, `importcheck`, `legal`, `test` targets. SonarCloud analysis runs automatically.
- **No speculative changes**: Only change what the task requires. Do not refactor unrelated code in the same PR.
- **New instrumentation package**: Update [`supported_versions.md`](supported_versions.md) with the new entry.

---

## 9. Documentation Expectations

- **README.md**: Update when a major new capability is added to the core.
- **`docs/`**: Update or add a doc file for new features (e.g. `docs/options.md`, `docs/manual_instrumentation.md`).
- **Package godoc**: Every exported type and function must have a godoc comment.
- **`supported_versions.md`**: Update for every new instrumentation package, including minimum and maximum tested upstream library versions.
- **Instrumentation `README.md`**: Each package under `instrumentation/` must have a `README.md` describing usage, with a working example using `InitCollector`.
- **`CHANGELOG.md`**: Every user-visible change requires a changelog entry before merging.
- **`example/`**: Add or update examples in `example/` for significant new public APIs.

---

## 10. Common Patterns

### Initialization

```go
col := instana.InitCollector(&instana.Options{
    Service: "my-service",
    Tracer:  instana.DefaultTracerOptions(),
})
```

`InitCollector` is a singleton — subsequent calls return the same `Collector`.

### HTTP server middleware

```go
http.HandleFunc("/path", instana.TracingHandlerFunc(col, "/path", handler))
// or with route ID:
instana.TracingNamedHandlerFunc(col, routeID, "/path/{id}", handler)
```

### Manual span

```go
sp := col.StartSpan("operation", ext.SpanKindRPCServer)
defer sp.Finish()
ctx = instana.ContextWithSpan(ctx, sp)
```

### Child span (exit)

```go
parent, ok := instana.SpanFromContext(ctx)
opts := []ot.StartSpanOption{ext.SpanKindRPCClient}
if ok {
    opts = append(opts, ot.ChildOf(parent.Context()))
}
sp := sensor.Tracer().StartSpan("kafka", opts...)
defer sp.Finish()
```

### Error recording

```go
if err != nil {
    sp.SetTag("error", true)
    sp.LogFields(otlog.Error(err))
}
```

### Wrapper type (messaging/DB)

```go
type SyncProducer struct {
    sarama.SyncProducer          // embed upstream type
    sensor instana.TracerLogger
}
// Override only the methods that need instrumentation.
```

### Middleware (HTTP frameworks)

```go
func middleware(sensor instana.TracerLogger) gin.HandlerFunc {
    return func(gc *gin.Context) {
        instana.TracingHandlerFunc(sensor, gc.FullPath(), func(w http.ResponseWriter, r *http.Request) {
            gc.Request = r
            gc.Next()
        })(gc.Writer, gc.Request)
    }
}
```

### gRPC interceptor pattern (server)

```go
func UnaryServerInterceptor(sensor instana.TracerLogger) grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        sp := startServerSpan(ctx, info.FullMethod, sensor)
        defer sp.Finish()
        defer func() {
            if err := recover(); err != nil { addRPCError(sp, err); panic(err) }
        }()
        m, err := handler(instana.ContextWithSpan(ctx, sp), req)
        if err != nil { addRPCError(sp, err) }
        return m, err
    }
}
```

### Context propagation in carrier (Kafka)

```go
// inject
sp.Tracer().Inject(sp.Context(), ot.TextMap, ProducerMessageCarrier{msg})
// extract
sc, err := sensor.Tracer().Extract(ot.TextMap, ProducerMessageCarrier{msg})
```

---

## 11. Agent Workflow

When implementing any task in this repository:

1. **Understand the existing implementation**: Read the relevant source files before writing anything. Never speculate about behavior.
2. **Search for similar instrumentation**: Look in `instrumentation/` for a package similar to what you're adding (e.g. another HTTP framework, another messaging library).
3. **Follow existing patterns exactly**: Match the structure of `handler.go`, `server.go`, `sync_producer.go`, etc. Do not invent new abstractions.
4. **Make the smallest correct change**: Do not refactor, rename, or restructure unrelated code.
5. **Registered span types first**: If a new span type is needed, add it to [`registered_span.go`](registered_span.go) in the core module and release it before adding the instrumentation package.
6. **Update tests**: Add unit tests using `instana.NewTestRecorder()`. Ensure `-race` passes.
7. **Format and lint**: Run `gofmt -w .` and `goimports -w .` before committing. Check copyright headers.
8. **Verify CI compatibility**: Run `make test` and `make legal` locally.
9. **Update documentation**: Add `README.md` to new packages, update `supported_versions.md`, update `CHANGELOG.md`.
10. **Prepare a clean PR**: One logical change per PR, Conventional Commits, changelog entry.

---

## 12. Repository-specific Knowledge

### Directory layout

```
go-sensor/
├── *.go                  # Core module (package instana)
├── go.mod                # Core module: github.com/instana/go-sensor
├── version.go            # Single source of truth for core version string
├── registered_span.go    # All registered span type constants
├── supported_versions.md # Compatibility matrix for instrumentation packages
├── instrumentation/
│   └── <instaXXX>/       # Each is a separate Go module
│       ├── go.mod        # module github.com/instana/go-sensor/instrumentation/<instaXXX>
│       ├── version.go    # Package-level version constant
│       ├── Makefile      # Delegates to Makefile.release
│       ├── README.md     # Required
│       └── *.go
├── docs/                 # User-facing documentation
├── example/              # Runnable usage examples
├── acceptor/             # HTTP transport to Instana backend
├── autoprofile/          # Continuous profiling
├── logger/               # Logger implementation
├── secrets/              # Secrets scrubbing
├── w3ctrace/             # W3C Trace Context
├── .github/workflows/    # GitHub Actions: release, govulncheck, support matrix
├── .circleci/config.yml  # Unit + integration tests
├── Makefile              # test, fmtcheck, importcheck, legal, integration
└── Makefile.release      # Versioning and release automation
```

### Module conventions

- Core: `github.com/instana/go-sensor`
- Instrumentation: `github.com/instana/go-sensor/instrumentation/<name>` (e.g. `github.com/instana/go-sensor/instrumentation/instagrpc`)
- Major version bumps (v2+): `instrumentation/<name>/v2` (directory and module path)
- Each instrumentation package pins the **latest released core** at the time of development. When core releases, an automated PR updates all packages.

### Registered span types (key subset)

| Constant | Operation name | Usage |
|---|---|---|
| `HTTPServerSpanType` | `g.http` | Incoming HTTP |
| `HTTPClientSpanType` | `http` | Outgoing HTTP |
| `RPCServerSpanType` | `rpc-server` | gRPC server |
| `RPCClientSpanType` | `rpc-client` | gRPC client |
| `KafkaSpanType` | `kafka` | Kafka producer/consumer |
| `LogSpanType` | `log.go` | Log entries |
| `SDKSpanType` | `sdk` | Custom/generic spans |
| Various DB types | `postgres`, `mysql`, `redis`, `mongo`, `couchbase`, etc. | DB clients |

### Release process

- Triggered via **GitHub Actions** → `Go Tracer Release` workflow (manual `workflow_dispatch`).
- Select package name (`.` for core, `instagin` for a package) and version type (`major`/`minor`/`patch`).
- After core release, an automated PR `update-instrumentations-core` bumps all instrumentation packages. Review and merge; a second action releases them all.
- Releases can be created as drafts for pre-review.

### Public API boundaries

- `instana.InitCollector(opts)` — primary entry point.
- `instana.TracerLogger` — interface passed to all instrumentation.
- `instana.TracingHandlerFunc` / `TracingNamedHandlerFunc` — HTTP middleware.
- `instana.ContextWithSpan` / `SpanFromContext` — context helpers.
- `instana.SpanContext`, `instana.SpanRecorder` — extensibility points.
- `instana.NewTestRecorder()` — test utility, part of public API.
- `instana.Options`, `instana.TracerOptions` — configuration types.

---
