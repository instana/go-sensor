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

	if opts.Tracer.Secrets == nil {
		opts.Tracer = DefaultTracerOptions()
	}
}
