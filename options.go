package instana

import (
	"os"
	"path/filepath"
)

// Options allows the user to configure the to-be-initialized
// sensor
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
}

// DefaultOptions returns the default set of options to configure Instana sensor.
// The service name is set to the name of current executable, the MaxBufferedSpans
// and ForceTransmissionStartingAt are set to instana.DefaultMaxBufferedSpans and
// instana.DefaultForceSpanSendAt correspondigly
func DefaultOptions() *Options {
	opts := &Options{}
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

	if opts.Service == "" {
		opts.Service = filepath.Base(os.Args[0])
	}
}
