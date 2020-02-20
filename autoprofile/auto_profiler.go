package autoprofile

import (
	"crypto/sha1"
	"encoding/hex"
	"math/rand"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var sensorPath = filepath.Join("github.com", "instana", "go-sensor")
var nextID int64

var randSource *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
var randLock *sync.Mutex = &sync.Mutex{}

var profiler *AutoProfiler = nil

type AutoProfiler struct {
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

func GetProfiler() *AutoProfiler {
	if profiler == nil {
		profiler = newAutoProfiler()
	}
	return profiler
}

func newAutoProfiler() *AutoProfiler {
	ap := &AutoProfiler{
		profileRecorder:            nil,
		cpuSamplerScheduler:        nil,
		allocationSamplerScheduler: nil,
		blockSamplerScheduler:      nil,

		enabled:       false,
		samplerActive: &Flag{},

		IncludeSensorFrames: false,
		MaxBufferedProfiles: 100,

		SendProfiles:   nil,
		GetExternalPID: nil,
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

func (ap *AutoProfiler) SetLogLevel(level int) {
	log.logLevel = level
}

func (ap *AutoProfiler) Enable() {
	if !ap.enabled {
		ap.profileRecorder.start()
		ap.cpuSamplerScheduler.start()
		ap.allocationSamplerScheduler.start()
		ap.blockSamplerScheduler.start()

		log.debug("profiler enabled")
	}
}

func (ap *AutoProfiler) Disable() {
	if ap.enabled {
		ap.profileRecorder.stop()
		ap.cpuSamplerScheduler.stop()
		ap.allocationSamplerScheduler.stop()
		ap.blockSamplerScheduler.stop()

		log.debug("profiler disabled")
	}
}

func recoverAndLog() {
	if err := recover(); err != nil {
		log.error("recovered from panic in agent:", err)
	}
}

func generateUUID() string {
	n := atomic.AddInt64(&nextID, 1)

	uuid :=
		strconv.FormatInt(time.Now().Unix(), 10) +
			strconv.FormatInt(random(1000000000), 10) +
			strconv.FormatInt(n, 10)

	return sha1String(uuid)
}

func random(max int64) int64 {
	randLock.Lock()
	defer randLock.Unlock()

	return randSource.Int63n(max)
}

func sha1String(s string) string {
	sha1 := sha1.New()
	sha1.Write([]byte(s))

	return hex.EncodeToString(sha1.Sum(nil))
}

type Timer struct {
	delayTimer         *time.Timer
	delayTimerDone     chan bool
	intervalTicker     *time.Ticker
	intervalTickerDone chan bool
	stopped            bool
}

func NewTimer(delay time.Duration, interval time.Duration, job func()) *Timer {
	t := &Timer{
		stopped: false,
	}

	t.delayTimerDone = make(chan bool)
	t.delayTimer = time.NewTimer(delay)
	go func() {
		defer recoverAndLog()

		select {
		case <-t.delayTimer.C:
			if interval > 0 {
				t.intervalTickerDone = make(chan bool)
				t.intervalTicker = time.NewTicker(interval)
				go func() {
					defer recoverAndLog()

					for {
						select {
						case <-t.intervalTicker.C:
							job()
						case <-t.intervalTickerDone:
							return
						}
					}
				}()
			}

			if delay > 0 {
				job()
			}
		case <-t.delayTimerDone:
			return
		}
	}()

	return t
}

func (t *Timer) Stop() {
	if !t.stopped {
		t.stopped = true

		t.delayTimer.Stop()
		close(t.delayTimerDone)

		if t.intervalTicker != nil {
			t.intervalTicker.Stop()
			close(t.intervalTickerDone)
		}
	}
}

func createTimer(delay time.Duration, interval time.Duration, job func()) *Timer {
	return NewTimer(delay, interval, job)
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
