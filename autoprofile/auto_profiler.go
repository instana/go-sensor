package autoprofile

import (
	"github.com/instana/go-sensor/autoprofile/internal"
	"github.com/instana/go-sensor/autoprofile/internal/logger"
	instalogger "github.com/instana/go-sensor/logger"
)

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

	profileRecorder.Stop()
	cpuSamplerScheduler.Stop()
	allocationSamplerScheduler.Stop()
	blockSamplerScheduler.Stop()

	logger.Debug("profiler disabled")
}

// SetGetExternalPIDFunc configures the profiler to use provided function to retrieve the current PID
func SetGetExternalPIDFunc(fn func() string) {
	if fn == nil {
		fn = internal.GetLocalPID
	}

	internal.GetPID = fn
}

// SetSendProfilesFunc configures the profiler to use provided function to write collected profiles
func SetSendProfilesFunc(fn internal.SendProfilesFunc) {
	if fn == nil {
		fn = internal.NoopSendProfiles
	}

	profileRecorder.SendProfiles = fn
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
