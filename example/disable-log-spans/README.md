# Disabling Log Spans Example

This example demonstrates how to disable log spans in an Instana-instrumented Go application using four different methods:
1. Using Code
2. Using Environment Variables
3. Using Configuration File
4. Using Instana Agent Configuration

The application uses logrus for logging and GORM with SQLite for database operations. It creates a simple HTTP server with two endpoints:
- `/start` - Creates a database table and inserts a record
- `/error` - Deliberately causes a database error

Both endpoints generate log entries that would normally create log spans in Instana.

## Prerequisites

- Go 1.23 or later
- The Instana Agent should either be running locally or accessible via the network. If no agent is available, a serverless endpoint must be configured.

## Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/instana/go-sensor.git
   cd go-sensor/example/disable-log-spans
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

## Running the Example

### Scenario 1: Disable Log Spans Using Code

In this scenario, we'll disable log spans directly in the code by modifying the Instana tracer options.

1. Edit `main.go` and update the `init()` function to include the `Disable` option:

```go
func init() {
    c = instana.InitCollector(&instana.Options{
        Service: "logrus-example",
        Tracer: instana.TracerOptions{
            DisableSpans: map[string]bool{
                "logging": true,
            },
        },
    })
    connectDB()
}
```

2. Run the application:
```bash
go run main.go
```

3. Expected output:
   - The application will start and make requests to both endpoints
   - You'll see log messages in the console
   - In the Instana UI, you'll see HTTP and database spans, but no log spans

### Scenario 2: Disable Log Spans Using Environment Variables

In this scenario, we'll disable log spans using the `INSTANA_TRACING_DISABLE` environment variable.

1. Revert any changes made to `main.go` for Scenario 1.

2. Run the application with the environment variable:
```bash
INSTANA_TRACING_DISABLE=logging go run main.go
```

3. Expected output:
   - The application will start and make requests to both endpoints
   - You'll see log messages in the console
   - In the Instana UI, you'll see HTTP and database spans, but no log spans

### Scenario 3: Disable Log Spans Using Configuration File

In this scenario, we'll disable log spans using a YAML configuration file.

1. Ensure the `config.yaml` file contains:
```yaml
tracing:
  disable:
    - logging: true
```

2. Run the application with the configuration file path:
```bash
INSTANA_CONFIG_PATH=./config.yaml go run main.go
```

3. Expected output:
   - The application will start and make requests to both endpoints
   - You'll see log messages in the console
   - In the Instana UI, you'll see HTTP and database spans, but no log spans

### Scenario 4: Disable Log Spans Using Instana Agent Configuration

In this scenario, we'll configure the Instana agent to disable log spans for all applications monitored by this agent.

1. Locate your Instana agent configuration file:

2. Add the following configuration to the agent's configuration file:
```yaml
com.instana.tracing:
  disable:
    - logging: true
```
3. Restart the Instana agent.

4. Run the application without any special configuration:
```bash
go run main.go
```

1. Expected output:
   - The application will start and make requests to both endpoints
   - You'll see log messages in the console
   - In the Instana UI, you'll see HTTP and database spans, but no log spans
   - All applications monitored by this agent will have log spans disabled

## Verifying Results

To verify that log spans are being disabled:

1. Run the application without disabling log spans:
```bash
go run main.go
```

2. Check the Instana UI - you should see log spans along with HTTP and database spans.

3. Run the application using one of the four methods to disable log spans.

4. Recheck the Instana UI - you should see HTTP and database spans, but no log spans.

## Priority Order

When multiple configuration methods are used, they are applied in the following order of precedence:

1. Code-level configuration
2. Configuration file (`INSTANA_CONFIG_PATH`)
3. Environment variable (`INSTANA_TRACING_DISABLE`)
4. Agent configuration 
