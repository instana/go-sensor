// (c) Copyright IBM Corp. 2023

//go:build go1.18
// +build go1.18

package instacosmos

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
	instana "github.com/instana/go-sensor"
)

// Client is the interface that wraps the methods of *azcosmos.Client
type Client interface {
	Endpoint() string
	NewContainer(t instana.TracerLogger, databaseID string, containerID string) (ContainerClient, error)
	NewDatabase(id string) (*azcosmos.DatabaseClient, error)
	CreateDatabase(
		ctx context.Context,
		databaseProperties azcosmos.DatabaseProperties,
		o *azcosmos.CreateDatabaseOptions) (azcosmos.DatabaseResponse, error)
	NewQueryDatabasesPager(query string, o *azcosmos.QueryDatabasesOptions) *runtime.Pager[azcosmos.QueryDatabasesResponse]
}

type instaClient struct {
	*azcosmos.Client
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
func NewClientWithKey(endpoint string,
	cred azcosmos.KeyCredential,
	o *azcosmos.ClientOptions) (Client, error) {
	client, err := azcosmos.NewClientWithKey(endpoint, cred, &azcosmos.ClientOptions{})
	if err != nil {
		return nil, err
	}
	return &instaClient{
		client,
	}, nil

}

// Endpoint return the cosmos service endpoint
func (ic *instaClient) Endpoint() string {
	return ic.Client.Endpoint()
}

// NewContainer returns the instance of instrumented *azcosmos.ContainerClient
// collector - instana go collector
// databaseID - azure cosmos database name
// containerID - azure cosmos container name
func (ic *instaClient) NewContainer(collector instana.TracerLogger, databaseID string, containerID string) (ContainerClient, error) {
	containerClient, err := ic.Client.NewContainer(databaseID, containerID)
	if err != nil {
		return nil, err
	}
	return &instaContainerClient{
		database:    databaseID,
		containerID: containerID,
		endpoint:    ic.Client.Endpoint(),
		T: newTracer(context.TODO(), collector, instana.DbConnDetails{
			DatabaseName: string(instana.CosmosSpanType),
			RawString:    ic.Client.Endpoint(),
		}),
		ContainerClient: containerClient,
	}, nil
}
