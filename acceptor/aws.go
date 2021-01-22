// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package acceptor

import "time"

// AWSContainerLimits is used to send container limits (CPU, memory) to the acceptor plugin
type AWSContainerLimits struct {
	CPU    int `json:"cpu"`
	Memory int `json:"memory"`
}

// ECSTaskData is a representation of an ECS task for com.instana.plugin.aws.ecs.task plugin
type ECSTaskData struct {
	TaskARN               string                 `json:"taskArn"`
	ClusterARN            string                 `json:"clusterArn"`
	AvailabilityZone      string                 `json:"availabilityZone,omitempty"`
	InstanaZone           string                 `json:"instanaZone,omitempty"`
	TaskDefinition        string                 `json:"taskDefinition"`
	TaskDefinitionVersion string                 `json:"taskDefinitionVersion"`
	DesiredStatus         string                 `json:"desiredStatus"`
	KnownStatus           string                 `json:"knownStatus"`
	Limits                AWSContainerLimits     `json:"limits"`
	PullStartedAt         time.Time              `json:"pullStartedAt"`
	PullStoppedAt         time.Time              `json:"pullStoppedAt"`
	Tags                  map[string]interface{} `json:"tags,omitempty"`
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

// NewAWSLambdaPluginPayload returns payload for the AWS Lambda plugin of Instana acceptor
func NewAWSLambdaPluginPayload(entityID string) PluginPayload {
	const pluginName = "com.instana.plugin.aws.lambda"

	return PluginPayload{
		Name:     pluginName,
		EntityID: entityID,
	}
}
