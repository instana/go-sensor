// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2016

package instana

import (
	"runtime"
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

func newMeter(logger LeveledLogger) *meterS {
	logger.Debug("initializing meter")

	return &meterS{
		done: make(chan struct{}, 1),
	}
}

func (m *meterS) Run(collectInterval time.Duration) {
	ticker := time.NewTicker(collectInterval)
	defer ticker.Stop()
	for {
		select {
		case <-m.done:
			return
		case <-ticker.C:
			if sensor.Agent().Ready() {
				go func() {
					_ = sensor.Agent().SendMetrics(m.collectMetrics())
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
