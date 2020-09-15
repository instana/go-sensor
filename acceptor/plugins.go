// Package acceptor provides marshaling structs for Instana serverless acceptor API
package acceptor

import (
	"strconv"
	"time"

	"github.com/instana/go-sensor/aws"
)

// PluginPayload represents the Instana acceptor message envelope containing plugin
// name and entity ID
type PluginPayload struct {
	Name     string      `json:"name"`
	EntityID string      `json:"entityId"`
	Data     interface{} `json:"data"`
}

// AWSContainerLimits is used to send container limits (CPU, memory) to the acceptor plugin
type AWSContainerLimits struct {
	CPU    int `json:"cpu"`
	Memory int `json:"memory"`
}

// ECSTaskData is a representation of an ECS task for com.instana.plugin.aws.ecs.task plugin
type ECSTaskData struct {
	TaskARN               string             `json:"taskArn"`
	ClusterARN            string             `json:"clusterArn"`
	AvailabilityZone      string             `json:"availabilityZone,omitempty"`
	TaskDefinition        string             `json:"taskDefinition"`
	TaskDefinitionVersion string             `json:"taskDefinitionVersion"`
	DesiredStatus         string             `json:"desiredStatus"`
	KnownStatus           string             `json:"knownStatus"`
	Limits                AWSContainerLimits `json:"limits"`
	PullStartedAt         time.Time          `json:"pullStartedAt"`
	PullStoppedAt         time.Time          `json:"pullStoppedAt"`
}

// NewECSTaskPluginPayload returns payload for the ECS task plugin of Instana acceptor
func NewECSTaskPluginPayload(entityID string, data ECSTaskData) PluginPayload {
	const pluginName = "com.instana.plugin.aws.ecs.task"

	return PluginPayload{
		Name:     pluginName,
		EntityID: entityID,
		Data:     data,
	}
}

// ECSContainerData is a representation of an ECS container for com.instana.plugin.aws.ecs.container plugin
type ECSContainerData struct {
	Runtime               string             `json:"runtime"`
	Instrumented          bool               `json:"instrumented,omitempty"`
	DockerID              string             `json:"dockerId"`
	DockerName            string             `json:"dockerName"`
	ContainerName         string             `json:"containerName"`
	Image                 string             `json:"image"`
	ImageID               string             `json:"imageId"`
	TaskARN               string             `json:"taskArn"`
	TaskDefinition        string             `json:"taskDefinition"`
	TaskDefinitionVersion string             `json:"taskDefinitionVersion"`
	ClusterARN            string             `json:"clusterArn"`
	DesiredStatus         string             `json:"desiredStatus"`
	KnownStatus           string             `json:"knownStatus"`
	Limits                AWSContainerLimits `json:"limits"`
	CreatedAt             time.Time          `json:"createdAt"`
	StartedAt             time.Time          `json:"startedAt"`
	Type                  string             `json:"type"`
}

// NewECSContainerPluginPayload returns payload for the ECS container plugin of Instana acceptor
func NewECSContainerPluginPayload(entityID string, data ECSContainerData) PluginPayload {
	const pluginName = "com.instana.plugin.aws.ecs.container"

	return PluginPayload{
		Name:     pluginName,
		EntityID: entityID,
		Data:     data,
	}
}

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

// ProcessData is a representation of a running process for com.instana.plugin.process plugin
type ProcessData struct {
	PID           int                          `json:"pid"`
	Exec          string                       `json:"exec"`
	Args          []string                     `json:"args,omitempty"`
	Env           map[string]string            `json:"env,omitempty"`
	User          string                       `json:"user,omitempty"`
	Group         string                       `json:"group,omitempty"`
	ContainerID   string                       `json:"container,omitempty"`
	ContainerPid  int                          `json:"containerPid,string,omitempty"`
	ContainerType string                       `json:"containerType,omitempty"`
	Start         int64                        `json:"start"`
	HostName      string                       `json:"com.instana.plugin.host.name"`
	HostPID       int                          `json:"com.instana.plugin.host.pid,string"`
	CPU           *ProcessCPUStatsDelta        `json:"cpu,omitempty"`
	Memory        *ProcessMemoryStatsUpdate    `json:"mem,omitempty"`
	OpenFiles     *ProcessOpenFilesStatsUpdate `json:"openFiles,omitempty"`
}

// NewProcessPluginPayload returns payload for the process plugin of Instana acceptor
func NewProcessPluginPayload(entityID string, data ProcessData) PluginPayload {
	const pluginName = "com.instana.plugin.process"

	return PluginPayload{
		Name:     pluginName,
		EntityID: entityID,
		Data:     data,
	}
}

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
