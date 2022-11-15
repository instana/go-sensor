// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2016

package instana

import (
	"os"
	"strconv"
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
	// Tracer contains tracer-specific configuration used by all tracers
	Tracer TracerOptions
	// AgentClient client to communicate with the agent. In most cases, there is no need to provide it.
	// If it is nil the default implementation will be used.
	AgentClient AgentClient

	disableW3CTraceCorrelation bool
}

// DefaultOptions returns the default set of options to configure Instana sensor.
// The service name is set to the name of current executable, the MaxBufferedSpans
// and ForceTransmissionStartingAt are set to instana.DefaultMaxBufferedSpans and
// instana.DefaultForceSpanSendAt correspondigly. The AgentHost and AgentPort are
// taken from the env INSTANA_AGENT_HOST and INSTANA_AGENT_PORT if set, and default
// to localhost and 42699 otherwise.
func DefaultOptions() *Options {
	opts := &Options{
		Tracer: DefaultTracerOptions(),
	}
	opts.setDefaults()

	return opts
}

func (opts *Options) setDefaults() {
	if opts.MaxBufferedSpans == 0 {
		opts.MaxBufferedSpans = DefaultMaxBufferedSpans
	}

	if opts.ForceTransmissionStartingAt == 0 {
		opts.ForceTransmissionStartingAt = DefaultForceSpanSendAt
	}

	if opts.AgentHost == "" {
		opts.AgentHost = agentDefaultHost

		if host := os.Getenv("INSTANA_AGENT_HOST"); host != "" {
			opts.AgentHost = host
		}
	}

	if opts.AgentPort == 0 {
		opts.AgentPort = agentDefaultPort

		if port, err := strconv.Atoi(os.Getenv("INSTANA_AGENT_PORT")); err == nil {
			opts.AgentPort = port
		}
	}

	secretsMatcher, err := parseInstanaSecrets(os.Getenv("INSTANA_SECRETS"))
	if err != nil {
		defaultLogger.Warn("invalid INSTANA_SECRETS= env variable value: ", err, ", ignoring")
		secretsMatcher = opts.Tracer.Secrets
	}

	if secretsMatcher == nil {
		secretsMatcher = DefaultSecretsMatcher()
	}

	opts.Tracer.Secrets = secretsMatcher

	if collectableHeaders, ok := os.LookupEnv("INSTANA_EXTRA_HTTP_HEADERS"); ok {
		opts.Tracer.CollectableHTTPHeaders = parseInstanaExtraHTTPHeaders(collectableHeaders)
	}

	opts.disableW3CTraceCorrelation = os.Getenv("INSTANA_DISABLE_W3C_TRACE_CORRELATION") != ""
}
