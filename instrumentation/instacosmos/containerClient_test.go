// (c) Copyright IBM Corp. 2023

//go:build integration
// +build integration

package instacosmos_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
	"github.com/google/uuid"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"
	"github.com/instana/go-sensor/instrumentation/instacosmos"
	"github.com/stretchr/testify/assert"
)

const (
	CONNECTION_URL = "COSMOS_CONNECTION_URL"
	KEY            = "COSMOS_KEY"
)

const (
	partitionKeyPath = "/SpanID"
	databasePrefix   = "test-db-"
	container        = "spans"
)

var (
	syncInstaClient   sync.Once
	syncInstaRecorder sync.Once
	client            instacosmos.Client
	rec               *instana.Recorder
)

type alwaysReadyClient struct{}

func (alwaysReadyClient) Ready() bool                                       { return true }
func (alwaysReadyClient) SendMetrics(data acceptor.Metrics) error           { return nil }
func (alwaysReadyClient) SendEvent(event *instana.EventData) error          { return nil }
func (alwaysReadyClient) SendSpans(spans []instana.Span) error              { return nil }
func (alwaysReadyClient) SendProfiles(profiles []autoprofile.Profile) error { return nil }
func (alwaysReadyClient) Flush(context.Context) error                       { return nil }

type SpanType string

var (
	endpoint   string = ""
	key               = ""
	databaseID        = ""
)

const (
	EntrySpan SpanType = "entry"
	ExitSpan  SpanType = "exit"
)

type Span struct {
	ID          string   `json:"id"`
	SpanID      string   `json:"SpanID"`
	Type        SpanType `json:"type"`
	Description string   `json:"description"`
}

func setup(collector instana.TracerLogger) {

	// validating azure creds are exported
	validateAzureCreds()

	// getting azure cosmos client
	client, err := getInstaClient()
	failOnError(err)

	// creating a database in azure test account
	databaseID = databasePrefix + uuid.New().String()
	dbProperties := azcosmos.DatabaseProperties{ID: databaseID}
	response, err := client.CreateDatabase(context.TODO(), dbProperties, nil)
	failOnError(err)

	if response.RawResponse.StatusCode != http.StatusCreated {
		err = fmt.Errorf("Failed to create database. Got response status %d",
			response.RawResponse.StatusCode)
		failOnError(err)
	}

	dbClient, err := client.NewDatabase(collector, databaseID)
	failOnError(err)

	// create a container in test database
	properties := azcosmos.ContainerProperties{
		ID: container,
		PartitionKeyDefinition: azcosmos.PartitionKeyDefinition{
			Paths: []string{partitionKeyPath},
		},
	}

	throughput := azcosmos.NewManualThroughputProperties(400)

	resp, err := dbClient.CreateContainer(context.TODO(), properties,
		&azcosmos.CreateContainerOptions{ThroughputProperties: &throughput})
	failOnError(err)

	if resp.RawResponse.StatusCode != http.StatusCreated {
		err = fmt.Errorf("Failed to create container. Got response status %d",
			resp.RawResponse.StatusCode)
		failOnError(err)
	}
}

func shutdown(collector instana.TracerLogger) {
	client, err := getInstaClient()
	failOnError(err)

	database, err := client.NewDatabase(collector, databaseID)
	failOnError(err)

	response, err := database.Delete(context.TODO(), &azcosmos.DeleteDatabaseOptions{})
	failOnError(err)

	if response.RawResponse.StatusCode != http.StatusNoContent {
		err = fmt.Errorf("Failed to delete database. Got response status %d",
			response.RawResponse.StatusCode)
		failOnError(err)
	}
}

func TestMain(m *testing.M) {

	// creating a sensor with instana recorder
	recorder := getInstaRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	sensor := instana.NewSensorWithTracer(tracer)

	// create a database and a container in azure test account
	setup(sensor)
	// run the tests
	code := m.Run()
	// delete the test database
	shutdown(sensor)
	os.Exit(code)
}

func TestInstaContainerClient_CreateItem(t *testing.T) {

	ctx, recorder, cc, a := prepareContainerClient(t)

	id := uuid.New().String()
	spanID := fmt.Sprintf("span-%s", id)

	data := Span{
		ID:          id,
		SpanID:      spanID,
		Type:        EntrySpan,
		Description: "sample-description",
	}

	jsonData, err := json.Marshal(data)
	a.NoError(err)

	pk := azcosmos.NewPartitionKeyString(spanID)

	resp, err := cc.CreateItem(ctx, pk, jsonData, &azcosmos.ItemOptions{})
	a.NoError(err)
	a.NotEmpty(resp)

	span := getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CosmosSpanData{}, span.Data)
	spData := span.Data.(instana.CosmosSpanData)
	a.Equal(instana.CosmosSpanTags{
		ConnectionURL: endpoint,
		Database:      databaseID,
		Type:          "Query",
		ReturnCode:    fmt.Sprintf("%d", resp.RawResponse.StatusCode),
		Error:         "",
	}, spData.Tags)
}

func validateAzureCreds() {
	endpoint, _ = os.LookupEnv(CONNECTION_URL)
	key, _ = os.LookupEnv(KEY)

	if endpoint == "" || key == "" {
		failOnError(fmt.Errorf("Azure credentials are not provided"))
	}
}

func failOnError(err error) {
	if err != nil {
		log.Fatal(err)
		os.Exit(0)
	}
}

func prepare(t *testing.T) (context.Context, *instana.Recorder, instacosmos.Client, *instana.Sensor, *assert.Assertions) {
	a := assert.New(t)
	rec = getInstaRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, rec)
	sensor := instana.NewSensorWithTracer(tracer)

	ctx := context.Background()

	client, err := getInstaClient()
	a.NoError(err)

	return ctx, rec, client, sensor, a

}

func getInstaClient() (instacosmos.Client, error) {
	var err error
	syncInstaClient.Do(func() {
		cred, e := instacosmos.NewKeyCredential(key)
		if e != nil {
			err = e
		}

		client, e = instacosmos.NewClientWithKey(endpoint, cred, &azcosmos.ClientOptions{})
		if e != nil {
			err = e
		}
	})

	return client, err
}

func getInstaRecorder() *instana.Recorder {
	syncInstaRecorder.Do(func() {
		rec = instana.NewRecorder()
	})
	return rec
}

func prepareContainerClient(t *testing.T) (context.Context, *instana.Recorder, instacosmos.ContainerClient, *assert.Assertions) {
	ctx, rec, client, sensor, a := prepare(t)
	containerClient, err := client.NewContainer(sensor, databaseID, container)
	a.NoError(err)
	return ctx, rec, containerClient, a
}

func getLatestSpan(recorder *instana.Recorder) instana.Span {
	spans := recorder.GetQueuedSpans()
	span := spans[len(spans)-1]
	return span
}
