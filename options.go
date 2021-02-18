// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2016

package instana

import (
	"os"
	"strconv"
)

// Options allows the user to configure the to-be-initialized sensor
type Options struct {
	Service                     string
	AgentHost                   string
	AgentPort                   int
	MaxBufferedSpans            int
	ForceTransmissionStartingAt int
	LogLevel                    int
	EnableAutoProfile           bool
	MaxBufferedProfiles         int
	IncludeProfilerFrames       bool
	Tracer                      TracerOptions

	disableW3CTraceCorrelation bool
}

// DefaultOptions returns the default set of options to configure Instana sensor.
// The service name is set to the name of current executable, the MaxBufferedSpans
// and ForceTransmissionStartingAt are set to instana.DefaultMaxBufferedSpans and
// instana.DefaultForceSpanSendAt correspondigly. The AgentHost and AgentPort are
// taken from the env INSTANA_AGENT_HOST and INSTANA_AGENT_PORT if set, and default
// to localhost and 46999 otherwise.
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
