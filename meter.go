// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2016

package instana

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/instana/go-sensor/acceptor"
)

const (
	// Metrics transmission interval constraints (in seconds)
	defaultTransmissionInterval = 1
	minTransmissionInterval     = 1
	maxTransmissionInterval     = 600
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
	numGC    atomic.Uint32
	once     sync.Once
	stopOnce sync.Once
	done     chan struct{}
}

// MetricsOptions contains configuration for metrics collection and transmission.
// This configuration is managed internally and populated from agent configuration.
type MetricsOptions struct {
	mu                   sync.RWMutex
	transmissionInterval time.Duration
}

// getTransmissionInterval returns the current metrics transmission interval.
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
// This is an internal method called when agent configuration is received during
// the initial handshake. Valid range: minTransmissionInterval-maxTransmissionInterval seconds.
// Values outside the range are clamped to the nearest bound.
func (m *MetricsOptions) setTransmissionInterval(seconds int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if seconds < minTransmissionInterval {
		defaultLogger.Warn("poll_rate value from agent (", seconds, ") is less than minimum. Setting to minimum value of ", minTransmissionInterval, " second.")
		m.transmissionInterval = minTransmissionInterval * time.Second
		return
	}

	if seconds > maxTransmissionInterval {
		defaultLogger.Warn("poll_rate value from agent (", seconds, ") exceeds maximum. Setting to maximum value of ", maxTransmissionInterval, " seconds.")
		m.transmissionInterval = maxTransmissionInterval * time.Second
		return
	}

	m.transmissionInterval = time.Duration(seconds) * time.Second
	defaultLogger.Info("Metrics transmission interval set to ", seconds, " second(s) from agent configuration")
}

func newMeter(logger LeveledLogger) *meterS {
	logger.Debug("initializing meter")

	return &meterS{
		done: make(chan struct{}),
	}
}

// Run starts the metrics collection loop at the given interval.
// It is safe to call Run multiple times — only the first call starts the loop;
// subsequent calls (e.g. on agent reconnect) are ignored so the running loop
// continues uninterrupted with the original interval.
// The interval is fixed at the first call; changing poll_rate in the agent
// configuration after startup requires an application restart to take effect.
func (m *meterS) Run(collectInterval time.Duration) {
	if m == nil {
		return
	}
	m.once.Do(func() {
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
	})
}

// Stop shuts down the metrics collection loop. Safe to call multiple times.
func (m *meterS) Stop() {
	if m == nil {
		return
	}
	m.stopOnce.Do(func() { close(m.done) })
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

	if m.numGC.Load() < memStats.NumGC {
		ret.PauseNs = memStats.PauseNs[(memStats.NumGC+255)%256]
		m.numGC.Store(memStats.NumGC)
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
