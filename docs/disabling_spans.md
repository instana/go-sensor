# Disabling Log Spans

The Instana Go Tracer allows you to disable log spans to reduce the amount of data being collected and processed. This can be useful in high-volume environments or when you want to focus on specific types of traces.

## Supported Span Categories

Currently, only the following span category can be disabled:

| Category  | Description | Affected Instrumentations |
| --------- | ----------- | ------------------------- |
| `logging` | Log spans   | `logrus`                  |

## Configuration Methods

There are three ways to disable log spans:

### 1. Using Code

You can disable log spans when initializing the tracer:

```go
col := instana.InitCollector(&instana.Options{
    Service: "My Service",
    Tracer: instana.TracerOptions{
        Disable: map[string]bool{
            "logging": true, // Disable log spans
        },
    },
})
```

### 2. Using Environment Variables

You can disable log spans using the `INSTANA_TRACING_DISABLE` environment variable:

```bash
# Disable log spans
export INSTANA_TRACING_DISABLE="logging"

# Disable all tracing
export INSTANA_TRACING_DISABLE=true
```

### 3. Using Configuration File

You can create a YAML configuration file and specify its path using the `INSTANA_CONFIG_PATH` environment variable:

```yaml
# config.yaml
tracing:
  disable:
    - logging
```

```bash
export INSTANA_CONFIG_PATH=/path/to/config.yaml
```

## Priority Order

When multiple configuration methods are used, they are applied in the following order of precedence:

1. Configuration file (`INSTANA_CONFIG_PATH`)
2. Environment variable (`INSTANA_TRACING_DISABLE`)
3. Code-level configuration

## Example

### Disable Log Spans

```go
col := instana.InitCollector(&instana.Options{
    Service: "My Service",
    Tracer: instana.TracerOptions{
        Disable: map[string]bool{
            "logging": true,
        },
    },
})
```

## Use Cases

- **Performance Optimization**: In high-throughput applications, disabling log spans can reduce the overhead of tracing.
- **Cost Management**: Reduce the volume of trace data sent to Instana to manage costs.
- **Focus on Specific Areas**: Disable log spans to focus on the parts of your application that need attention.
- **Testing**: Temporarily disable log spans during testing or development.