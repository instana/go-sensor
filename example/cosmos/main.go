// (c) Copyright IBM Corp. 2024

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
	"github.com/google/uuid"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instacosmos"
)

// The following variables are database id and container name created in azure.
// this values needs to be changed to if the database and container are created
// with different names.
const (
	database  = "trace-data"
	container = "spans"
)

// environment variables to be exported
const (
	CONNECTION_URL = "COSMOS_CONNECTION_URL"
	KEY            = "COSMOS_KEY"
)

type SpanType string

const (
	EntrySpan SpanType = "entry"
	ExitSpan  SpanType = "exit"
)

// Span is the item to be inserted in the azure container.
type Span struct {
	ID          string   `json:"id"`
	SpanID      string   `json:"SpanID"`
	Type        SpanType `json:"type"`
	Description string   `json:"description"`
}

var (
	collector instana.TracerLogger
)

var (
	endpoint = ""
	key      = ""
)

func init() {
	validateAzureCreds()
	collector = instana.InitCollector(&instana.Options{
		Service: "sample-app-cosmos",
		Tracer:  instana.DefaultTracerOptions(),
	})
}

func main() {
	http.HandleFunc("/cosmos-test", instana.TracingHandlerFunc(collector, "/cosmos-test", handler))
	log.Fatal(http.ListenAndServe("localhost:9990", nil))
}

func handler(w http.ResponseWriter, r *http.Request) {

	erStr := r.URL.Query().Get("error")
	needError := erStr == "true"

	itemResponse, err := cosmosTest(r.Context(), needError)
	if err != nil {
		var responseErr = new(azcore.ResponseError)
		ok := errors.As(err, &responseErr)
		if !ok {
			log.Println("Error:", err.Error())
			sendErrResp(w, http.StatusInternalServerError)
			return
		}
		defer responseErr.RawResponse.Body.Close()
		errBytes, err := io.ReadAll(responseErr.RawResponse.Body)
		if err != nil {
			log.Fatal("Failed to read error body")
		}
		log.Println("Error:", string(errBytes))
		sendErrResp(w, responseErr.StatusCode)
	} else {
		sendOkResp(w)
		log.Printf("Status %d. ActivityId %s. Consuming %v Request Units.\n",
			itemResponse.RawResponse.StatusCode,
			itemResponse.ActivityID,
			itemResponse.RequestCharge)
	}
}

func sendErrResp(w http.ResponseWriter, statusCode int) {
	w.WriteHeader(statusCode)
	_, _ = w.Write([]byte(`{message:"something went wrong"}`))
}

func sendOkResp(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("{message:Status OK! Check terminal for full log!}"))
}

func cosmosTest(ctx context.Context, needError bool) (azcosmos.ItemResponse, error) {

	// Create a CosmosDB client
	client, err := getClient(collector)
	if err != nil {
		log.Fatal("Failed to create Azure Cosmos DB client: ", err)
	}

	// Create container client
	containerClient, err := client.NewContainer(database, container)
	if err != nil {
		log.Fatal("Failed to create a container client:", err)
	}

	id := uuid.New().String()
	partitionKey := fmt.Sprintf("span-%s", id)

	// Specifies the value of the partition key
	pk := containerClient.NewPartitionKeyString(partitionKey)

	if needError {
		partitionKey = "invalidPartitionKey"
	}

	span := Span{
		ID:          id,
		SpanID:      partitionKey,
		Type:        EntrySpan,
		Description: "sample span",
	}

	b, err := json.Marshal(span)
	if err != nil {
		log.Fatal("Failed to marshal span data:", err)
	}

	// setting item options upon creating ie. consistency level
	itemOptions := azcosmos.ItemOptions{
		ConsistencyLevel: azcosmos.ConsistencyLevelSession.ToPtr(),
	}

	itemResponse, err := containerClient.CreateItem(ctx, pk, b, &itemOptions)
	return itemResponse, err
}

func getClient(sensor instana.TracerLogger) (instacosmos.Client, error) {
	cred, err := instacosmos.NewKeyCredential(key)
	if err != nil {
		return nil, err
	}

	client, err := instacosmos.NewClientWithKey(sensor, endpoint, cred, &azcosmos.ClientOptions{})
	if err != nil {
		return nil, err
	}

	return client, nil
}

func validateAzureCreds() {
	endpoint, _ = os.LookupEnv(CONNECTION_URL)
	key, _ = os.LookupEnv(KEY)

	if endpoint == "" || key == "" {
		log.Fatal("Azure credentials are not provided")
	}
}
