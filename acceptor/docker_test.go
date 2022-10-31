// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package acceptor_test

import (
	"testing"

	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/docker"
	"github.com/stretchr/testify/assert"
)

func TestNewDockerPluginPayload(t *testing.T) {
	data := acceptor.DockerData{
		ID: "docker1",
	}

	assert.Equal(t, acceptor.PluginPayload{
		Name:     "com.instana.plugin.docker",
		EntityID: "id1",
		Data:     data,
	}, acceptor.NewDockerPluginPayload("id1", data))
}

func TestNewDockerNetworkAggregatedStatsDelta(t *testing.T) {
	stats := map[string]docker.ContainerNetworkStats{
		"eth0": {
			RxBytes:   1,
			RxDropped: 10,
			RxErrors:  100,
			RxPackets: 1000,
			TxBytes:   10000,
			TxDropped: 100000,
			TxErrors:  1000000,
			TxPackets: 10000000,
		},
		"eth1": {
			RxBytes:   2,
			RxDropped: 20,
			RxErrors:  200,
			RxPackets: 2000,
			TxBytes:   20000,
			TxDropped: 200000,
			TxErrors:  2000000,
			TxPackets: 20000000,
		},
		"eth2": {
			RxBytes:   3,
			RxDropped: 30,
			RxErrors:  300,
			RxPackets: 3000,
			TxBytes:   30000,
			TxDropped: 300000,
			TxErrors:  3000000,
			TxPackets: 30000000,
		},
	}

	t.Run("equal", func(t *testing.T) {
		assert.Nil(t, acceptor.NewDockerNetworkAggregatedStatsDelta(stats, stats))
	})

	t.Run("increase", func(t *testing.T) {
		assert.Equal(t,
			&acceptor.DockerNetworkAggregatedStatsDelta{
				Rx: &acceptor.DockerNetworkStatsDelta{
					Bytes:   6,
					Dropped: 60,
					Errors:  600,
					Packets: 6000,
				},
				Tx: &acceptor.DockerNetworkStatsDelta{
					Bytes:   60000,
					Dropped: 600000,
					Errors:  6000000,
					Packets: 60000000,
				},
			},
			acceptor.NewDockerNetworkAggregatedStatsDelta(nil, stats),
		)
	})

	t.Run("decrease", func(t *testing.T) {
		assert.Equal(t,
			&acceptor.DockerNetworkAggregatedStatsDelta{
				Rx: &acceptor.DockerNetworkStatsDelta{
					Bytes:   -6,
					Dropped: -60,
					Errors:  -600,
					Packets: -6000,
				},
				Tx: &acceptor.DockerNetworkStatsDelta{
					Bytes:   -60000,
					Dropped: -600000,
					Errors:  -6000000,
					Packets: -60000000,
				},
			},
			acceptor.NewDockerNetworkAggregatedStatsDelta(stats, nil),
		)
	})
}

func TestNewDockerCPUStatsDelta(t *testing.T) {
	stats := docker.ContainerCPUStats{
		Usage: docker.CPUUsageStats{
			Total:  1,
			User:   10,
			Kernel: 100,
		},
		Throttling: docker.CPUThrottlingStats{
			Periods: 1000,
			Time:    10000,
		},
		System:     100000,
		OnlineCPUs: 16,
	}

	t.Run("equal", func(t *testing.T) {
		assert.Nil(t, acceptor.NewDockerCPUStatsDelta(stats, stats))
	})

	t.Run("increase", func(t *testing.T) {
		assert.Equal(t,
			&acceptor.DockerCPUStatsDelta{
				Total:           0.00016,
				User:            0.0016,
				System:          0.016,
				ThrottlingCount: 1000,
				ThrottlingTime:  10000,
			},
			acceptor.NewDockerCPUStatsDelta(docker.ContainerCPUStats{}, stats),
		)
	})

	t.Run("decrease", func(t *testing.T) {
		assert.Equal(t,
			&acceptor.DockerCPUStatsDelta{
				Total:           0,
				User:            0,
				System:          0,
				ThrottlingCount: -1000,
				ThrottlingTime:  -10000,
			},
			acceptor.NewDockerCPUStatsDelta(stats, docker.ContainerCPUStats{}),
		)
	})
}

func TestNewDockerMemoryStatsDelta(t *testing.T) {
	stats := docker.ContainerMemoryStats{
		Stats: docker.MemoryStats{
			ActiveAnon:   1,
			ActiveFile:   10,
			InactiveAnon: 100,
			InactiveFile: 1000,
			TotalRss:     10000,
			TotalCache:   100000,
		},
		MaxUsage: 1000000,
		Usage:    10000000,
		Limit:    100000000,
	}

	t.Run("equal", func(t *testing.T) {
		assert.Nil(t, acceptor.NewDockerMemoryStatsUpdate(stats, stats))
	})

	t.Run("changed", func(t *testing.T) {
		assert.Equal(t,
			&acceptor.DockerMemoryStatsUpdate{
				ActiveAnon:   &stats.Stats.ActiveAnon,
				ActiveFile:   &stats.Stats.ActiveFile,
				InactiveAnon: &stats.Stats.InactiveAnon,
				InactiveFile: &stats.Stats.InactiveFile,
				TotalRss:     &stats.Stats.TotalRss,
				TotalCache:   &stats.Stats.TotalCache,
				MaxUsage:     &stats.MaxUsage,
				Usage:        &stats.Usage,
				Limit:        &stats.Limit,
			},
			acceptor.NewDockerMemoryStatsUpdate(docker.ContainerMemoryStats{}, stats),
		)
	})

	t.Run("changed some", func(t *testing.T) {
		assert.Equal(t,
			&acceptor.DockerMemoryStatsUpdate{
				ActiveFile:   &stats.Stats.ActiveFile,
				InactiveFile: &stats.Stats.InactiveFile,
				TotalCache:   &stats.Stats.TotalCache,
				MaxUsage:     &stats.MaxUsage,
				Limit:        &stats.Limit,
			},
			acceptor.NewDockerMemoryStatsUpdate(docker.ContainerMemoryStats{
				Stats: docker.MemoryStats{
					ActiveAnon:   1,
					ActiveFile:   20,
					InactiveAnon: 100,
					InactiveFile: 2000,
					TotalRss:     10000,
					TotalCache:   200000,
				},
				MaxUsage: 2000000,
				Usage:    10000000,
				Limit:    200000000,
			}, stats),
		)
	})
}

func TestNewDockerBlockIOStatsDelta(t *testing.T) {
	stats := docker.ContainerBlockIOStats{
		ServiceBytes: []docker.BlockIOOpStats{
			{Operation: docker.BlockIOReadOp, Value: 1},
			{Operation: docker.BlockIOWriteOp, Value: 10},
			{Operation: docker.BlockIOReadOp, Value: 100},
		},
	}

	t.Run("equal", func(t *testing.T) {
		assert.Nil(t, acceptor.NewDockerBlockIOStatsDelta(stats, stats))
	})

	t.Run("increase", func(t *testing.T) {
		assert.Equal(t,
			&acceptor.DockerBlockIOStatsDelta{
				Read:  101,
				Write: 10,
			},
			acceptor.NewDockerBlockIOStatsDelta(docker.ContainerBlockIOStats{}, stats),
		)
	})

	t.Run("decrease", func(t *testing.T) {
		assert.Equal(t,
			&acceptor.DockerBlockIOStatsDelta{
				Read:  -101,
				Write: -10,
			},
			acceptor.NewDockerBlockIOStatsDelta(stats, docker.ContainerBlockIOStats{}),
		)
	})

	t.Run("equal, shuffled ops", func(t *testing.T) {
		assert.Nil(t, acceptor.NewDockerBlockIOStatsDelta(stats, docker.ContainerBlockIOStats{
			ServiceBytes: []docker.BlockIOOpStats{
				{Operation: docker.BlockIOReadOp, Value: 100},
				{Operation: docker.BlockIOReadOp, Value: 1},
				{Operation: docker.BlockIOWriteOp, Value: 10},
			},
		}))
	})
}
