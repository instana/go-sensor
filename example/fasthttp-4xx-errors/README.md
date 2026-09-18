# fasthttp-4xx-errors example

Demonstrates the **opt-in HTTP 4xx error classification** feature of the Instana Go tracer
using the [`instafasthttp`](../../instrumentation/instafasthttp) instrumentation package.

By default, the tracer does not treat HTTP 4xx responses as errors on exit spans.
This example opts in via a local `config.yaml` so that only **401** and **403** responses
are marked as errors. It exercises both the `RoundTripper` and `GetInstrumentedClient` paths.

## What you will see

| Request status | `span.ec` | Reason |
|---|---|---|
| 200 | 0 | Success |
| 401 | 1 | In `classify-as-errors` list |
| 403 | 1 | In `classify-as-errors` list |
| 404 | 0 | 4xx but **not** in the list |
| 500 | 1 | 5xx is always an error |

## Running the example

```bash
cd example/fasthttp-4xx-errors
go run main.go
```

The server starts on port `7071`. In another terminal, trigger the instrumented requests:

```bash
curl http://localhost:7071/test
```

This invokes the instrumented `/test` entry endpoint which performs upstream requests via both `RoundTripper` and `GetInstrumentedClient`. Since the client calls carry the entry span's context, the exit spans attach cleanly under the `/test` parent span in the Instana trace view.

## Configuration options

### 1. Environment variables (highest priority)

```bash
# Mark specific codes as errors
INSTANA_TRACING_HTTP_EXIT_CLASSIFY_AS_ERRORS=401,403 go run .

# Or mark all 4xx as errors
INSTANA_TRACING_HTTP_EXIT_CLASSIFY_ALL_4XX_AS_ERRORS=true go run .
```

### 2. Config file (`config.yaml`)

Point the tracer at a YAML file with `INSTANA_CONFIG_PATH`.
Note the root key is `tracing:` for config files — **not** `com.instana.tracing:` (that is only for the agent's `configuration.yaml`).

```yaml
tracing:
  http:
    exit:
      classify-as-errors:
        - 401
        - 403
      # or: classify-all-4xx-as-errors: true
```

This example sets `INSTANA_CONFIG_PATH=config.yaml` in `main.go`.

### 3. In-code options

```go
instana.Options{
    Tracer: instana.TracerOptions{
        HTTP: struct{ Exit instana.HTTPExitSettings }{
            Exit: instana.HTTPExitSettings{
                ClassifyAsErrors: []int{401, 403},
            },
        },
    },
}
```

### 4. Instana agent configuration (lowest priority)

In the agent's `configuration.yaml`:

```yaml
com.instana.tracing:
  http:
    exit:
      classify-as-errors:
        - 401
        - 403
```

## Entry spans are never affected

The `TraceHandler` server-side (entry) span always follows the spec:
- 5xx → `ec=1`  
- 4xx → `ec=0` regardless of any configuration

Only outbound exit spans (`RoundTripper` / `GetInstrumentedClient`) are affected by the opt-in configuration.
