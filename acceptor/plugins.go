// Package acceptor provides marshaling structs for Instana serverless acceptor API
package acceptor

import (
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
	ID               string              `json:"Id"`
	Command          string              `json:"Command"`
	CreatedAt        time.Time           `json:"Created"`
	StartedAt        time.Time           `json:"Started"`
	Image            string              `json:"Image"`
	Labels           aws.ContainerLabels `json:"Labels,omitempty"`
	Ports            string              `json:"Ports,omitempty"`
	PortBindings     string              `json:"PortBindings,omitempty"`
	Names            []string            `json:"Names,omitempty"`
	NetworkMode      string              `json:"NetworkMode,omitempty"`
	StorageDriver    string              `json:"StorageDriver,omitempty"`
	DockerVersion    string              `json:"docker_version,omitempty"`
	DockerAPIVersion string              `json:"docker_api_version,omitempty"`
	Memory           int                 `json:"Memory"`
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
	PID           int               `json:"pid"`
	Exec          string            `json:"exec"`
	Args          []string          `json:"args,omitempty"`
	Env           map[string]string `json:"env,omitempty"`
	User          string            `json:"user,omitempty"`
	Group         string            `json:"group,omitempty"`
	ContainerID   string            `json:"container,omitempty"`
	ContainerPid  int               `json:"containerPid,string,omitempty"`
	ContainerType string            `json:"containerType,omitempty"`
	Start         int64             `json:"start"`
	HostName      string            `json:"com.instana.plugin.host.name"`
	HostPID       int               `json:"com.instana.plugin.host.pid,string"`
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
