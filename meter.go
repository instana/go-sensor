package instana

import (
	"runtime"
	"time"
)

// SnapshotS struct to hold snapshot data.
type SnapshotS struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	Root     string `json:"goroot"`
	MaxProcs int    `json:"maxprocs"`
	Compiler string `json:"compiler"`
	NumCPU   int    `json:"cpu"`
}

// MemoryS struct to hold snapshot data.
type MemoryS struct {
	Alloc         uint64  `json:"alloc"`
	TotalAlloc    uint64  `json:"total_alloc"`
	Sys           uint64  `json:"sys"`
	Lookups       uint64  `json:"lookups"`
	Mallocs       uint64  `json:"mallocs"`
	Frees         uint64  `json:"frees"`
	HeapAlloc     uint64  `json:"heap_alloc"`
	HeapSys       uint64  `json:"heap_sys"`
	HeapIdle      uint64  `json:"heap_idle"`
	HeapInuse     uint64  `json:"heap_in_use"`
	HeapReleased  uint64  `json:"heap_released"`
	HeapObjects   uint64  `json:"heap_objects"`
	PauseTotalNs  uint64  `json:"pause_total_ns"`
	PauseNs       uint64  `json:"pause_ns"`
	NumGC         uint32  `json:"num_gc"`
	GCCPUFraction float64 `json:"gc_cpu_fraction"`
}

// MetricsS struct to hold snapshot data.
type MetricsS struct {
	CgoCall   int64    `json:"cgo_call"`
	Goroutine int      `json:"goroutine"`
	Memory    *MemoryS `json:"memory"`
}

// EntityData struct to hold snapshot data.
type EntityData struct {
	PID      int        `json:"pid"`
	Snapshot *SnapshotS `json:"snapshot,omitempty"`
	Metrics  *MetricsS  `json:"metrics"`
}

type meterS struct {
	sensor *sensorS
	numGC  uint32
}

func newMeter(sensor *sensorS) *meterS {
	sensor.logger.Debug("initializing meter")

	meter := &meterS{
		sensor: sensor,
	}

	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for range ticker.C {
			if !meter.sensor.agent.Ready() {
				continue
			}

			go meter.sensor.agent.SendMetrics(meter.collectMetrics())
		}
	}()

	return meter
}

func (r *meterS) collectMemoryMetrics() *MemoryS {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	ret := &MemoryS{
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

	if r.numGC < memStats.NumGC {
		ret.PauseNs = memStats.PauseNs[(memStats.NumGC+255)%256]
		r.numGC = memStats.NumGC
	} else {
		ret.PauseNs = 0
	}

	return ret
}

func (r *meterS) collectMetrics() *MetricsS {
	return &MetricsS{
		CgoCall:   runtime.NumCgoCall(),
		Goroutine: runtime.NumGoroutine(),
		Memory:    r.collectMemoryMetrics()}
}
