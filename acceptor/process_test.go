// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package acceptor_test

import (
	"testing"

	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/process"
	"github.com/stretchr/testify/assert"
)

func TestNewProcessPluginPayload(t *testing.T) {
	data := acceptor.ProcessData{
		PID: 42,
	}

	assert.Equal(t, acceptor.PluginPayload{
		Name:     "com.instana.plugin.process",
		EntityID: "id1",
		Data:     data,
	}, acceptor.NewProcessPluginPayload("id1", data))
}

func TestNewProcessCPUStatsDelta(t *testing.T) {
	stats := process.CPUStats{
		User:   1,
		System: 10,
	}

	t.Run("equal stats", func(t *testing.T) {
		assert.Nil(t, acceptor.NewProcessCPUStatsDelta(stats, stats, 2))
	})

	t.Run("same time", func(t *testing.T) {
		assert.Nil(t, acceptor.NewProcessCPUStatsDelta(process.CPUStats{}, stats, 0))
	})

	t.Run("increased insignificantly", func(t *testing.T) {
		assert.Nil(t, acceptor.NewProcessCPUStatsDelta(process.CPUStats{}, stats, 10000))
	})

	t.Run("decreased both", func(t *testing.T) {
		assert.Nil(t, acceptor.NewProcessCPUStatsDelta(stats, process.CPUStats{}, 2))
	})

	t.Run("increased", func(t *testing.T) {
		assert.Equal(t,
			&acceptor.ProcessCPUStatsDelta{
				User:   0.5,
				System: 5,
			},
			acceptor.NewProcessCPUStatsDelta(process.CPUStats{}, stats, 2),
		)
	})

	t.Run("one has increased insignificantly", func(t *testing.T) {
		assert.Equal(t,
			&acceptor.ProcessCPUStatsDelta{
				System: 0.01,
			},
			acceptor.NewProcessCPUStatsDelta(process.CPUStats{}, stats, 1000),
		)
	})

	t.Run("one has decreased", func(t *testing.T) {
		assert.Equal(t,
			&acceptor.ProcessCPUStatsDelta{
				User: 0.5,
			},
			acceptor.NewProcessCPUStatsDelta(stats, process.CPUStats{
				User:   stats.User * 2,
				System: stats.System / 2,
			}, 2),
		)
	})
}

func TestNewProcessMemoryStatsUpdate(t *testing.T) {
	stats := process.MemStats{
		Total:  1,
		Rss:    10,
		Shared: 100,
	}

	t.Run("equal", func(t *testing.T) {
		assert.Nil(t, acceptor.NewProcessMemoryStatsUpdate(stats, stats))
	})

	t.Run("changed", func(t *testing.T) {
		assert.Equal(t,
			&acceptor.ProcessMemoryStatsUpdate{
				Total:  &stats.Total,
				Rss:    &stats.Rss,
				Shared: &stats.Shared,
			},
			acceptor.NewProcessMemoryStatsUpdate(process.MemStats{}, stats),
		)
	})

	t.Run("changed some", func(t *testing.T) {
		assert.Equal(t,
			&acceptor.ProcessMemoryStatsUpdate{
				Total:  &stats.Total,
				Shared: &stats.Shared,
			},
			acceptor.NewProcessMemoryStatsUpdate(process.MemStats{
				Total:  stats.Total * 2,
				Rss:    stats.Rss,
				Shared: stats.Shared * 2,
			}, stats),
		)
	})
}

func TestNewProcessOpenFilesStatsUpdate(t *testing.T) {
	stats := process.ResourceLimits{
		OpenFiles: process.LimitedResource{
			Current: 1,
			Max:     10,
		},
	}

	t.Run("equal", func(t *testing.T) {
		assert.Nil(t, acceptor.NewProcessOpenFilesStatsUpdate(stats, stats))
	})

	t.Run("changed", func(t *testing.T) {
		assert.Equal(t,
			&acceptor.ProcessOpenFilesStatsUpdate{
				Current: &stats.OpenFiles.Current,
				Max:     &stats.OpenFiles.Max,
			},
			acceptor.NewProcessOpenFilesStatsUpdate(process.ResourceLimits{}, stats),
		)
	})

	t.Run("changed some", func(t *testing.T) {
		assert.Equal(t,
			&acceptor.ProcessOpenFilesStatsUpdate{
				Current: &stats.OpenFiles.Current,
			},
			acceptor.NewProcessOpenFilesStatsUpdate(process.ResourceLimits{
				OpenFiles: process.LimitedResource{
					Current: stats.OpenFiles.Current * 2,
					Max:     stats.OpenFiles.Max,
				},
			}, stats),
		)
	})
}
