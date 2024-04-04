// (c) Copyright IBM Corp. 2024

//go:build go1.18
// +build go1.18

package instacosmos

import (
	"context"
	"errors"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/tracing"
	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
)

// data operation types
const (
	Query   = "Query"
	Write   = "Write"
	Execute = "Execute"
	Update  = "Update"
	Upsert  = "Upsert"
	Replace = "Replace"
)

// ContainerClient is the interface that wraps the methods of *azcosmos.ContainerClient
type ContainerClient interface {
	PartitionKey
	DatabaseID() string
	CreateItem(
		ctx context.Context,
		partitionKey azcosmos.PartitionKey,
		item []byte,
		o *azcosmos.ItemOptions) (azcosmos.ItemResponse, error)
	DeleteItem(ctx context.Context,
		partitionKey azcosmos.PartitionKey,
		itemId string,
		o *azcosmos.ItemOptions) (azcosmos.ItemResponse, error)
	ExecuteTransactionalBatch(ctx context.Context,
		b azcosmos.TransactionalBatch,
		o *azcosmos.TransactionalBatchOptions) (azcosmos.TransactionalBatchResponse, error)
	ID() string
	NewQueryItemsPager(query string,
		partitionKey azcosmos.PartitionKey,
		o *azcosmos.QueryOptions) *runtime.Pager[azcosmos.QueryItemsResponse]
	NewTransactionalBatch(partitionKey azcosmos.PartitionKey) azcosmos.TransactionalBatch
	PatchItem(
		ctx context.Context,
		partitionKey azcosmos.PartitionKey,
		itemId string,
		ops azcosmos.PatchOperations,
		o *azcosmos.ItemOptions) (azcosmos.ItemResponse, error)
	Read(
		ctx context.Context,
		o *azcosmos.ReadContainerOptions) (azcosmos.ContainerResponse, error)
	ReadItem(
		ctx context.Context,
		partitionKey azcosmos.PartitionKey,
		itemId string,
		o *azcosmos.ItemOptions) (azcosmos.ItemResponse, error)
	ReadThroughput(
		ctx context.Context,
		o *azcosmos.ThroughputOptions) (azcosmos.ThroughputResponse, error)
	Replace(
		ctx context.Context,
		containerProperties azcosmos.ContainerProperties,
		o *azcosmos.ReplaceContainerOptions) (azcosmos.ContainerResponse, error)
	ReplaceItem(
		ctx context.Context,
		partitionKey azcosmos.PartitionKey,
		itemId string,
		item []byte,
		o *azcosmos.ItemOptions) (azcosmos.ItemResponse, error)
	ReplaceThroughput(
		ctx context.Context,
		throughputProperties azcosmos.ThroughputProperties,
		o *azcosmos.ThroughputOptions) (azcosmos.ThroughputResponse, error)
	UpsertItem(
		ctx context.Context,
		partitionKey azcosmos.PartitionKey,
		item []byte,
		o *azcosmos.ItemOptions) (azcosmos.ItemResponse, error)
}

type instaContainerClient struct {
	PartitionKey
	database    string
	containerID string
	endpoint    string
	t           tracing.Tracer
	*azcosmos.ContainerClient
}

// DatabaseID returns the azure cosmos database name
func (icc *instaContainerClient) DatabaseID() string {
	return icc.database
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
	_, s := icc.t.Start(ctx, "CREATE", &tracing.SpanOptions{
		Attributes: []tracing.Attribute{
			{
				Key:   dataCommand,
				Value: "INSERT INTO " + icc.containerID,
			}},
	})
	defer s.End()

	icc.setAttributes(s, Write)

	resp, err := icc.ContainerClient.CreateItem(ctx, partitionKey, item, o)
	if err != nil {
		icc.setError(s, err)
	}

	if resp.RawResponse != nil {
		icc.setStatus(s, resp.RawResponse.StatusCode)
	}
	return resp, err
}

// DeleteItem deletes an item in a Cosmos container.
// ctx - The context for the request.
// partitionKey - The partition key for the item.
// itemId - The id of the item to delete.
// o - Options for the operation.
func (icc *instaContainerClient) DeleteItem(ctx context.Context,
	partitionKey azcosmos.PartitionKey,
	itemId string,
	o *azcosmos.ItemOptions) (azcosmos.ItemResponse, error) {
	_, s := icc.t.Start(ctx, "DELETE", &tracing.SpanOptions{
		Attributes: []tracing.Attribute{
			{
				Key:   dataCommand,
				Value: "DELETE FROM " + icc.containerID,
			}},
	})
	defer s.End()

	icc.setAttributes(s, Write)

	resp, err := icc.ContainerClient.DeleteItem(ctx, partitionKey, itemId, o)
	if err != nil {
		icc.setError(s, err)
	}

	if resp.RawResponse != nil {
		icc.setStatus(s, resp.RawResponse.StatusCode)
	}
	return resp, err
}

// ExecuteTransactionalBatch executes a transactional batch.
// Once executed, verify the Success property of the response to determine if the batch was committed
func (icc *instaContainerClient) ExecuteTransactionalBatch(ctx context.Context,
	b azcosmos.TransactionalBatch,
	o *azcosmos.TransactionalBatchOptions) (azcosmos.TransactionalBatchResponse, error) {

	_, s := icc.t.Start(ctx, "EXECUTE", &tracing.SpanOptions{
		Attributes: []tracing.Attribute{
			{
				Key:   dataCommand,
				Value: "Execute",
			}},
	})
	defer s.End()

	icc.setAttributes(s, Execute)

	resp, err := icc.ContainerClient.ExecuteTransactionalBatch(ctx, b, o)
	if err != nil {
		icc.setError(s, err)
	}

	if resp.RawResponse != nil {
		icc.setStatus(s, resp.RawResponse.StatusCode)
	}
	return resp, err

}

// ID returns the identifier of the Cosmos container.
func (icc *instaContainerClient) ID() string {
	return icc.ContainerClient.ID()
}

// NewQueryItemsPager executes a single partition query in a Cosmos container.
// query - The SQL query to execute.
// partitionKey - The partition key to scope the query on.
// o - Options for the operation.
func (icc *instaContainerClient) NewQueryItemsPager(query string,
	partitionKey azcosmos.PartitionKey,
	o *azcosmos.QueryOptions) *runtime.Pager[azcosmos.QueryItemsResponse] {

	_, s := icc.t.Start(context.TODO(), "QUERY", &tracing.SpanOptions{
		Attributes: []tracing.Attribute{
			{
				Key:   dataCommand,
				Value: query,
			}},
	})
	defer s.End()

	icc.setAttributes(s, Query)

	resp := icc.ContainerClient.NewQueryItemsPager(query, partitionKey, o)
	return resp

}

// NewTransactionalBatch creates a batch of operations to be committed as a single unit.
// See https://docs.microsoft.com/azure/cosmos-db/sql/transactional-batch
func (icc *instaContainerClient) NewTransactionalBatch(partitionKey azcosmos.PartitionKey) azcosmos.TransactionalBatch {
	resp := icc.ContainerClient.NewTransactionalBatch(partitionKey)
	return resp
}

// PatchItem patches an item in a Cosmos container.
// ctx - The context for the request.
// partitionKey - The partition key for the item.
// itemId - The id of the item to patch.
// ops - Operations to perform on the patch
// o - Options for the operation.
func (icc *instaContainerClient) PatchItem(
	ctx context.Context,
	partitionKey azcosmos.PartitionKey,
	itemId string,
	ops azcosmos.PatchOperations,
	o *azcosmos.ItemOptions) (azcosmos.ItemResponse, error) {

	_, s := icc.t.Start(ctx, "PATCH_ITEM", &tracing.SpanOptions{
		Attributes: []tracing.Attribute{
			{
				Key:   dataCommand,
				Value: "UPDATE " + icc.containerID,
			}},
	})
	defer s.End()

	icc.setAttributes(s, Update)

	resp, err := icc.ContainerClient.PatchItem(ctx, partitionKey, itemId, ops, o)
	if err != nil {
		icc.setError(s, err)
	}

	if resp.RawResponse != nil {
		icc.setStatus(s, resp.RawResponse.StatusCode)
	}
	return resp, err

}

// Read obtains the information for a Cosmos container.
// ctx - The context for the request.
// o - Options for the operation.
func (icc *instaContainerClient) Read(
	ctx context.Context,
	o *azcosmos.ReadContainerOptions) (azcosmos.ContainerResponse, error) {

	_, s := icc.t.Start(ctx, "READ", &tracing.SpanOptions{
		Attributes: []tracing.Attribute{
			{
				Key:   dataCommand,
				Value: "Read",
			}},
	})
	defer s.End()

	icc.setAttributes(s, Query)

	resp, err := icc.ContainerClient.Read(ctx, o)
	if err != nil {
		icc.setError(s, err)
	}

	if resp.RawResponse != nil {
		icc.setStatus(s, resp.RawResponse.StatusCode)
	}
	return resp, err

}

// ReadItem reads an item in a Cosmos container.
// ctx - The context for the request.
// partitionKey - The partition key for the item.
// itemId - The id of the item to read.
// o - Options for the operation.
func (icc *instaContainerClient) ReadItem(
	ctx context.Context,
	partitionKey azcosmos.PartitionKey,
	itemId string,
	o *azcosmos.ItemOptions) (azcosmos.ItemResponse, error) {

	_, s := icc.t.Start(ctx, "READ_ITEM", &tracing.SpanOptions{
		Attributes: []tracing.Attribute{
			{
				Key:   dataCommand,
				Value: "SELECT " + itemId + " FROM " + icc.containerID,
			}},
	})
	defer s.End()

	icc.setAttributes(s, Query)

	resp, err := icc.ContainerClient.ReadItem(ctx, partitionKey, itemId, o)
	if err != nil {
		icc.setError(s, err)
	}

	if resp.RawResponse != nil {
		icc.setStatus(s, resp.RawResponse.StatusCode)
	}
	return resp, err

}

// ReadThroughput obtains the provisioned throughput information for the container.
// ctx - The context for the request.
// o - Options for the operation.
func (icc *instaContainerClient) ReadThroughput(
	ctx context.Context,
	o *azcosmos.ThroughputOptions) (azcosmos.ThroughputResponse, error) {

	_, s := icc.t.Start(ctx, "READ_THROUGHPUT", &tracing.SpanOptions{
		Attributes: []tracing.Attribute{
			{
				Key:   dataCommand,
				Value: "ReadThroughput",
			}},
	})
	defer s.End()

	icc.setAttributes(s, Query)

	resp, err := icc.ContainerClient.ReadThroughput(ctx, o)
	if err != nil {
		icc.setError(s, err)
	}

	if resp.RawResponse != nil {
		icc.setStatus(s, resp.RawResponse.StatusCode)
	}
	return resp, err

}

// Replace a Cosmos container.
// ctx - The context for the request.
// o - Options for the operation.
func (icc *instaContainerClient) Replace(
	ctx context.Context,
	containerProperties azcosmos.ContainerProperties,
	o *azcosmos.ReplaceContainerOptions) (azcosmos.ContainerResponse, error) {

	_, s := icc.t.Start(ctx, "REPLACE", &tracing.SpanOptions{
		Attributes: []tracing.Attribute{
			{
				Key:   dataCommand,
				Value: "Replace",
			}},
	})
	defer s.End()

	icc.setAttributes(s, Replace)

	resp, err := icc.ContainerClient.Replace(ctx, containerProperties, o)
	if err != nil {
		icc.setError(s, err)
	}

	if resp.RawResponse != nil {
		icc.setStatus(s, resp.RawResponse.StatusCode)
	}
	return resp, err

}

// ReplaceItem replaces an item in a Cosmos container.
// ctx - The context for the request.
// partitionKey - The partition key of the item to replace.
// itemId - The id of the item to replace.
// item - The content to be used to replace.
// o - Options for the operation.
func (icc *instaContainerClient) ReplaceItem(
	ctx context.Context,
	partitionKey azcosmos.PartitionKey,
	itemId string,
	item []byte,
	o *azcosmos.ItemOptions) (azcosmos.ItemResponse, error) {

	_, s := icc.t.Start(ctx, "REPLACE_ITEM", &tracing.SpanOptions{
		Attributes: []tracing.Attribute{
			{
				Key:   dataCommand,
				Value: "ReplaceItem",
			}},
	})
	defer s.End()

	icc.setAttributes(s, Replace)

	resp, err := icc.ContainerClient.ReplaceItem(ctx, partitionKey, itemId, item, o)
	if err != nil {
		icc.setError(s, err)
	}

	if resp.RawResponse != nil {
		icc.setStatus(s, resp.RawResponse.StatusCode)
	}
	return resp, err

}

// ReplaceThroughput updates the provisioned throughput for the container.
// ctx - The context for the request.
// throughputProperties - The throughput configuration of the container.
// o - Options for the operation.
func (icc *instaContainerClient) ReplaceThroughput(
	ctx context.Context,
	throughputProperties azcosmos.ThroughputProperties,
	o *azcosmos.ThroughputOptions) (azcosmos.ThroughputResponse, error) {

	_, s := icc.t.Start(ctx, "REPLACE_THROUGHPUT", &tracing.SpanOptions{
		Attributes: []tracing.Attribute{
			{
				Key:   dataCommand,
				Value: "ReplaceThroughput",
			}},
	})
	defer s.End()

	icc.setAttributes(s, Replace)

	resp, err := icc.ContainerClient.ReplaceThroughput(ctx, throughputProperties, o)
	if err != nil {
		icc.setError(s, err)
	}

	if resp.RawResponse != nil {
		icc.setStatus(s, resp.RawResponse.StatusCode)
	}
	return resp, err

}

// UpsertItem creates or replaces an item in a Cosmos container.
// ctx - The context for the request.
// partitionKey - The partition key for the item.
// item - The item to upsert.
// o - Options for the operation.
func (icc *instaContainerClient) UpsertItem(
	ctx context.Context,
	partitionKey azcosmos.PartitionKey,
	item []byte,
	o *azcosmos.ItemOptions) (azcosmos.ItemResponse, error) {

	_, s := icc.t.Start(ctx, "UPSERT", &tracing.SpanOptions{
		Attributes: []tracing.Attribute{
			{
				Key:   dataCommand,
				Value: "UpsertItem",
			}},
	})
	defer s.End()

	icc.setAttributes(s, Upsert)

	resp, err := icc.ContainerClient.UpsertItem(ctx, partitionKey, item, o)
	if err != nil {
		icc.setError(s, err)
	}

	if resp.RawResponse != nil {
		icc.setStatus(s, resp.RawResponse.StatusCode)
	}
	return resp, err

}

// helper functions
func (icc *instaContainerClient) setAttributes(s tracing.Span, dt string) {
	attrs := []tracing.Attribute{
		{
			Key:   dataDB,
			Value: icc.database + ":" + icc.containerID,
		},
		{
			Key:   dataURL,
			Value: icc.endpoint,
		},
		{
			Key:   dataType,
			Value: dt,
		},
		{
			Key:   dataObj,
			Value: icc.containerID,
		},
		{
			Key:   dataPartitionKey,
			Value: icc.PartitionKey.getPartitionKey(),
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
		Key:   "error",
		Value: err.Error(),
	})

	// setting status code of error response
	var responseErr = new(azcore.ResponseError)
	if ok := errors.As(err, &responseErr); ok {
		icc.setStatus(s, responseErr.StatusCode)
	}
}

func (icc *instaContainerClient) setStatus(s tracing.Span, statusCode int) {
	s.SetStatus(tracing.SpanStatus(statusCode), dataReturnCode)
}
