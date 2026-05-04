// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2016

package instana

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/instana/go-sensor/acceptor"
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
	numGC uint32
	done  chan struct{}
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
		return 1 * time.Second // default
	}
	return m.transmissionInterval
}

// setTransmissionInterval sets the metrics transmission interval.
// This is an internal method called when agent configuration is received.
// Only 1 or 5 seconds are valid values; others default to 1 second.
func (m *MetricsOptions) setTransmissionInterval(seconds int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate: only 1 or 5 seconds allowed
	if seconds != 1 && seconds != 5 {
		defaultLogger.Warn("Invalid poll_rate value from agent: ", seconds, ", using default 1 second. Valid values are 1 or 5.")
		m.transmissionInterval = 1 * time.Second
		return
	}

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
	fmt.Println("collectInterval: ", collectInterval)
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
}

func (m *meterS) Stop() {
	m.done <- struct{}{}
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
