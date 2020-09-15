package acceptor

import "github.com/instana/go-sensor/process"

// ProcessCPUStatsUpdate represents the CPU stats that have changed since the last measurement
type ProcessCPUStatsDelta struct {
	User   float64 `json:"user,omitempty"`
	System float64 `json:"sys,omitempty"`
}

// NewDockerCPUStatsDelta calculates the difference between two CPU usage stats.
// It returns nil if stats are equal or if the stats were taken at the same time (ticks).
func NewProcessCPUStatsDelta(prev, next process.CPUStats, ticksElapsed int) *ProcessCPUStatsDelta {
	if prev == next || ticksElapsed == 0 {
		return nil
	}

	delta := &ProcessCPUStatsDelta{}
	if prev.System < next.System {
		delta.System = float64(next.System-prev.System) / float64(ticksElapsed)
	}
	if prev.User < next.User {
		delta.User = float64(next.User-prev.User) / float64(ticksElapsed)
	}

	return delta
}

// ProcessMemoryStatsUpdate represents the memory stats that have changed since the last measurement
type ProcessMemoryStatsUpdate struct {
	Total  *int `json:"virtual,omitempty"`
	Rss    *int `json:"resident,omitempty"`
	Shared *int `json:"share,omitempty"`
}

// NewProcessMemoryStatsUpdate returns the fields that have been updated since the last measurement.
// It returns nil if nothing has changed.
func NewProcessMemoryStatsUpdate(prev, next process.MemStats) *ProcessMemoryStatsUpdate {
	if prev == next {
		return nil
	}

	update := &ProcessMemoryStatsUpdate{}
	if prev.Total != next.Total {
		update.Total = &next.Total
	}
	if prev.Rss != next.Rss {
		update.Rss = &next.Rss
	}
	if prev.Shared != next.Shared {
		update.Shared = &next.Shared
	}

	return update
}

// ProcessOpenFilesStatsUpdate represents the open file stats and limits that have changed since the last measurement
type ProcessOpenFilesStatsUpdate struct {
	Current *int `json:"current,omitempty"`
	Max     *int `json:"max,omitempty"`
}

// NewProcessOpenFilesStatsUpdate returns the (process.ResourceLimits).OpenFiles fields that have been updated
// since the last measurement. It returns nil if nothing has changed.
func NewProcessOpenFilesStatsUpdate(prev, next process.ResourceLimits) *ProcessOpenFilesStatsUpdate {
	if prev.OpenFiles == next.OpenFiles {
		return nil
	}

	update := &ProcessOpenFilesStatsUpdate{}
	if prev.OpenFiles.Current != next.OpenFiles.Current {
		update.Current = &next.OpenFiles.Current
	}
	if prev.OpenFiles.Max != next.OpenFiles.Max {
		update.Max = &next.OpenFiles.Max
	}

	return update
}
