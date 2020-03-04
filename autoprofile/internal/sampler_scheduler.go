package internal

import (
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/instana/go-sensor/autoprofile/internal/logger"
)

var (
	GetPID        = GetLocalPID
	samplerActive Flag
)

func GetLocalPID() string {
	logger.Warn("using the local process pid as a default")
	return strconv.Itoa(os.Getpid())
}

type SamplerConfig struct {
	LogPrefix          string
	ReportOnly         bool
	MaxProfileDuration int64
	MaxSpanDuration    int64
	MaxSpanCount       int32
	SamplingInterval   int64
	ReportInterval     int64
}

type Sampler interface {
	Profile(duration int64, timespan int64) (*Profile, error)
	Start() error
	Stop() error
	Reset()
}

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

func NewSamplerScheduler(profileRecorder *Recorder, samp Sampler, config SamplerConfig) *SamplerScheduler {
	pr := &SamplerScheduler{
		profileRecorder: profileRecorder,
		sampler:         samp,
		config:          config,
		profileLock:     &sync.Mutex{},
	}

	return pr
}

func (ss *SamplerScheduler) Start() {
	if !ss.started.SetIfUnset() {
		return
	}

	ss.profileLock.Lock()
	defer ss.profileLock.Unlock()

	ss.Reset()

	if !ss.config.ReportOnly {
		ss.samplerTimer = NewTimer(0, time.Duration(ss.config.SamplingInterval)*time.Second, func() {
			time.Sleep(time.Duration(rand.Int63n(ss.config.SamplingInterval-ss.config.MaxSpanDuration)) * time.Second)
			ss.startProfiling()
		})
	}

	ss.reportTimer = NewTimer(0, time.Duration(ss.config.ReportInterval)*time.Second, func() {
		ss.Report()
	})
}

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

func (ss *SamplerScheduler) Reset() {
	ss.sampler.Reset()
	ss.profileStart = time.Now().Unix()
	ss.samplingDuration = 0
}

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
	} else {
		if len(profile.Roots) == 0 {
			logger.Debug(ss.config.LogPrefix, "not recording empty profile")
			ss.Reset()
			return
		}

		externalPID := GetPID()
		if externalPID != "" {
			profile.ProcessID = externalPID
			logger.Debug("using external PID", externalPID)
		} else {
			logger.Info("external PID from agent is not available, using own PID")
		}

		ss.profileRecorder.Record(profile.ToMap())
		logger.Debug(ss.config.LogPrefix, "recorded profile")
	}

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
