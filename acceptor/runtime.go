package acceptor

import "strconv"

// RuntimeInfo represents Go runtime info to be sent to com.insana.plugin.golang
type RuntimeInfo struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	Root     string `json:"goroot"`
	MaxProcs int    `json:"maxprocs"`
	Compiler string `json:"compiler"`
	NumCPU   int    `json:"cpu"`
}

// MemoryStats represents Go runtime memory stats to be sent to com.insana.plugin.golang
type MemoryStats struct {
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

// Metrics represents Go process metrics to be sent to com.insana.plugin.golang
type Metrics struct {
	CgoCall     int64 `json:"cgo_call"`
	Goroutine   int   `json:"goroutine"`
	MemoryStats `json:"memory"`
}

// GoProcessData is a representation of a Go process for com.instana.plugin.golang plugin
type GoProcessData struct {
	PID      int          `json:"pid"`
	Snapshot *RuntimeInfo `json:"snapshot,omitempty"`
	Metrics  Metrics      `json:"metrics"`
}

// NewGoProcessPluginPayload returns payload for the Go process plugin of Instana acceptor
func NewGoProcessPluginPayload(data GoProcessData) PluginPayload {
	const pluginName = "com.instana.plugin.golang"

	return PluginPayload{
		Name:     pluginName,
		EntityID: strconv.Itoa(data.PID),
		Data:     data,
	}
}
