# Disabling Server-Timing Header Injection

For HTTP entry spans, the Instana Go Tracer adds an Instana-specific metric to the standard
[`Server-Timing`](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Server-Timing) response header:

```
Server-Timing: intid;desc=<trace-id>
```

The Instana browser agent (EUM) reads this value to correlate front-end page loads and XHR calls with the
back-end trace. If the response already contains `Server-Timing` values, the `intid` metric is appended to them.

Because the header is visible to anyone who can see the response, you might not want to expose it on public-facing APIs.
The tracer therefore lets you disable the injection.

## What happens when it is disabled

- The tracer does not add the `intid` metric to the `Server-Timing` header.
- The tracer does not add a new `Server-Timing` header.
- `Server-Timing` values that were not added by Instana are never removed or modified.
- Back-end trace propagation is not affected. The `X-INSTANA-T`, `X-INSTANA-S` and `X-INSTANA-L` headers and the
  W3C `traceparent`/`tracestate` headers are still propagated.

> **Warning:** Disabling the `Server-Timing` injection also disables **EUM-to-backend correlation** for the affected
> requests. The browser agent can no longer link front-end page loads and XHR calls to the back-end traces, so
> distributed traces spanning EUM and the Go service will be broken.

The default behavior is unchanged: the `Server-Timing` value is injected unless you disable it.

## Configuration methods

The setting is a boolean and defaults to `false`. When the same setting is configured in several places,
the first one in the following list wins.

### 1. `INSTANA_CONFIG_PATH` configuration file

Point `INSTANA_CONFIG_PATH` to a YAML file containing:

```yaml
tracing:
  global:
    disable-server-timing: true
```

```bash
export INSTANA_CONFIG_PATH=/path/to/config.yaml
```

### 2. Environment variable

```bash
export INSTANA_TRACING_DISABLE_SERVER_TIMING=true
```

Accepted values are `true`/`1` to disable the injection and `false`/`0` to keep it enabled (case-insensitive).
Any other value is ignored with a warning.

### 3. In code

```go
col := instana.InitCollector(&instana.Options{
    Service: "My Service",
    Tracer: instana.TracerOptions{
        DisableServerTiming: true,
    },
})
```

### 4. Instana host agent configuration

In the host agent `configuration.yaml`:

```yaml
com.instana.tracing:
  global:
    disable-server-timing: true
```

The host agent setting is only used if the option has not been set with any of the methods above.

## Configuration reloadability

The Go tracer reads the host agent configuration when it announces itself to the agent, and applies it again
whenever the connection to the agent is re-established. Changes to the host agent `configuration.yaml`
are therefore picked up on the next announce, not instantly. Environment variables, `INSTANA_CONFIG_PATH`
and in-code configuration are read once at initialization and require an application restart to change.
