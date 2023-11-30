package instacosmos

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/tracing"
	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
)

type ContainerClient interface {
	CreateItem(
		ctx context.Context,
		partitionKey azcosmos.PartitionKey,
		item []byte,
		o *azcosmos.ItemOptions) (azcosmos.ItemResponse, error)
}

type InstaContainerClient struct {
	database    string
	containerID string
	endpoint    string
	T           tracing.Tracer
	*azcosmos.ContainerClient
}

func (icc *InstaContainerClient) DatabaseID() string {
	return icc.database
}

func (icc *InstaContainerClient) ContainerID() string {
	return icc.containerID
}

func (icc *InstaContainerClient) EndPoint() string {
	return icc.endpoint
}

func (icc *InstaContainerClient) CreateItem(ctx context.Context,
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

	icc.SetAttributes(s)

	resp, err := icc.ContainerClient.CreateItem(ctx, partitionKey, item, o)
	icc.SetStatus(s, resp.RawResponse.StatusCode)
	if err != nil {
		icc.SetError(s, err)
	}
	return resp, err
}

func (icc *InstaContainerClient) SetAttributes(s tracing.Span) {
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

func (icc *InstaContainerClient) SetError(s tracing.Span, err error) {
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

func (icc *InstaContainerClient) SetStatus(s tracing.Span, statusCode int) {
	s.SetStatus(tracing.SpanStatus(statusCode), dataReturnCode)
}
