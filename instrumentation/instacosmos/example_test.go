// (c) Copyright IBM Corp. 2024

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

// given empty string for example. Replace the credentials securely in your code
var (
	cosmosEndpoint = ""
	cosmosKey      = ""
)

// This example shows how to instrument Azure Cosmos DB using instacosmos library
func Example() {

	t := instana.InitCollector(&instana.Options{
		Service: "cosmos-example",
	})

	// creates an KeyCredential containing the account's primary or secondary key.
	cred, err := instacosmos.NewKeyCredential(cosmosKey)
	if err != nil {
		log.Fatal("Failed to create KeyCredential:", err)
	}

	client, err := instacosmos.NewClientWithKey(t, cosmosEndpoint, cred, &azcosmos.ClientOptions{})
	if err != nil {
		log.Fatal("Failed to create cosmos DB client:", err)
	}

	// Create container client instance
	containerClient, err := client.NewContainer(dbName, containerName)
	if err != nil {
		log.Fatal("Failed to create a container client:", err)
	}

	pk := containerClient.NewPartitionKeyString("newPartitionKey")

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
	// NOTE: All Cosmos DB operations requires a parent context to be passed in.
	// Otherwise, the trace will not occur, unless the user explicitly allows opt-in exit spans without an entry span.
	itemResponse, err := containerClient.CreateItem(context.Background(), pk, marshalled, nil)
	if err != nil {
		log.Print("Failed to create the item:", err)
	}

	log.Printf("Status %d. Item %v created. ActivityId %s. Consuming %v Request Units.\n",
		itemResponse.RawResponse.StatusCode,
		pk, itemResponse.ActivityID,
		itemResponse.RequestCharge)
}
