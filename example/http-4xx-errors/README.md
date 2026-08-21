# http-4xx-errors example

Demonstrates the **opt-in HTTP 4xx error classification** feature of the Instana Go tracer.

By default, the tracer does not treat HTTP 4xx responses as errors on exit spans.
This example opts in via a local `config.yaml` so that only **401** and **403** responses
are marked as errors.

## What you will see

| Request status | span.ec | Reason |
|---|---|---|
| 200 | 0 | Success |
| 401 | 1 | In `classify-as-errors` list |
| 403 | 1 | In `classify-as-errors` list |
| 404 | 0 | 4xx but **not** in the list |
| 500 | 1 | 5xx is always an error |

## Running the example

```bash
go run .
```

A local `httptest.Server` acts as the upstream — no real backend needed.
Check the log output to see which requests were made, then inspect your Instana
dashboard for the resulting exit spans.

## Configuration options

There are three ways to enable 4xx error classification (in decreasing priority):

### 1. Environment variables

```bash
# Mark specific codes as errors
INSTANA_TRACING_HTTP_EXIT_CLASSIFY_AS_ERRORS=401,403

# Or mark all 4xx as errors
INSTANA_TRACING_HTTP_EXIT_CLASSIFY_ALL_4XX_AS_ERRORS=true
```

### 2. Config file (`config.yaml`)

Point the tracer at a YAML file with `INSTANA_CONFIG_PATH`:

```yaml
com.instana.tracing:
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

In-code options are overridden by environment variables and the config file.

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

Agent configuration is applied at announce time and has the lowest precedence.
