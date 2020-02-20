package autoprofile

const (
	defaultMaxBufferedProfiles = 100
)

var (
	samplerActive = &flag{}

	profileRecorder     = newRecorder()
	cpuSamplerScheduler = newSamplerScheduler(profileRecorder, newCPUSampler(), samplerConfig{
		logPrefix:          "CPU sampler:",
		maxProfileDuration: 20,
		maxSpanDuration:    2,
		maxSpanCount:       30,
		samplingInterval:   8,
		reportInterval:     120,
	})
	allocationSamplerScheduler = newSamplerScheduler(profileRecorder, newAllocationSampler(), samplerConfig{
		logPrefix:      "Allocation sampler:",
		reportOnly:     true,
		reportInterval: 120,
	})
	blockSamplerScheduler = newSamplerScheduler(profileRecorder, newBlockSampler(), samplerConfig{
		logPrefix:          "Block sampler:",
		maxProfileDuration: 20,
		maxSpanDuration:    4,
		maxSpanCount:       30,
		samplingInterval:   16,
		reportInterval:     120,
	})

	enabled bool
)

// Enable enables the auto profiling (disabled by default)
func Enable() {
	if enabled {
		return
	}

	profileRecorder.start()
	cpuSamplerScheduler.start()
	allocationSamplerScheduler.start()
	blockSamplerScheduler.start()

	log.debug("profiler enabled")
}

// Disable disables the auto profiling (default)
func Disable() {
	if !enabled {
		return
	}

	profileRecorder.stop()
	cpuSamplerScheduler.stop()
	allocationSamplerScheduler.stop()
	blockSamplerScheduler.stop()

	log.debug("profiler disabled")
}

// SetGetExternalPIDFunc configures the profiler to use provided function to retrieve the current PID
func SetGetExternalPIDFunc(fn func() string) {
	if fn == nil {
		fn = getLocalPID
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
