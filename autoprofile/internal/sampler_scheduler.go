// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package internal

import (
	"crypto/rand"
	"math/big"
	"sync"
	"time"

	"github.com/instana/go-sensor/autoprofile/internal/logger"
)

var samplerActive Flag

// SamplerConfig holds profile sampler setting
type SamplerConfig struct {
	LogPrefix          string
	ReportOnly         bool
	MaxProfileDuration int64
	MaxSpanDuration    int64
	MaxSpanCount       int32
	SamplingInterval   int64
	ReportInterval     int64
}

// Sampler gathers continuous profile samples over a period of time
type Sampler interface {
	Profile(duration int64, timespan int64) (*Profile, error)
	Start() error
	Stop() error
	Reset()
}

// SamplerScheduler periodically runs the sampler for a time period
type SamplerScheduler struct {
	profileRecorder  *Recorder
	sampler          Sampler
	config           SamplerConfig
	started          Flag
	samplerTimer     *Timer
	reportTimer      *Timer
	profileLock      *sync.Mutex
	profileStart     int64
	samplingDuration int64
	samplerStart     int64
	samplerTimeout   *Timer
}

// NewSamplerScheduler initializes a new SamplerScheduler for a sampler
func NewSamplerScheduler(profileRecorder *Recorder, samp Sampler, config SamplerConfig) *SamplerScheduler {
	pr := &SamplerScheduler{
		profileRecorder: profileRecorder,
		sampler:         samp,
		config:          config,
		profileLock:     &sync.Mutex{},
	}

	return pr
}

// Start runs the SampleScheduler
func (ss *SamplerScheduler) Start() {
	if !ss.started.SetIfUnset() {
		return
	}

	ss.profileLock.Lock()
	defer ss.profileLock.Unlock()

	ss.Reset()

	if !ss.config.ReportOnly {
		ss.samplerTimer = NewTimer(0, time.Duration(ss.config.SamplingInterval)*time.Second, func() {
			upperLimit := big.NewInt(ss.config.SamplingInterval - ss.config.MaxSpanDuration)
			upperLimit.Abs(upperLimit)
			var val *big.Int
			var err error
			if val, err = rand.Int(rand.Reader, upperLimit); err != nil {
				logger.Error(ss.config.LogPrefix, "error generating random number: ", err)
				return
			}
			dur := val.Int64()

			time.Sleep(time.Duration(dur) * time.Second)
			ss.startProfiling()
		})
	}

	ss.reportTimer = NewTimer(0, time.Duration(ss.config.ReportInterval)*time.Second, func() {
		ss.Report()
	})
}

// Stop prevents the SamplerScheduler from running the sampler
func (ss *SamplerScheduler) Stop() {
	if !ss.started.UnsetIfSet() {
		return
	}

	if ss.samplerTimer != nil {
		ss.samplerTimer.Stop()
	}

	if ss.reportTimer != nil {
		ss.reportTimer.Stop()
	}
}

// Reset resets the sampler and clears the internal state of the scheduler
func (ss *SamplerScheduler) Reset() {
	ss.sampler.Reset()
	ss.profileStart = time.Now().Unix()
	ss.samplingDuration = 0
}

// Report retrieves the collected profile from the sampler and enqueues it for submission
func (ss *SamplerScheduler) Report() {
	if !ss.started.IsSet() {
		return
	}

	profileTimespan := time.Now().Unix() - ss.profileStart

	ss.profileLock.Lock()
	defer ss.profileLock.Unlock()

	if !ss.config.ReportOnly && ss.samplingDuration == 0 {
		return
	}

	logger.Debug(ss.config.LogPrefix, "recording profile")

	profile, err := ss.sampler.Profile(ss.samplingDuration, profileTimespan)
	if err != nil {
		logger.Error(err)
		return
	}

	if len(profile.Roots) == 0 {
		logger.Debug(ss.config.LogPrefix, "not recording empty profile")
		ss.Reset()
		return
	}

	ss.profileRecorder.Record(NewAgentProfile(profile))
	logger.Debug(ss.config.LogPrefix, "recorded profile")

	ss.Reset()
}

func (ss *SamplerScheduler) startProfiling() bool {
	if !ss.started.IsSet() {
		return false
	}

	ss.profileLock.Lock()
	defer ss.profileLock.Unlock()

	if ss.samplingDuration > ss.config.MaxProfileDuration*1e9 {
		logger.Debug(ss.config.LogPrefix, "max sampling duration reached")
		return false
	}

	if !samplerActive.SetIfUnset() {
		return false
	}

	logger.Debug(ss.config.LogPrefix, "starting")

	err := ss.sampler.Start()
	if err != nil {
		samplerActive.Unset()
		logger.Error(err)
		return false
	}
	ss.samplerStart = time.Now().UnixNano()
	ss.samplerTimeout = NewTimer(time.Duration(ss.config.MaxSpanDuration)*time.Second, 0, func() {
		ss.stopSampler()
		samplerActive.Unset()
	})

	return true
}

func (ss *SamplerScheduler) stopSampler() {
	ss.profileLock.Lock()
	defer ss.profileLock.Unlock()

	if ss.samplerTimeout != nil {
		ss.samplerTimeout.Stop()
	}

	err := ss.sampler.Stop()
	if err != nil {
		logger.Error(err)
		return
	}
	logger.Debug(ss.config.LogPrefix, "stopped")

	ss.samplingDuration += time.Now().UnixNano() - ss.samplerStart
}
