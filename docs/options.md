## Options

#### Service

**Type:** ``string``

Service is the global service name that will be used to identify the program in the Instana backend.

#### AgentHost

**Type:** ``string``

AgentHost is the Instana host agent host name.

#### AgentPort

**Type:** ``int``

AgentPort is the Instana host agent port.

> [!NOTE]
> `AgentHost` and `AgentPort` options have no effect in serverless environments. To specify the serverless acceptor endpoint, define the ``INSTANA_ENDPOINT_URL`` env var.
> See [Serverless Monitoring](https://www.ibm.com/docs/en/instana-observability/current?topic=references-environment-variables#serverless-monitoring) for more details.

#### MaxBufferedSpans

**Type:** ``int``

MaxBufferedSpans is the maximum number of spans to buffer.

#### ForceTransmissionStartingAt

**Type:** ``int``

ForceTransmissionStartingAt is the number of spans to collect before flushing the buffer to the agent.

#### LogLevel

**Type:** ``int``

LogLevel is the initial logging level for the logger used by Instana tracer. 
The valid log levels are logger.{Error,Warn,Info,Debug}Level provided by the https://github.com/instana/go-sensor/logger package.

> [!NOTE]
> This setting is only used to initialize the default logger and has no effect if a custom logger is set via ``instana.SetLogger()``

#### EnableAutoProfile

**Type:** ``bool``

EnableAutoProfile enables automatic continuous process profiling when set to true.

#### MaxBufferedProfiles

**Type:** ``int``

MaxBufferedProfiles is the maximum number of profiles to buffer.

#### IncludeProfilerFrames

**Type:** ``bool``

IncludeProfilerFrames is whether to include profiler calls into the profile or not.

#### Tracer

**Type:** [TracerOptions](https://pkg.go.dev/github.com/instana/go-sensor#TracerOptions)

Tracer contains tracer-specific configuration used by all tracers

#### AgentClient

**Type:** [AgentClient](https://pkg.go.dev/github.com/instana/go-sensor#AgentClient)

AgentClient client to communicate with the agent. In most cases, there is no need to provide it.
If it is nil the default implementation will be used.

#### Recorder
**Type:** [SpanRecorder](https://pkg.go.dev/github.com/instana/go-sensor#SpanRecorder)

Recorder records and manages spans. When this option is not set, instana.NewRecorder() will be used.

#### MaxLogsPerSpan

**Type:** ``int``

`MaxLogsPerSpan` defines the maximum number of log records that can be attached to a span. If a span contains more logs than this limit, the excess logs will be dropped. Default value is `2`. If set to `0` default value will be used.

Go Tracer only captures log spans with severity `warn` or higher.

> [!NOTE]
> It is recommended to use the default value unless there's a specific need to retain more logs per span. 

-----
[README](../README.md) |
[Tracing HTTP Outgoing Requests](roundtripper.md) |
[Tracing SQL Driver Databases](sql.md) |
[Tracing Other Go Packages](other_packages.md) |
[Instrumenting Code Manually](manual_instrumentation.md)
