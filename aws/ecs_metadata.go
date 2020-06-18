package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ContainerLimits represents the resource limits specified at the task level
type ContainerLimits struct {
	CPU    int `json:"CPU"`
	Memory int `json:"Memory"`
}

type ContainerLabels struct {
	Cluster               string `json:"com.amazonaws.ecs.cluster"`
	TaskARN               string `json:"com.amazonaws.ecs.task-arn"`
	TaskDefinition        string `json:"com.amazonaws.ecs.task-definition-family"`
	TaskDefinitionVersion string `json:"com.amazonaws.ecs.task-definition-version"`
}

type ContainerNetwork struct {
	Mode          string   `json:"NetworkMode"`
	IPv4Addresses []string `json:"IPv4Addresses"`
}

// ECSContainerV3Metadata represents the ECS container metadata as described in
// https://docs.aws.amazon.com/AmazonECS/latest/developerguide/container-metadata.html#metadata-file-format
type ECSContainerMetadata struct {
	DockerID        string          `json:"DockerId"`
	Name            string          `json:"Name"`
	DockerName      string          `json:"DockerName"`
	Image           string          `json:"Image"`
	ImageID         string          `json:"ImageID"`
	DesiredStatus   string          `json:"DesiredStatus"`
	KnownStatus     string          `json:"KnownStatus"`
	Limits          ContainerLimits `json:"Limits"`
	CreatedAt       time.Time       `json:"CreatedAt"`
	StartedAt       time.Time       `json:"StartedAt"`
	Type            string          `json:"Type"`
	ContainerLabels `json:"Labels"`
	Networks        []ContainerNetwork `json:"Networks"`
}

// ECSTaskMetadata represents the ECS task metadata as described in
// https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-metadata-endpoint-v3.html#task-metadata-endpoint-v3-response
type ECSTaskMetadata struct {
	TaskARN       string                 `json:"TaskARN"`
	Family        string                 `json:"Family"`
	Revision      string                 `json:"Revision"`
	DesiredStatus string                 `json:"DesiredStatus"`
	KnownStatus   string                 `json:"KnownStatus"`
	Containers    []ECSContainerMetadata `json:"Containers"`
	PullStartedAt time.Time              `json:"PullStartedAt"`
	PullStoppedAt time.Time              `json:"PullStoppedAt"`
}

// ECSMetadataProvider retireves ECS service metadata from the ECS_CONTAINER_METADATA_URI endpoint
type ECSMetadataProvider struct {
	Endpoint string
	client   *http.Client
}

// NewECSMetadataProvider initializes a new ECSMetadataClient with given endpoint and HTTP client.
// If there is no HTTP client provided, the provider will use http.DefaultClient
func NewECSMetadataProvider(endpoint string, c *http.Client) *ECSMetadataProvider {
	if c == nil {
		c = http.DefaultClient
	}

	return &ECSMetadataProvider{
		Endpoint: endpoint,
		client:   c,
	}
}

// ContainerMetadata returns ECS metadata for current container
func (c *ECSMetadataProvider) ContainerMetadata(ctx context.Context) (ECSContainerMetadata, error) {
	var data ECSContainerMetadata

	req, err := http.NewRequest(http.MethodGet, c.Endpoint, nil)
	if err != nil {
		return data, fmt.Errorf("failed to prepare request: %s", err)
	}

	body, err := c.executeRequest(req.WithContext(ctx))
	if err != nil {
		return data, fmt.Errorf("failed to fetch container metadata: %s", err)
	}
	defer body.Close()

	if err := json.NewDecoder(body).Decode(&data); err != nil {
		return data, fmt.Errorf("malformed container metadata response: %s", err)
	}

	return data, nil
}

// TaskMetadata returns ECS metadata for current task
func (c *ECSMetadataProvider) TaskMetadata(ctx context.Context) (ECSTaskMetadata, error) {
	var data ECSTaskMetadata

	req, err := http.NewRequest(http.MethodGet, c.Endpoint+"/task", nil)
	if err != nil {
		return data, fmt.Errorf("failed to prepare request: %s", err)
	}

	body, err := c.executeRequest(req.WithContext(ctx))
	if err != nil {
		return data, fmt.Errorf("failed to fetch task metadata: %s", err)
	}
	defer body.Close()

	if err := json.NewDecoder(body).Decode(&data); err != nil {
		return data, fmt.Errorf("malformed task metadata response: %s", err)
	}

	return data, nil
}

func (c *ECSMetadataProvider) executeRequest(req *http.Request) (io.ReadCloser, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("the endpoint responded with %s", resp.Status)
	}

	return resp.Body, nil
}
