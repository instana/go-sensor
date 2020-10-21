package acceptor

import (
	"time"

	"github.com/instana/go-sensor/aws"
	"github.com/instana/go-sensor/docker"
)

// DockerData is a representation of a Docker container for com.instana.plugin.docker plugin
type DockerData struct {
	ID               string                             `json:"Id"`
	Command          string                             `json:"Command"`
	CreatedAt        time.Time                          `json:"Created"`
	StartedAt        time.Time                          `json:"Started"`
	Image            string                             `json:"Image"`
	Labels           aws.ContainerLabels                `json:"Labels,omitempty"`
	Ports            string                             `json:"Ports,omitempty"`
	PortBindings     string                             `json:"PortBindings,omitempty"`
	Names            []string                           `json:"Names,omitempty"`
	NetworkMode      string                             `json:"NetworkMode,omitempty"`
	StorageDriver    string                             `json:"StorageDriver,omitempty"`
	DockerVersion    string                             `json:"docker_version,omitempty"`
	DockerAPIVersion string                             `json:"docker_api_version,omitempty"`
	Network          *DockerNetworkAggregatedStatsDelta `json:"network,omitempty"`
	CPU              *DockerCPUStatsDelta               `json:"cpu,omitempty"`
	Memory           *DockerMemoryStatsUpdate           `json:"memory,omitempty"`
	BlockIO          *DockerBlockIOStatsDelta           `json:"blkio,omitempty"`
}

// NewDockerPluginPayload returns payload for the Docker plugin of Instana acceptor
func NewDockerPluginPayload(entityID string, data DockerData) PluginPayload {
	const pluginName = "com.instana.plugin.docker"

	return PluginPayload{
		Name:     pluginName,
		EntityID: entityID,
		Data:     data,
	}
}

// DockerNetworkStatsDelta represents the difference between two network interface stats
type DockerNetworkStatsDelta struct {
	Bytes   int `json:"bytes,omitempty"`
	Packets int `json:"packets,omitempty"`
	Dropped int `json:"dropped,omitempty"`
	Errors  int `json:"errors,omitempty"`
}

// IsZero returns true is there is no difference between interface stats
func (d DockerNetworkStatsDelta) IsZero() bool {
	return d.Bytes == 0 && d.Packets == 0 && d.Dropped == 0 && d.Errors == 0
}

// DockerNetworkStatsDelta represents the difference between two network interface stats
type DockerNetworkAggregatedStatsDelta struct {
	Rx *DockerNetworkStatsDelta `json:"rx,omitempty"`
	Tx *DockerNetworkStatsDelta `json:"tx,omitempty"`
}

// NewDockerNetworkAggregatedStatsDelta calculates the aggregated difference between two snapshots
// of network interface stats. It returns nil if aggregated stats for both snapshots are equal.
func NewDockerNetworkAggregatedStatsDelta(prev, next map[string]docker.ContainerNetworkStats) *DockerNetworkAggregatedStatsDelta {
	var rxDelta, txDelta DockerNetworkStatsDelta

	for _, stats := range next {
		rxDelta.Bytes += stats.RxBytes
		rxDelta.Packets += stats.RxPackets
		rxDelta.Dropped += stats.RxDropped
		rxDelta.Errors += stats.RxErrors

		txDelta.Bytes += stats.TxBytes
		txDelta.Packets += stats.TxPackets
		txDelta.Dropped += stats.TxDropped
		txDelta.Errors += stats.TxErrors
	}

	for _, stats := range prev {
		rxDelta.Bytes -= stats.RxBytes
		rxDelta.Packets -= stats.RxPackets
		rxDelta.Dropped -= stats.RxDropped
		rxDelta.Errors -= stats.RxErrors

		txDelta.Bytes -= stats.TxBytes
		txDelta.Packets -= stats.TxPackets
		txDelta.Dropped -= stats.TxDropped
		txDelta.Errors -= stats.TxErrors
	}

	if rxDelta.IsZero() && txDelta.IsZero() {
		return nil
	}

	var delta DockerNetworkAggregatedStatsDelta
	if !rxDelta.IsZero() {
		delta.Rx = &rxDelta
	}
	if !txDelta.IsZero() {
		delta.Tx = &txDelta
	}

	return &delta
}

// DockerCPUStatsDelta represents the difference between two CPU usage stats
type DockerCPUStatsDelta struct {
	Total           float64 `json:"total_usage,omitempty"`
	User            float64 `json:"user_usage,omitempty"`
	System          float64 `json:"system_usage,omitempty"`
	ThrottlingCount int     `json:"throttling_count,omitempty"`
	ThrottlingTime  int     `json:"throttling_time,omitempty"`
}

// NewDockerCPUStatsDelta calculates the difference between two CPU usage stats. It returns nil if stats are equal.
func NewDockerCPUStatsDelta(prev, next docker.ContainerCPUStats) *DockerCPUStatsDelta {
	if prev == next {
		return nil
	}

	delta := DockerCPUStatsDelta{
		ThrottlingCount: next.Throttling.Periods - prev.Throttling.Periods,
		ThrottlingTime:  next.Throttling.Time - prev.Throttling.Time,
	}

	if systemDelta := next.System - prev.System; systemDelta > 0 {
		if totalDelta := next.Usage.Total - prev.Usage.Total; totalDelta > 0 {
			delta.Total = (float64(totalDelta) / float64(systemDelta)) * float64(next.OnlineCPUs)
		}
		if kernelDelta := next.Usage.Kernel - prev.Usage.Kernel; kernelDelta > 0 {
			delta.System = (float64(kernelDelta) / float64(systemDelta)) * float64(next.OnlineCPUs)
		}
		if userDelta := next.Usage.User - prev.Usage.User; userDelta > 0 {
			delta.User = (float64(userDelta) / float64(systemDelta)) * float64(next.OnlineCPUs)
		}
	}

	return &delta
}

// DockerMemoryStatsUpdate represents the memory stats that have changed since the last measurement
type DockerMemoryStatsUpdate struct {
	ActiveAnon   *int `json:"active_anon,omitempty"`
	ActiveFile   *int `json:"active_file,omitempty"`
	InactiveAnon *int `json:"inactive_anon,omitempty"`
	InactiveFile *int `json:"inactive_file,omitempty"`
	TotalCache   *int `json:"total_cache,omitempty"`
	TotalRss     *int `json:"total_rss,omitempty"`
	Usage        *int `json:"usage,omitempty"`
	MaxUsage     *int `json:"max_usage,omitempty"`
	Limit        *int `json:"limit,omitempty"`
}

// NewDockerMemoryStatsUpdate returns the fields that have been updated since the last measurement.
// It returns nil if nothing has changed.
func NewDockerMemoryStatsUpdate(prev, next docker.ContainerMemoryStats) *DockerMemoryStatsUpdate {
	if prev == next {
		return nil
	}

	var delta DockerMemoryStatsUpdate
	if prev.Usage != next.Usage {
		delta.Usage = &next.Usage
	}
	if prev.MaxUsage != next.MaxUsage {
		delta.MaxUsage = &next.MaxUsage
	}
	if prev.Limit != next.Limit {
		delta.Limit = &next.Limit
	}

	if prev.Stats == next.Stats {
		return &delta
	}

	if prev.Stats.ActiveAnon != next.Stats.ActiveAnon {
		delta.ActiveAnon = &next.Stats.ActiveAnon
	}
	if prev.Stats.ActiveFile != next.Stats.ActiveFile {
		delta.ActiveFile = &next.Stats.ActiveFile
	}
	if prev.Stats.InactiveAnon != next.Stats.InactiveAnon {
		delta.InactiveAnon = &next.Stats.InactiveAnon
	}
	if prev.Stats.InactiveFile != next.Stats.InactiveFile {
		delta.InactiveFile = &next.Stats.InactiveFile
	}
	if prev.Stats.TotalCache != next.Stats.TotalCache {
		delta.TotalCache = &next.Stats.TotalCache
	}
	if prev.Stats.TotalRss != next.Stats.TotalRss {
		delta.TotalRss = &next.Stats.TotalRss
	}

	return &delta
}

// DockerMemoryStatsDelta represents the difference between two block I/O usage stats
type DockerBlockIOStatsDelta struct {
	Read  int `json:"blk_read,omitempty"`
	Write int `json:"blk_write,omitempty"`
}

// IsZero returns true if both usage stats are equal
func (d DockerBlockIOStatsDelta) IsZero() bool {
	return d.Read == 0 && d.Write == 0
}

// NewDockerMemoryStatsDelta sums up block I/O reads and writes and calculates the difference between two stat snapshots.
// It returns nil if aggregated stats are equal.
func NewDockerBlockIOStatsDelta(prev, next docker.ContainerBlockIOStats) *DockerBlockIOStatsDelta {
	var delta DockerBlockIOStatsDelta

	for _, stat := range next.ServiceBytes {
		switch stat.Operation {
		case docker.BlockIOReadOp:
			delta.Read += stat.Value
		case docker.BlockIOWriteOp:
			delta.Write += stat.Value
		}
	}

	for _, stat := range prev.ServiceBytes {
		switch stat.Operation {
		case docker.BlockIOReadOp:
			delta.Read -= stat.Value
		case docker.BlockIOWriteOp:
			delta.Write -= stat.Value
		}
	}

	if delta.IsZero() {
		return nil
	}

	return &delta
}
