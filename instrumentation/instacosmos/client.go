package instacosmos

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
	instana "github.com/instana/go-sensor"
)

type Client interface {
	Endpoint() string
	NewContainer(t instana.TracerLogger, databaseId string, containerId string) (ContainerClient, error)
	NewDatabase(id string) (*azcosmos.DatabaseClient, error)
	CreateDatabase(
		ctx context.Context,
		databaseProperties azcosmos.DatabaseProperties,
		o *azcosmos.CreateDatabaseOptions) (azcosmos.DatabaseResponse, error)
	NewQueryDatabasesPager(query string, o *azcosmos.QueryDatabasesOptions) *runtime.Pager[azcosmos.QueryDatabasesResponse]
}

type InstaClient struct {
	*azcosmos.Client
}

func NewKeyCredential(key string) (azcosmos.KeyCredential, error) {
	return azcosmos.NewKeyCredential(key)
}

func NewClientWithKey(endpoint string,
	cred azcosmos.KeyCredential,
	o *azcosmos.ClientOptions) (Client, error) {
	client, err := azcosmos.NewClientWithKey(endpoint, cred, &azcosmos.ClientOptions{})
	if err != nil {
		return nil, err
	}
	return &InstaClient{
		client,
	}, nil

}

func (ic *InstaClient) Endpoint() string {
	return ic.Client.Endpoint()
}

func (ic *InstaClient) NewContainer(t instana.TracerLogger, databaseId string, containerId string) (ContainerClient, error) {
	containerClient, err := ic.Client.NewContainer(databaseId, containerId)
	if err != nil {
		return nil, err
	}
	return &InstaContainerClient{
		database:    databaseId,
		containerID: containerId,
		endpoint:    ic.Client.Endpoint(),
		T: newTracer(context.TODO(), t, instana.DbConnDetails{
			DatabaseName: databaseId,
			RawString:    ic.Client.Endpoint(),
		}),
		ContainerClient: containerClient,
	}, nil
}
