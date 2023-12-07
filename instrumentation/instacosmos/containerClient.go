// (c) Copyright IBM Corp. 2023

//go:build go1.18
// +build go1.18

package instacosmos

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/tracing"
	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
)

// ContainerClient is the interface that wraps the methods of *azcosmos.ContainerClient
type ContainerClient interface {
	CreateItem(
		ctx context.Context,
		partitionKey azcosmos.PartitionKey,
		item []byte,
		o *azcosmos.ItemOptions) (azcosmos.ItemResponse, error)
}

type instaContainerClient struct {
	database    string
	containerID string
	endpoint    string
	T           tracing.Tracer
	*azcosmos.ContainerClient
}

// DatabaseID returns the azure cosmos database name
func (icc *instaContainerClient) DatabaseID() string {
	return icc.database
}

// ContainerID returns the azure cosmos container name
func (icc *instaContainerClient) ContainerID() string {
	return icc.containerID
}

// Endpoint returns the cosmos service endpoint
func (icc *instaContainerClient) Endpoint() string {
	return icc.endpoint
}

// CreateItem creates an item in a Cosmos container.
// ctx - The context for the request.
// partitionKey - The partition key for the item.
// item - The item to create.
// o - Options for the operation.
func (icc *instaContainerClient) CreateItem(ctx context.Context,
	partitionKey azcosmos.PartitionKey,
	item []byte,
	o *azcosmos.ItemOptions) (azcosmos.ItemResponse, error) {
	_, s := icc.T.Start(ctx, "CREATE_ITEM", &tracing.SpanOptions{
		Attributes: []tracing.Attribute{
			{
				Key:   dataCommand,
				Value: "INSERT INTO " + icc.containerID,
			}},
	})
	defer s.End()

	icc.setAttributes(s)

	resp, err := icc.ContainerClient.CreateItem(ctx, partitionKey, item, o)
	icc.setStatus(s, resp.RawResponse.StatusCode)
	if err != nil {
		icc.setError(s, err)
	}
	return resp, err
}

// helper functions
func (icc *instaContainerClient) setAttributes(s tracing.Span) {
	attrs := []tracing.Attribute{
		{
			Key:   dataDB,
			Value: icc.database,
		},
		{
			Key:   dataURL,
			Value: icc.endpoint,
		},
		{
			Key:   dataType,
			Value: "Query",
		},
	}
	s.SetAttributes(attrs...)
}

func (icc *instaContainerClient) setError(s tracing.Span, err error) {
	errAttrs := []tracing.Attribute{
		{
			Key:   dataError,
			Value: err.Error(),
		},
		{
			Key:   "string(ext.Error)",
			Value: err.Error(),
		},
	}

	s.SetAttributes(errAttrs...)
	s.AddEvent(errorEvent, tracing.Attribute{
		Key:   dataError,
		Value: err.Error(),
	})
}

func (icc *instaContainerClient) setStatus(s tracing.Span, statusCode int) {
	s.SetStatus(tracing.SpanStatus(statusCode), dataReturnCode)
}
