// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2016

package instana

import (
	"os"
	"strconv"
	"strings"
)

// Options allows the user to configure the to-be-initialized sensor
type Options struct {
	// Service is the global service name that will be used to identify the program in the Instana backend
	Service string
	// AgentHost is the Instana host agent host name
	//
	// Note: This setting has no effect in serverless environments. To specify the serverless acceptor endpoint,
	// INSTANA_ENDPOINT_URL env var. See https://www.instana.com/docs/reference/environment_variables/#serverless-monitoring
	// for more details.
	AgentHost string
	// AgentPort is the Instana host agent port
	//
	// Note: This setting has no effect in serverless environments. To specify the serverless acceptor endpoint,
	// INSTANA_ENDPOINT_URL env var. See https://www.instana.com/docs/reference/environment_variables/#serverless-monitoring
	// for more details.
	AgentPort int
	// MaxBufferedSpans is the maximum number of spans to buffer
	MaxBufferedSpans int
	// ForceTransmissionStartingAt is the number of spans to collect before flushing the buffer to the agent
	ForceTransmissionStartingAt int
	// LogLevel is the initial logging level for the logger used by Instana tracer. The valid log levels are
	// logger.{Error,Warn,Info,Debug}Level provided by the github.com/instana/go-sensor/logger package.
	//
	// Note: this setting is only used to initialize the default logger and has no effect if a custom logger is set via instana.SetLogger()
	LogLevel int
	// EnableAutoProfile enables automatic continuous process profiling when set to true
	EnableAutoProfile bool
	// MaxBufferedProfiles is the maximum number of profiles to buffer
	MaxBufferedProfiles int
	// IncludeProfilerFrames is whether to include profiler calls into the profile or not
	IncludeProfilerFrames bool
	// Metrics contains metrics collection and transmission configuration.
	Metrics MetricsOptions
	// Tracer contains tracer-specific configuration used by all tracers
	Tracer TracerOptions
	// AgentClient client to communicate with the agent. In most cases, there is no need to provide it.
	// If it is nil the default implementation will be used.
	AgentClient AgentClient

	// Recorder records and manages spans. When this option is not set, instana.NewRecorder() will be used.
	Recorder SpanRecorder

	disableW3CTraceCorrelation bool
}

// DefaultOptions returns the default set of options to configure Instana sensor.
func DefaultOptions() *Options {
	opts := &Options{
		Recorder: NewRecorder(),
		Tracer:   DefaultTracerOptions(),
	}
	return opts
}

// applyConfiguration applies configuration values following the precedence order:
// environment variables > in-code configuration > agent config (configuration.yaml) > default value
//
// This is the central place where all configuration sources are resolved and merged.
func (opts *Options) applyConfiguration() {
	opts.applyBasicDefaults()
	opts.applyAgentConfiguration()
	opts.applyServiceConfiguration()
	opts.applyProfilingConfiguration()
	opts.applyTracerConfiguration()
}

// applyBasicDefaults sets default values for basic options if not already configured
func (opts *Options) applyBasicDefaults() {
	if opts.MaxBufferedSpans == 0 {
		opts.MaxBufferedSpans = DefaultMaxBufferedSpans
	}

	if opts.ForceTransmissionStartingAt == 0 {
		opts.ForceTransmissionStartingAt = DefaultForceSpanSendAt
	}
}

// applyAgentConfiguration resolves agent connection settings
// Precedence: ENV > in-code > default
func (opts *Options) applyAgentConfiguration() {
	// AgentHost
	if opts.AgentHost == "" {
		opts.AgentHost = agentDefaultHost
	}
	if host := os.Getenv("INSTANA_AGENT_HOST"); host != "" {
		opts.AgentHost = host
	}

	// AgentPort
	if opts.AgentPort == 0 {
		opts.AgentPort = agentDefaultPort
	}
	if port, err := strconv.Atoi(os.Getenv("INSTANA_AGENT_PORT")); err == nil {
		opts.AgentPort = port
	}
}

// applyServiceConfiguration resolves service identification settings
// Precedence: ENV > in-code > binaryName
func (opts *Options) applyServiceConfiguration() {
	if name, ok := os.LookupEnv("INSTANA_SERVICE_NAME"); ok {
		name := strings.TrimSpace(name)
		if name != "" {
			opts.Service = name
		}
	}
}

// applyProfilingConfiguration resolves profiling settings
// Precedence: ENV > in-code > default
func (opts *Options) applyProfilingConfiguration() {
	if _, ok := os.LookupEnv("INSTANA_AUTO_PROFILE"); ok {
		opts.EnableAutoProfile = true
	}
}

// applyTracerConfiguration resolves tracer-specific settings
// Precedence: ENV > in-code > agent config > default
func (opts *Options) applyTracerConfiguration() {
	opts.applySecretsConfiguration()
	opts.applyTracerDefaults()
	opts.applyHTTPHeadersConfiguration()
	opts.applyTracingDisableConfiguration()
	opts.applyW3CConfiguration()
}

// applySecretsConfiguration resolves secrets matcher configuration
// Precedence: ENV > in-code > agent config > default
func (opts *Options) applySecretsConfiguration() {
	secretsMatcher, err := parseInstanaSecrets(os.Getenv("INSTANA_SECRETS"))
	if err != nil {
		defaultLogger.Warn("invalid INSTANA_SECRETS= env variable value: ", err, ", ignoring")
		secretsMatcher = opts.Tracer.Secrets
	}

	if secretsMatcher == nil {
		// No ENV or in-code configuration - use default and mark for agent override
		secretsMatcher = DefaultSecretsMatcher()
		opts.Tracer.tracerDefaultSecrets = true
	}

	opts.Tracer.Secrets = secretsMatcher
}

// applyTracerDefaults sets default values for tracer options
func (opts *Options) applyTracerDefaults() {
	if opts.Tracer.MaxLogsPerSpan == 0 {
		opts.Tracer.MaxLogsPerSpan = MaxLogsPerSpan
	}
}

// applyHTTPHeadersConfiguration resolves HTTP headers collection settings
// Precedence: ENV > agent config
func (opts *Options) applyHTTPHeadersConfiguration() {
	if collectableHeaders, ok := os.LookupEnv("INSTANA_EXTRA_HTTP_HEADERS"); ok {
		opts.Tracer.CollectableHTTPHeaders = parseInstanaExtraHTTPHeaders(collectableHeaders)
	}
}

// applyTracingDisableConfiguration resolves tracing disable settings
// Precedence: INSTANA_CONFIG_PATH > INSTANA_TRACING_DISABLE > agent config
func (opts *Options) applyTracingDisableConfiguration() {
	if configPath, ok := lookupValidatedEnv("INSTANA_CONFIG_PATH"); ok {
		if err := parseConfigFile(configPath, &opts.Tracer); err != nil {
			defaultLogger.Warn("invalid INSTANA_CONFIG_PATH= env variable value: ", err, ", ignoring")
		}
	} else if tracingDisable, ok := lookupValidatedEnv("INSTANA_TRACING_DISABLE"); ok {
		parseInstanaTracingDisable(tracingDisable, &opts.Tracer)
	}
}

// applyW3CConfiguration resolves W3C trace correlation settings
// Precedence: ENV only
func (opts *Options) applyW3CConfiguration() {
	opts.disableW3CTraceCorrelation = os.Getenv("INSTANA_DISABLE_W3C_TRACE_CORRELATION") != ""
}
