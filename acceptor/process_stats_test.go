package acceptor_test

import (
	"testing"

	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/process"
	"github.com/instana/testify/assert"
)

func TestNewProcessCPUStatsUpdate(t *testing.T) {
	stats := process.CPUStats{
		User:   1,
		System: 10,
	}

	t.Run("equal", func(t *testing.T) {
		assert.Nil(t, acceptor.NewProcessCPUStatsUpdate(stats, stats))
	})

	t.Run("changed", func(t *testing.T) {
		assert.Equal(t,
			&acceptor.ProcessCPUStatsUpdate{
				User:   &stats.User,
				System: &stats.System,
			},
			acceptor.NewProcessCPUStatsUpdate(process.CPUStats{}, stats),
		)
	})

	t.Run("changed some", func(t *testing.T) {
		assert.Equal(t,
			&acceptor.ProcessCPUStatsUpdate{
				System: &stats.System,
			},
			acceptor.NewProcessCPUStatsUpdate(process.CPUStats{
				User:   stats.User,
				System: stats.System * 2,
			}, stats),
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
