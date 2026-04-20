# Configurable Metrics Transmission Interval Examples

This directory contains examples demonstrating how to configure the metrics transmission interval in the Instana Go Sensor.

## Overview

The Instana Go Sensor allows you to configure how frequently metrics are transmitted to the Instana agent. By default, metrics are sent every 1000ms (1 second), but this can be customized based on your application's needs.

## Configuration Methods

### 1. Environment Variable (env-config/)

Configure the interval using the `INSTANA_METRICS_TRANSMISSION_DELAY` environment variable.

```bash
export INSTANA_METRICS_TRANSMISSION_DELAY=2000
go run example/metrics-interval/env-config/main.go
```

**Advantages:**
- No code changes required
- Easy to adjust per environment (dev, staging, prod)
- Takes precedence over code configuration

### 2. Code Configuration (code-config/)

Configure the interval programmatically in your application code.

```go
instana.InitCollector(&instana.Options{
    Service: "my-service",
    Metrics: instana.MetricsOptions{
        TransmissionDelay: 3000, // 3 seconds
    },
})
```

**Advantages:**
- Explicit and visible in code
- Can be set conditionally based on application logic
- Type-safe configuration

## Configuration Rules

- **Valid Range**: 1000ms to 5000ms
- **Default**: 1000ms (if not specified or invalid)
- **Maximum Cap**: Values above 5000ms are automatically capped at 5000ms
- **Invalid Values**: Non-numeric, zero, or negative values fall back to default 1000ms
- **Precedence**: Environment variable > Code configuration > Default

## Running the Examples

### Environment Variable Example
environment variable is set programmatically in the code.
```bash
cd example/metrics-interval/env-config
go run main.go
```

### Code Configuration Example
```bash
cd example/metrics-interval/code-config
go run main.go
```

## Validation and Error Handling

The sensor validates all configuration values and provides warning logs for invalid inputs:

- **Invalid format**: Falls back to default 1000ms with warning
- **Out of range**: Caps at 5000ms or uses default for values ≤ 1000
- **Graceful degradation**: Application continues running with safe defaults

## See Also

- [Main Documentation](../../README.md)
- [Options Documentation](../../docs/options.md)
- [Instana Go Sensor API](https://pkg.go.dev/github.com/instana/go-sensor)