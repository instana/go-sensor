package gcloud

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ProjectMetadata represents Google Cloud project metadata returned by /computeMetadata/v1/project endpoint.
//
// See https://cloud.google.com/compute/docs/storing-retrieving-metadata for further details.
type ProjectMetadata struct {
	ProjectID        string `json:"projectId"`
	NumericProjectID int    `json:"numericProjectId"`
}

// InstanceMetadata represents Google Cloud project metadata returned by /computeMetadata/v1/instance endpoint.
//
// See https://cloud.google.com/compute/docs/storing-retrieving-metadata for further details.
type InstanceMetadata struct {
	ID     string `json:"id"`
	Region string `json:"region"`
}

// ComputeMetadata represents Google Cloud project metadata returned by /computeMetadata/v1 endpoint.
//
// See https://cloud.google.com/compute/docs/storing-retrieving-metadata for further details.
type ComputeMetadata struct {
	Project  ProjectMetadata  `json:"project"`
	Instance InstanceMetadata `json:"instance"`
}

// ComputeMetadataProvider retireves Google Cloud service compute metadata from provided endpoint
type ComputeMetadataProvider struct {
	Endpoint string
	client   *http.Client
}

// NewComputeMetadataProvider initializes a new ComputeMetadataClient with given endpoint and HTTP client.
// If there is no HTTP client provided, the provider will use http.DefaultClient
func NewComputeMetadataProvider(endpoint string, c *http.Client) *ComputeMetadataProvider {
	if c == nil {
		c = http.DefaultClient
	}

	return &ComputeMetadataProvider{
		Endpoint: endpoint,
		client:   c,
	}
}

// ComputeMetadata returns compute metadata for current instance
func (p *ComputeMetadataProvider) ComputeMetadata(ctx context.Context) (ComputeMetadata, error) {
	var data ComputeMetadata

	req, err := http.NewRequest(http.MethodGet, p.Endpoint+"/computeMetadata/v1?recursive=true", nil)
	if err != nil {
		return data, fmt.Errorf("failed to prepare request: %s", err)
	}

	req.Header.Set("Metadata-Flavor", "Google")

	resp, err := p.client.Do(req)
	if err != nil {
		return data, fmt.Errorf("failed to fetch compute metadata: %s", err)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return data, fmt.Errorf("malformed compute metadata response: %s", err)
	}

	return data, nil
}
