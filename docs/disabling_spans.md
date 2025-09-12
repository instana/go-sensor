# Disabling Spans by Category

The Instana Go Tracer allows you to disable specific categories of spans to reduce the amount of data being collected and processed. This can be useful in high-volume environments or when you want to focus on specific types of traces.

## Supported Span Categories

The following span categories can be disabled:

| Category    | Description            | Affected Instrumentations                                                                  |
| ----------- | ---------------------- | ------------------------------------------------------------------------------------------ |
| `http`      | HTTP requests          | `beego`, `fiber`, `fasthttp`, `gin`, `mux`, `httprouter`                                   |
| `rpc`       | Remote procedure calls | `grpc`                                                                                     |
| `graphql`   | GraphQL operations     | `graphql`                                                                                  |
| `messaging` | Messaging systems      | `kafka`, `rabbitmq`                                                                        |
| `databases` | Database operations    | `aws dynamodb`, `aws s3`, `mongodb`, `postgresql`, `mysql`, `redis`, `couchbase`, `cosmos` |
| `logging`   | Log spans              | `logrus`                                                                                   |

## Configuration Methods

There are three ways to disable span categories:

### 1. Using Code

You can disable specific span categories when initializing the tracer:

```go
col := instana.InitCollector(&instana.Options{
    Service: "My Service",
    Tracer: instana.TracerOptions{
        Disable: map[string]bool{
            "http": true,      // Disable HTTP spans
            "databases": true, // Disable database spans
        },
    },
})
```

To disable all categories at once:

```go
opts := &instana.Options{
    Service: "My Service",
}
opts.Tracer.DisableAllCategories()

col := instana.InitCollector(opts)
```

### 2. Using Environment Variables

You can disable span categories using the `INSTANA_TRACING_DISABLE` environment variable:

```bash
# Disable specific categories (comma-separated list)
export INSTANA_TRACING_DISABLE="http,databases,logging"

# Disable all tracing
export INSTANA_TRACING_DISABLE=true
```

### 3. Using Configuration File

You can create a YAML configuration file and specify its path using the `INSTANA_CONFIG_PATH` environment variable:

```yaml
# config.yaml
tracing:
  disable:
    - http
    - databases
```

```bash
export INSTANA_CONFIG_PATH=/path/to/config.yaml
```

## Priority Order

When multiple configuration methods are used, they are applied in the following order of precedence:

1. Configuration file (`INSTANA_CONFIG_PATH`)
2. Environment variable (`INSTANA_TRACING_DISABLE`)
3. Code-level configuration

## Examples

### Disable HTTP Spans

```go
col := instana.InitCollector(&instana.Options{
    Service: "My Service",
    Tracer: instana.TracerOptions{
        Disable: map[string]bool{
            "http": true,
        },
    },
})
```

### Disable GraphQL Spans

```go
col := instana.InitCollector(&instana.Options{
    Service: "My Service",
    Tracer: instana.TracerOptions{
        Disable: map[string]bool{
            "graphql": true,
        },
    },
})
```

### Disable Database Spans

```go
col := instana.InitCollector(&instana.Options{
    Service: "My Service",
    Tracer: instana.TracerOptions{
        Disable: map[string]bool{
            "databases": true,
        },
    },
})
```

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

- **Performance Optimization**: In high-throughput applications, disabling certain span categories can reduce the overhead of tracing.
- **Cost Management**: Reduce the volume of trace data sent to Instana to manage costs.
- **Focus on Specific Areas**: Disable less relevant categories to focus on the parts of your application that need attention.
- **Testing**: Temporarily disable certain categories during testing or development.

## Limitations

- Disabling a category affects all span types within that category.
- You cannot selectively enable specific span types within a disabled category.
- The `unknown` category cannot be disabled explicitly.