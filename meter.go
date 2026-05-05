// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2016

package instana

import (
	"runtime"
	"sync"
	"time"

	"github.com/instana/go-sensor/acceptor"
)

const (
	// Metrics transmission interval constraints (in seconds)
	defaultTransmissionInterval = 1
	minTransmissionInterval     = 1
	maxTransmissionInterval     = 3600
)

// SnapshotS struct to hold snapshot data
type SnapshotS acceptor.RuntimeInfo

// MemoryS struct to hold snapshot data
type MemoryS acceptor.MemoryStats

// MetricsS struct to hold snapshot data
type MetricsS acceptor.Metrics

// EntityData struct to hold snapshot data
type EntityData acceptor.GoProcessData

type meterS struct {
	numGC   uint32
	running bool
	done    chan struct{}
	mu      sync.Mutex
}

// MetricsOptions contains configuration for metrics collection and transmission.
// This configuration is managed internally and populated from agent configuration.
type MetricsOptions struct {
	mu                   sync.RWMutex
	transmissionInterval time.Duration
}

// GetTransmissionInterval returns the current metrics transmission interval.
// This value is configured through the agent's configuration.yaml file.
func (m *MetricsOptions) getTransmissionInterval() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.transmissionInterval == 0 {
		return defaultTransmissionInterval * time.Second
	}
	return m.transmissionInterval
}

// setTransmissionInterval sets the metrics transmission interval.
// This is an internal method called when agent configuration is received.
// Valid range: minTransmissionInterval-maxTransmissionInterval seconds.
// Values < minTransmissionInterval are set to minTransmissionInterval,
// values > maxTransmissionInterval are set to maxTransmissionInterval.
func (m *MetricsOptions) setTransmissionInterval(seconds int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Apply minimum value constraint
	if seconds < minTransmissionInterval {
		defaultLogger.Warn("poll_rate value from agent (", seconds, ") is less than minimum. Setting to minimum value of ", minTransmissionInterval, " second.")
		m.transmissionInterval = minTransmissionInterval * time.Second
		return
	}

	// Apply maximum value constraint
	if seconds > maxTransmissionInterval {
		defaultLogger.Warn("poll_rate value from agent (", seconds, ") exceeds maximum. Setting to maximum value of ", maxTransmissionInterval, " seconds.")
		m.transmissionInterval = maxTransmissionInterval * time.Second
		return
	}

	// Valid value within range
	m.transmissionInterval = time.Duration(seconds) * time.Second
	defaultLogger.Info("Metrics transmission interval set to ", seconds, " second(s) from agent configuration")
}

func newMeter(logger LeveledLogger) *meterS {
	logger.Debug("initializing meter")

	return &meterS{
		done: make(chan struct{}, 1),
	}
}

func (m *meterS) Run(collectInterval time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// If already running, stop first
	if m.running {
		close(m.done)
		m.running = false
	}

	// Create new channel and start
	m.done = make(chan struct{})
	m.running = true

	go func() {
		ticker := time.NewTicker(collectInterval)
		defer ticker.Stop()
		for {
			select {
			case <-m.done:
				return
			case <-ticker.C:
				if isAgentReady() {
					go func() {
						s, err := getSensor()
						if err != nil {
							defaultLogger.Error("meter: ", err.Error())
							return
						}

						_ = s.Agent().SendMetrics(m.collectMetrics())
					}()
				}
			}
		}
	}()
}

func (m *meterS) reset(interval time.Duration) {
	if m == nil {
		return
	}
	m.Stop()
	m.Run(interval)
}

func (m *meterS) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m == nil {
		return
	}

	if m.running {
		close(m.done)
		m.running = false
	}
}

func (m *meterS) collectMemoryMetrics() acceptor.MemoryStats {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	ret := acceptor.MemoryStats{
		Alloc:         memStats.Alloc,
		TotalAlloc:    memStats.TotalAlloc,
		Sys:           memStats.Sys,
		Lookups:       memStats.Lookups,
		Mallocs:       memStats.Mallocs,
		Frees:         memStats.Frees,
		HeapAlloc:     memStats.HeapAlloc,
		HeapSys:       memStats.HeapSys,
		HeapIdle:      memStats.HeapIdle,
		HeapInuse:     memStats.HeapInuse,
		HeapReleased:  memStats.HeapReleased,
		HeapObjects:   memStats.HeapObjects,
		PauseTotalNs:  memStats.PauseTotalNs,
		NumGC:         memStats.NumGC,
		GCCPUFraction: memStats.GCCPUFraction}

	if m.numGC < memStats.NumGC {
		ret.PauseNs = memStats.PauseNs[(memStats.NumGC+255)%256]
		m.numGC = memStats.NumGC
	}

	return ret
}

func (m *meterS) collectMetrics() acceptor.Metrics {
	return acceptor.Metrics{
		CgoCall:     runtime.NumCgoCall(),
		Goroutine:   runtime.NumGoroutine(),
		MemoryStats: m.collectMemoryMetrics(),
	}
}
