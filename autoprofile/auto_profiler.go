package autoprofile

const (
	defaultMaxBufferedProfiles = 100
)

var (
	samplerActive = &flag{}

	profileRecorder     = NewRecorder()
	cpuSamplerScheduler = NewSamplerScheduler(profileRecorder, NewCPUSampler(), SamplerConfig{
		LogPrefix:          "CPU sampler:",
		MaxProfileDuration: 20,
		MaxSpanDuration:    2,
		MaxSpanCount:       30,
		SamplingInterval:   8,
		ReportInterval:     120,
	})
	allocationSamplerScheduler = NewSamplerScheduler(profileRecorder, NewAllocationSampler(), SamplerConfig{
		LogPrefix:      "Allocation sampler:",
		ReportOnly:     true,
		ReportInterval: 120,
	})
	blockSamplerScheduler = NewSamplerScheduler(profileRecorder, NewBlockSampler(), SamplerConfig{
		LogPrefix:          "Block sampler:",
		MaxProfileDuration: 20,
		MaxSpanDuration:    4,
		MaxSpanCount:       30,
		SamplingInterval:   16,
		ReportInterval:     120,
	})

	enabled bool
)

// Enable enables the auto profiling (disabled by default)
func Enable() {
	if enabled {
		return
	}

	profileRecorder.Start()
	cpuSamplerScheduler.Start()
	allocationSamplerScheduler.Start()
	blockSamplerScheduler.Start()

	log.debug("profiler enabled")
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

	log.debug("profiler disabled")
}

// SetGetExternalPIDFunc configures the profiler to use provided function to retrieve the current PID
func SetGetExternalPIDFunc(fn func() string) {
	if fn == nil {
		fn = GetLocalPID
	}

	getPID = fn
}

// SetSendProfilesFunc configures the profiler to use provided function to write collected profiles
func SetSendProfilesFunc(fn SendProfilesFunc) {
	if fn == nil {
		fn = noopSendProfiles
	}

	profileRecorder.SendProfiles = fn
}

// Options contains profiler configuration
type Options struct {
	IncludeSensorFrames bool
	MaxBufferedProfiles int
}

// DefaultOptions returns profiler defaults
func DefaultOptions() Options {
	return Options{
		MaxBufferedProfiles: defaultMaxBufferedProfiles,
	}
}

// SetOptions configures the profiler with provided settings
func SetOptions(opts Options) {
	if opts.MaxBufferedProfiles < 1 {
		opts.MaxBufferedProfiles = defaultMaxBufferedProfiles
	}

	profileRecorder.MaxBufferedProfiles = opts.MaxBufferedProfiles
	includeSensorFrames = opts.IncludeSensorFrames
}
