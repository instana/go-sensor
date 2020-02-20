package autoprofile

import (
	"path/filepath"
	"sync/atomic"
)

const (
	defaultMaxBufferedProfiles = 100
)

var (
	sensorPath = filepath.Join("github.com", "instana", "go-sensor")
	profiler   = newAutoProfiler()
)

// Enable enables the auto profiling (disabled by default)
func Enable() {
	profiler.Enable()
}

// Disable disables the auto profiling (default)
func Disable() {
	profiler.Disable()
}

// SetGetExternalPIDFunc configures the profiler to use provided function to retrieve the current PID
func SetGetExternalPIDFunc(fn func() string) {
	profiler.GetExternalPID = fn
}

// SetSendProfilesFunc configures the profiler to use provided function to write collected profiles
func SetSendProfilesFunc(fn func(interface{}) error) {
	profiler.SendProfiles = fn
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

	profiler.MaxBufferedProfiles = opts.MaxBufferedProfiles
	profiler.IncludeSensorFrames = opts.IncludeSensorFrames
}

type autoProfiler struct {
	profileRecorder            *ProfileRecorder
	cpuSamplerScheduler        *SamplerScheduler
	allocationSamplerScheduler *SamplerScheduler
	blockSamplerScheduler      *SamplerScheduler

	enabled       bool
	samplerActive *Flag

	// Options
	IncludeSensorFrames bool
	MaxBufferedProfiles int

	SendProfiles   func(profiles interface{}) error
	GetExternalPID func() string
}

func newAutoProfiler() *autoProfiler {
	ap := &autoProfiler{
		samplerActive:       &Flag{},
		MaxBufferedProfiles: 100,
	}

	ap.profileRecorder = newProfileRecorder(ap)

	cpuSampler := newCPUSampler(ap)
	cpuSamplerConfig := &SamplerConfig{
		logPrefix:          "CPU sampler:",
		maxProfileDuration: 20,
		maxSpanDuration:    2,
		maxSpanCount:       30,
		samplingInterval:   8,
		reportInterval:     120,
	}
	ap.cpuSamplerScheduler = newSamplerScheduler(ap, cpuSampler, cpuSamplerConfig)

	allocationSampler := newAllocationSampler(ap)
	allocationSamplerConfig := &SamplerConfig{
		logPrefix:      "Allocation sampler:",
		reportOnly:     true,
		reportInterval: 120,
	}
	ap.allocationSamplerScheduler = newSamplerScheduler(ap, allocationSampler, allocationSamplerConfig)

	blockSampler := newBlockSampler(ap)
	blockSamplerConfig := &SamplerConfig{
		logPrefix:          "Block sampler:",
		maxProfileDuration: 20,
		maxSpanDuration:    4,
		maxSpanCount:       30,
		samplingInterval:   16,
		reportInterval:     120,
	}
	ap.blockSamplerScheduler = newSamplerScheduler(ap, blockSampler, blockSamplerConfig)

	return ap
}

func (ap *autoProfiler) Enable() {
	if ap.enabled {
		return
	}

	ap.profileRecorder.start()
	ap.cpuSamplerScheduler.start()
	ap.allocationSamplerScheduler.start()
	ap.blockSamplerScheduler.start()

	log.debug("profiler enabled")
}

func (ap *autoProfiler) Disable() {
	if !ap.enabled {
		return
	}

	ap.profileRecorder.stop()
	ap.cpuSamplerScheduler.stop()
	ap.allocationSamplerScheduler.stop()
	ap.blockSamplerScheduler.stop()

	log.debug("profiler disabled")
}

type Flag struct {
	value int32
}

func (f *Flag) SetIfUnset() bool {
	return atomic.CompareAndSwapInt32(&f.value, 0, 1)
}

func (f *Flag) UnsetIfSet() bool {
	return atomic.CompareAndSwapInt32(&f.value, 1, 0)
}

func (f *Flag) Set() {
	atomic.StoreInt32(&f.value, 1)
}

func (f *Flag) Unset() {
	atomic.StoreInt32(&f.value, 0)
}

func (f *Flag) IsSet() bool {
	return atomic.LoadInt32(&f.value) == 1
}
