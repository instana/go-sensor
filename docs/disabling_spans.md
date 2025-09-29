# Disabling Spans

The Instana Go Tracer allows you to disable spans to reduce the amount of data being collected and processed. This can be useful in high-volume environments or when you want to focus on specific types of traces.

## Supported Span Categories

Currently, only the following span category can be disabled:

| Category  | Description | Affected Instrumentations |
| --------- | ----------- | ------------------------- |
| `logging` | Log spans   | `logrus`                  |

## Configuration Methods

There are four ways to disable spans:

### 1. Using Code

You can disable spans when initializing the tracer:

```go
col := instana.InitCollector(&instana.Options{
    Service: "My Service",
    Tracer: instana.TracerOptions{
        DisableSpans: map[string]bool{
            "logging": true, // Disable log spans
        },
    },
})
```

### 2. Using Environment Variables

You can disable spans using the `INSTANA_TRACING_DISABLE` environment variable:

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

**Note:** The tracer enforces a maximum config file size of 1 MB. Any file exceeding this limit will be rejected during validation. This safeguard is in place to prevent accidental or malicious use of excessively large files, since configuration data is expected to remain lightweight.

### 4. Using Instana Agent Configuration

You can configure the Instana agent to disable spans for all applications monitored by this agent:

1. Locate your Instana agent configuration file.
2. Add the following configuration to the agent's configuration file:
```yaml
com.instana.tracing:
  disable:
    - logging
```
1. Restart the Instana agent:

## Priority Order

When multiple configuration methods are used, they are applied in the following decreasing order of precedence:

1. Configuration file (`INSTANA_CONFIG_PATH`)
2. Environment variable (`INSTANA_TRACING_DISABLE`)
3. Code-level configuration
4. Agent configuration 

## Use Cases

- Users can control the log ingestion with the newly added configuration - fair usage policy data optimization.