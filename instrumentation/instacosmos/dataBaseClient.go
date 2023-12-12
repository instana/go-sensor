// (c) Copyright IBM Corp. 2023

//go:build go1.18
// +build go1.18

package instacosmos

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/tracing"
	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
)

// DatabaseClient is the interface that wraps the methods of *azcosmos.DatabaseClient
type DatabaseClient interface {
	CreateContainer(
		ctx context.Context,
		containerProperties azcosmos.ContainerProperties,
		o *azcosmos.CreateContainerOptions) (azcosmos.ContainerResponse, error)
	Delete(
		ctx context.Context,
		o *azcosmos.DeleteDatabaseOptions) (azcosmos.DatabaseResponse, error)
	// ID() string
	// NewContainer(id string) (*ContainerClient, error)
	// NewQueryContainersPager(query string,
	// 	o *azcosmos.QueryContainersOptions) *runtime.Pager[azcosmos.QueryContainersResponse]
	// Read(
	// 	ctx context.Context,
	// 	o *azcosmos.ReadDatabaseOptions) (azcosmos.DatabaseResponse, error)
	// ReadThroughput(
	// 	ctx context.Context,
	// 	o *azcosmos.ThroughputOptions) (azcosmos.ThroughputResponse, error)
	// ReplaceThroughput(
	// 	ctx context.Context,
	// 	throughputProperties azcosmos.ThroughputProperties,
	// 	o *azcosmos.ThroughputOptions) (azcosmos.ThroughputResponse, error)
}

type instaDatabaseClient struct {
	database string
	endpoint string
	T        tracing.Tracer
	*azcosmos.DatabaseClient
}

// Delete a Cosmos database.
// ctx - The context for the request.
// o - Options for Read operation.
func (idc *instaDatabaseClient) Delete(ctx context.Context,
	o *azcosmos.DeleteDatabaseOptions) (azcosmos.DatabaseResponse, error) {
	return idc.DatabaseClient.Delete(ctx, o)
}

// CreateContainer creates a container in the Cosmos database.
// ctx - The context for the request.
// containerProperties - The properties for the container.
// o - Options for the create container operation.
func (idc *instaDatabaseClient) CreateContainer(
	ctx context.Context,
	containerProperties azcosmos.ContainerProperties,
	o *azcosmos.CreateContainerOptions) (azcosmos.ContainerResponse, error) {
	return idc.DatabaseClient.CreateContainer(ctx, containerProperties, o)
}
