package instacosmos_test

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instacosmos"
)

const (
	dbName        = "trace-data"
	containerName = "spans"
)

// This example shows how to instrument Azure Cosmos DB using instacosmos library
func Example() {

	t := instana.InitCollector(&instana.Options{
		Service:           "cosmos-example",
		EnableAutoProfile: true,
	})

	// creates an KeyCredential containing the account's primary or secondary key.
	cred, err := instacosmos.NewKeyCredential(key)
	if err != nil {
		log.Fatal("Failed to create KeyCredential:", err)
	}

	client, err := instacosmos.NewClientWithKey(endpoint, cred, &azcosmos.ClientOptions{})
	if err != nil {
		log.Fatal("Failed to create cosmos DB client:", err)
	}

	// Create container client instance
	containerClient, err := client.NewContainer(t, dbName, containerName)
	if err != nil {
		log.Fatal("Failed to create a container client:", err)
	}

	pk := azcosmos.NewPartitionKeyString("newPartitionKey")

	item := map[string]string{
		"id":             "anId",
		"value":          "2",
		"myPartitionKey": "newPartitionKey",
	}

	marshalled, err := json.Marshal(item)
	if err != nil {
		log.Fatal("Failed to marshal span data:", err)
	}

	// create the item in the Cosmos DB container
	itemResponse, err := containerClient.CreateItem(context.Background(), pk, marshalled, nil)
	if err != nil {
		log.Print("Failed to create the item:", err)
	}

	log.Printf("Status %d. Item %v created. ActivityId %s. Consuming %v Request Units.\n",
		itemResponse.RawResponse.StatusCode,
		pk, itemResponse.ActivityID,
		itemResponse.RequestCharge)
}
