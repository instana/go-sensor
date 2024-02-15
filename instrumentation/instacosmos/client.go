// (c) Copyright IBM Corp. 2024

//go:build go1.18
// +build go1.18

package instacosmos

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
	instana "github.com/instana/go-sensor"
)

const emptyPrimaryKey string = ""

// Client is the interface that wraps the methods of *azcosmos.Client
type Client interface {
	Endpoint() string
	NewContainer(databaseID string, containerID string) (ContainerClient, error)
	NewDatabase(id string) (*azcosmos.DatabaseClient, error)
	CreateDatabase(
		ctx context.Context,
		databaseProperties azcosmos.DatabaseProperties,
		o *azcosmos.CreateDatabaseOptions) (azcosmos.DatabaseResponse, error)
	NewQueryDatabasesPager(query string, o *azcosmos.QueryDatabasesOptions) *runtime.Pager[azcosmos.QueryDatabasesResponse]
}

type instaClient struct {
	*azcosmos.Client
	collector instana.TracerLogger
}

// NewKeyCredential creates an KeyCredential containing the
// account's primary or secondary key.
func NewKeyCredential(key string) (azcosmos.KeyCredential, error) {
	return azcosmos.NewKeyCredential(key)
}

// NewClientWithKey creates an instance of instrumented *azcosmos.Client
// endpoint - The cosmos service endpoint to use.
// cred - The credential used to authenticate with the cosmos service.
// options - Optional Cosmos client options.  Pass nil to accept default values.
func NewClientWithKey(collector instana.TracerLogger,
	endpoint string,
	cred azcosmos.KeyCredential,
	o *azcosmos.ClientOptions) (Client, error) {
	client, err := azcosmos.NewClientWithKey(endpoint, cred, o)
	if err != nil {
		return nil, err
	}
	return &instaClient{
		client,
		collector,
	}, nil

}

// NewClient creates an instance of instrumented *azcosmos.Client
// endpoint - The cosmos service endpoint to use.
// cred - The credential used to authenticate with the cosmos service.
// options - Optional Cosmos client options.  Pass nil to accept default values.
func NewClient(collector instana.TracerLogger, endpoint string, cred azcore.TokenCredential, o *azcosmos.ClientOptions) (Client, error) {
	client, err := azcosmos.NewClient(endpoint, cred, o)
	if err != nil {
		return nil, err
	}
	return &instaClient{
		client,
		collector,
	}, nil
}

// NewClientFromConnectionString creates an instance of instrumented *azcosmos.Client
// connectionString - The cosmos service connection string.
// options - Optional Cosmos client options.  Pass nil to accept default values.
func NewClientFromConnectionString(collector instana.TracerLogger, connectionString string, o *azcosmos.ClientOptions) (Client, error) {
	client, err := azcosmos.NewClientFromConnectionString(connectionString, o)
	if err != nil {
		return nil, err
	}
	return &instaClient{
		client,
		collector,
	}, nil
}

// Endpoint return the cosmos service endpoint
func (ic *instaClient) Endpoint() string {
	return ic.Client.Endpoint()
}

// NewContainer returns the instance of instrumented *azcosmos.DatabaseClient
// id - azure cosmos database name
func (ic *instaClient) NewDatabase(id string) (*azcosmos.DatabaseClient, error) {
	return ic.Client.NewDatabase(id)
}

// NewContainer returns the instance of instrumented *azcosmos.ContainerClient
// databaseID - azure cosmos database name
// containerID - azure cosmos container name
func (ic *instaClient) NewContainer(databaseID string, containerID string) (ContainerClient, error) {
	containerClient, err := ic.Client.NewContainer(databaseID, containerID)
	if err != nil {
		return nil, err
	}
	return &instaContainerClient{
		database:    databaseID,
		containerID: containerID,
		endpoint:    ic.Client.Endpoint(),
		t: newTracer(context.TODO(), ic.collector, instana.DbConnDetails{
			DatabaseName: string(instana.CosmosSpanType),
			RawString:    ic.Client.Endpoint(),
		}),
		ContainerClient: containerClient,
		PartitionKey:    NewPartitionKey(emptyPrimaryKey),
	}, nil
}

// CreateDatabase creates a new database in azure account
// ctx - The context for the request.
// databaseProperties - The definition of the database
// o - Options for the create database operation.
func (ic *instaClient) CreateDatabase(ctx context.Context,
	dbProperties azcosmos.DatabaseProperties, o *azcosmos.CreateDatabaseOptions) (azcosmos.DatabaseResponse, error) {
	return ic.Client.CreateDatabase(ctx, dbProperties, nil)
}
