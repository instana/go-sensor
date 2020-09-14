package docker

import (
	"fmt"
	"strings"
	"time"
)

// ContainerNetworkStats represents networking stats for a container.
//
// See https://docs.docker.com/config/containers/runmetrics/#network-metrics
type ContainerNetworkStats struct {
	RxBytes   int `json:"rx_bytes"`
	RxDropped int `json:"rx_dropped"`
	RxErrors  int `json:"rx_errors"`
	RxPackets int `json:"rx_packets"`
	TxBytes   int `json:"tx_bytes"`
	TxDropped int `json:"tx_dropped"`
	TxErrors  int `json:"tx_errors"`
	TxPackets int `json:"tx_packets"`
}

// MemoryStats represents the cgroups memory stats.
//
// See https://docs.docker.com/config/containers/runmetrics/#metrics-from-cgroups-memory-cpu-block-io
type MemoryStats struct {
	ActiveAnon   int `json:"active_anon"`
	ActiveFile   int `json:"active_file"`
	InactiveAnon int `json:"inactive_anon"`
	InactiveFile int `json:"inactive_file"`
	TotalRss     int `json:"total_rss"`
	TotalCache   int `json:"total_cache"`
}

// ContainerMemoryStats represents the memory usage stats for a container.
//
// See https://docs.docker.com/config/containers/runmetrics/#metrics-from-cgroups-memory-cpu-block-io
type ContainerMemoryStats struct {
	Stats    MemoryStats `json:"stats"`
	MaxUsage int         `json:"max_usage"`
	Usage    int         `json:"usage"`
	Limit    int         `json:"limit"`
}

type blockIOOp uint8

func (biop *blockIOOp) UnmarshalJSON(data []byte) error {
	switch strings.ToLower(string(data)) {
	case `"read"`:
		*biop = BlockIOReadOp
	case `"write"`:
		*biop = BlockIOWriteOp
	default:
		return fmt.Errorf("unexpected block i/o op %s", string(data))
	}

	return nil
}

func (biop blockIOOp) MarshalJSON() ([]byte, error) {
	switch biop {
	case BlockIOReadOp:
		return []byte(`"read"`), nil
	case BlockIOWriteOp:
		return []byte(`"write"`), nil
	default:
		return []byte(`"unknown"`), nil
	}
}

// Valid block I/O operations
const (
	BlockIOReadOp = iota + 1
	BlockIOWriteOp
)

// BlockIOOpStats represents cgroups stats for a block I/O operation type. The zero-value does not represent
// a valid value until the Operation is specified.
type BlockIOOpStats struct {
	Operation blockIOOp `json:"op"`
	Value     int       `json:"value"`
}

// ContainerBlockIOStats represents the block I/O usage stats for a container.
//
// See https://docs.docker.com/config/containers/runmetrics/#metrics-from-cgroups-memory-cpu-block-io
type ContainerBlockIOStats struct {
	ServiceBytes []BlockIOOpStats `json:"io_service_bytes_recursive"`
}

// CPUThrottlingStats represents the cgroups CPU usage stats.
//
// See https://docs.docker.com/config/containers/runmetrics/#metrics-from-cgroups-memory-cpu-block-io
type CPUUsageStats struct {
	Total  int `json:"total_usage"`
	Kernel int `json:"usage_in_kernelmode"`
	User   int `json:"usage_in_usermode"`
}

// CPUThrottlingStats represents the cgroupd CPU throttling stats.
//
// See https://docs.docker.com/config/containers/runmetrics/#metrics-from-cgroups-memory-cpu-block-io
type CPUThrottlingStats struct {
	Periods int `json:"periods"`
	Time    int `json:"throttled_time"`
}

// ContainerCPUStats represents the CPU usage stats for a container.
//
// See https://docs.docker.com/config/containers/runmetrics/#metrics-from-cgroups-memory-cpu-block-io
type ContainerCPUStats struct {
	Usage      CPUUsageStats      `json:"cpu_usage"`
	Throttling CPUThrottlingStats `json:"throttling_data"`
	System     int                `json:"system_cpu_usage"`
	OnlineCPUs int                `json:"online_cpus"`
}

// ContainerStats represents the container resource usage stats as returned by Docker Engine.
//
// See https://docs.docker.com/engine/api/v1.30/#operation/ContainerExport
type ContainerStats struct {
	ReadAt   time.Time                        `json:"read"`
	Networks map[string]ContainerNetworkStats `json:"networks"`
	Memory   ContainerMemoryStats             `json:"memory_stats"`
	BlockIO  ContainerBlockIOStats            `json:"blkio_stats"`
	CPU      ContainerCPUStats                `json:"cpu_stats"`
}
