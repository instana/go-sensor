// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package autoprofile

import (
	"os"

	"github.com/instana/go-sensor/autoprofile/internal"
	"github.com/instana/go-sensor/autoprofile/internal/logger"
	instalogger "github.com/instana/go-sensor/logger"
)

// Profile represents profiler data sent to the host agent
//
// The type alias here is needed to expose the type defined inside the internal package.
// Ideally this type should've been defined in the same package with instana.agentS, however
// due to the way we activate profiling, this would introduce a circular dependency.
type Profile internal.AgentProfile

// SendProfilesFunc is a function that submits profiles to the host agent
type SendProfilesFunc func(profiles []Profile) error

var (
	profileRecorder     = internal.NewRecorder()
	cpuSamplerScheduler = internal.NewSamplerScheduler(profileRecorder, internal.NewCPUSampler(), internal.SamplerConfig{
		LogPrefix:          "CPU sampler:",
		MaxProfileDuration: 20,
		MaxSpanDuration:    2,
		MaxSpanCount:       30,
		SamplingInterval:   8,
		ReportInterval:     120,
	})
	allocationSamplerScheduler = internal.NewSamplerScheduler(profileRecorder, internal.NewAllocationSampler(), internal.SamplerConfig{
		LogPrefix:      "Allocation sampler:",
		ReportOnly:     true,
		ReportInterval: 120,
	})
	blockSamplerScheduler = internal.NewSamplerScheduler(profileRecorder, internal.NewBlockSampler(), internal.SamplerConfig{
		LogPrefix:          "Block sampler:",
		MaxProfileDuration: 20,
		MaxSpanDuration:    4,
		MaxSpanCount:       30,
		SamplingInterval:   16,
		ReportInterval:     120,
	})

	enabled bool
)

// SetLogLevel sets the min log level for autoprofiler
//
// Deprecated: use autoprofile.SetLogger() to set the logger and configure the min log level directly
func SetLogLevel(level int) {
	switch logger.Level(level) {
	case logger.ErrorLevel:
		logger.SetLogLevel(instalogger.ErrorLevel)
	case logger.WarnLevel:
		logger.SetLogLevel(instalogger.WarnLevel)
	case logger.InfoLevel:
		logger.SetLogLevel(instalogger.InfoLevel)
	default:
		logger.SetLogLevel(instalogger.DebugLevel)
	}
}

// SetLogger sets the leveled logger to use to output the diagnostic messages and errors
func SetLogger(l logger.LeveledLogger) {
	logger.SetLogger(l)
}

// Enable enables the auto profiling (disabled by default)
func Enable() {
	if enabled {
		return
	}

	profileRecorder.Start()
	cpuSamplerScheduler.Start()
	allocationSamplerScheduler.Start()
	blockSamplerScheduler.Start()

	logger.Debug("profiler enabled")
}

// Disable disables the auto profiling (default)
func Disable() {
	if !enabled {
		return
	}

	if _, ok := os.LookupEnv("INSTANA_AUTO_PROFILE"); ok {
		logger.Info("INSTANA_AUTO_PROFILE is set, ignoring the attempt to disable AutoProfile™")
		return
	}

	profileRecorder.Stop()
	cpuSamplerScheduler.Stop()
	allocationSamplerScheduler.Stop()
	blockSamplerScheduler.Stop()

	logger.Debug("profiler disabled")
}

// SetGetExternalPIDFunc configures the profiler to use provided function to retrieve the current PID
//
// Deprecated: this is a noop function, the PID is populated by the agent before sending
func SetGetExternalPIDFunc(fn func() string) {}

// SetSendProfilesFunc configures the profiler to use provided function to write collected profiles
func SetSendProfilesFunc(fn SendProfilesFunc) {
	if fn == nil {
		profileRecorder.SendProfiles = internal.NoopSendProfiles
		return
	}

	profileRecorder.SendProfiles = func(data []internal.AgentProfile) error {
		profiles := make([]Profile, 0, len(data))
		for _, p := range data {
			profiles = append(profiles, Profile(p))
		}

		return fn(profiles)
	}
}

// Options contains profiler configuration
type Options struct {
	IncludeProfilerFrames bool
	MaxBufferedProfiles   int
}

// DefaultOptions returns profiler defaults
func DefaultOptions() Options {
	return Options{
		MaxBufferedProfiles: internal.DefaultMaxBufferedProfiles,
	}
}

// SetOptions configures the profiler with provided settings
func SetOptions(opts Options) {
	if opts.MaxBufferedProfiles < 1 {
		opts.MaxBufferedProfiles = internal.DefaultMaxBufferedProfiles
	}

	profileRecorder.MaxBufferedProfiles = opts.MaxBufferedProfiles
	internal.IncludeProfilerFrames = opts.IncludeProfilerFrames
}
