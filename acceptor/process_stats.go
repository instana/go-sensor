package acceptor

import "github.com/instana/go-sensor/process"

// ProcessCPUStatsUpdate represents the CPU stats that have changed since the last measurement
type ProcessCPUStatsUpdate struct {
	User   *int `json:"user,omitempty"`
	System *int `json:"sys,omitempty"`
}

// NewProcessCPUStatsUpdate returns the fields that have been updated since the last measurement.
// It returns nil if nothing has changed.
func NewProcessCPUStatsUpdate(prev, next process.CPUStats) *ProcessCPUStatsUpdate {
	if prev == next {
		return nil
	}

	update := &ProcessCPUStatsUpdate{}
	if prev.System != next.System {
		update.System = &next.System
	}
	if prev.User != next.User {
		update.User = &next.User
	}

	return update
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
