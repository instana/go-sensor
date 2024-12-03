// (c) Copyright IBM Corp. 2024

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
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
	"github.com/google/uuid"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"
	"github.com/instana/go-sensor/instrumentation/instacosmos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// data operation types
const (
	Query   = "Query"
	Write   = "Write"
	Execute = "Execute"
	Update  = "Update"
	Upsert  = "Upsert"
	Replace = "Replace"
)

// test items IDs
const (
	ID1 = "A001" // item to be read
	ID2 = "A002" // item to be replaced
	ID3 = "A003" // item to be patched
	ID4 = "A004" // item to be deleted
	ID5 = "A005" // item to be upsert
	ID6 = "A006" // item to be insert as part of transaction
	ID7 = "A007" // item to be delete as part of transaction
)

var (
	syncInstaClient   sync.Once
	syncInstaRecorder sync.Once
	client            instacosmos.Client
	rec               *instana.Recorder
)

// exit error code
var exitNotOk int = 1

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
	client, err := getInstaClient(collector)
	if err != nil {
		log.Fatalf("instacosmos integration test failed : %s \n", err.Error())
	}

	// creating a database in azure test account
	databaseID = databasePrefix + uuid.New().String()
	dbProperties := azcosmos.DatabaseProperties{ID: databaseID}
	response, err := client.CreateDatabase(context.TODO(), dbProperties, nil)
	if err != nil {
		log.Fatalf("instacosmos integration test failed : %s \n", err.Error())
	}

	if response.RawResponse.StatusCode != http.StatusCreated {
		err = fmt.Errorf("Failed to create database. Got response status %d",
			response.RawResponse.StatusCode)
		log.Fatalf("instacosmos integration test failed : %s \n", err.Error())
	}

	dbClient, err := client.NewDatabase(databaseID)
	failOnErrorAndTearDown(collector, err)

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
	failOnErrorAndTearDown(collector, err)

	if resp.RawResponse.StatusCode != http.StatusCreated {
		err = fmt.Errorf("Failed to create container. Got response status %d",
			resp.RawResponse.StatusCode)
		failOnErrorAndTearDown(collector, err)
	}

	containerClient, err := client.NewContainer(databaseID, container)
	failOnErrorAndTearDown(collector, err)

	err = prepareTestData(containerClient)
	failOnErrorAndTearDown(collector, err)
}

func gracefulShutdown(collector instana.TracerLogger, code int) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdown(collector, code)
}

func shutdown(collector instana.TracerLogger, exitCode int) {
	client, err := getInstaClient(collector)
	if err != nil {
		log.Fatalln(err)
	}

	database, err := client.NewDatabase(databaseID)
	if err != nil {
		log.Fatalln(err)
	}

	response, err := database.Delete(context.TODO(), &azcosmos.DeleteDatabaseOptions{})
	if err != nil {
		log.Fatalln(err)
	}

	if response.RawResponse.StatusCode != http.StatusNoContent {
		err = fmt.Errorf("Failed to delete database. Got response status %d",
			response.RawResponse.StatusCode)
		log.Fatalln(err)
	}

	os.Exit(exitCode)
}

func TestMain(m *testing.M) {

	// creating a sensor with instana recorder
	recorder := getInstaRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	sensor := instana.NewSensorWithTracer(tracer)

	// handles panic errors
	// this will only work for the panic errors before m.Run()
	// check this issue for more details: https://github.com/golang/go/issues/37206
	defer func() {
		if err := recover(); err != nil {
			shutdown(sensor, exitNotOk)
		}
	}()

	// handles interrupt from user
	go gracefulShutdown(sensor, exitNotOk)

	// create a database and a container in azure test account
	setup(sensor)

	// flush all the created spans while test data creation
	recorder.Flush(context.TODO())

	// run the tests
	code := m.Run()

	shutdown(sensor, code)
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

	pk := cc.NewPartitionKeyString(spanID)

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
		Database:      databaseID + ":" + container,
		Type:          Write,
		Sql:           "INSERT INTO " + container,
		Object:        container,
		PartitionKey:  spanID,
		ReturnCode:    fmt.Sprintf("%d", resp.RawResponse.StatusCode),
		Error:         "",
	}, spData.Tags)
}

func TestInstaContainerClient_CreateItem_WithError(t *testing.T) {

	ctx, recorder, cc, a := prepareContainerClient(t)

	id := uuid.New().String()
	spanID := fmt.Sprintf("span-%s", id)

	data := Span{
		ID:          id,
		SpanID:      "invalidPartitionKey",
		Type:        EntrySpan,
		Description: "sample-description",
	}

	jsonData, err := json.Marshal(data)
	a.NoError(err)

	pk := cc.NewPartitionKeyString(spanID)

	_, err = cc.CreateItem(ctx, pk, jsonData, &azcosmos.ItemOptions{})
	a.Error(err)

	time.Sleep(2 * time.Second)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 2)

	span, logSpan := spans[0], spans[1]
	assert.Equal(t, 1, span.Ec)

	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CosmosSpanData{}, span.Data)

	spanData := span.Data.(instana.CosmosSpanData)
	a.Equal(instana.CosmosSpanTags{
		ConnectionURL: endpoint,
		Database:      databaseID + ":" + container,
		Type:          Write,
		Sql:           "INSERT INTO " + container,
		Object:        container,
		PartitionKey:  spanID,
		ReturnCode:    fmt.Sprintf("%d", 400),
		Error:         err.Error(),
	}, spanData.Tags)

	assert.Equal(t, span.TraceID, logSpan.TraceID)
	assert.Equal(t, span.SpanID, logSpan.ParentID)
	assert.Equal(t, "log.go", logSpan.Name)
}

func TestInstaContainerClient_DeleteItem(t *testing.T) {

	ctx, recorder, cc, a := prepareContainerClient(t)

	spanID := fmt.Sprintf("span-%s", ID4)
	pk := cc.NewPartitionKeyString(spanID)
	resp, err := cc.DeleteItem(ctx, pk, ID4, &azcosmos.ItemOptions{})
	a.NoError(err)
	a.NotEmpty(resp)

	span := getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CosmosSpanData{}, span.Data)
	spData := span.Data.(instana.CosmosSpanData)
	a.Equal(instana.CosmosSpanTags{
		ConnectionURL: endpoint,
		Database:      databaseID + ":" + container,
		Type:          Write,
		Sql:           "DELETE FROM " + container,
		Object:        container,
		PartitionKey:  spanID,
		ReturnCode:    fmt.Sprintf("%d", resp.RawResponse.StatusCode),
		Error:         "",
	}, spData.Tags)
}

func TestInstaContainerClient_NewQueryItemsPager(t *testing.T) {

	os.Setenv("INSTANA_ALLOW_ROOT_EXIT_SPAN", "1")
	defer os.Unsetenv("INSTANA_ALLOW_ROOT_EXIT_SPAN")

	_, recorder, cc, a := prepareContainerClient(t)

	spanID := fmt.Sprintf("span-%s", ID1)
	pk := cc.NewPartitionKeyString(spanID)

	query := fmt.Sprintf("SELECT * FROM %v", container)
	resp := cc.NewQueryItemsPager(query, pk, &azcosmos.QueryOptions{})
	a.NotEmpty(resp)

	span := getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CosmosSpanData{}, span.Data)
	spData := span.Data.(instana.CosmosSpanData)
	a.Equal(instana.CosmosSpanTags{
		ConnectionURL: endpoint,
		Database:      databaseID + ":" + container,
		Type:          Query,
		Sql:           query,
		Object:        container,
		PartitionKey:  spanID,
		ReturnCode:    "",
		Error:         "",
	}, spData.Tags)
}

func TestInstaContainerClient_PatchItem(t *testing.T) {

	ctx, recorder, cc, a := prepareContainerClient(t)

	spanID := fmt.Sprintf("span-%s", ID3)
	pk := cc.NewPartitionKeyString(spanID)

	patch := azcosmos.PatchOperations{}

	patch.AppendAdd("/updatedTime", time.Now().Unix())
	patch.AppendRemove("/description")

	resp, err := cc.PatchItem(ctx, pk, ID3, patch, &azcosmos.ItemOptions{})
	a.NoError(err)
	a.NotEmpty(resp)

	span := getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CosmosSpanData{}, span.Data)
	spData := span.Data.(instana.CosmosSpanData)
	a.Equal(instana.CosmosSpanTags{
		ConnectionURL: endpoint,
		Database:      databaseID + ":" + container,
		Type:          Update,
		Sql:           "UPDATE " + container,
		Object:        container,
		PartitionKey:  spanID,
		ReturnCode:    fmt.Sprintf("%d", resp.RawResponse.StatusCode),
		Error:         "",
	}, spData.Tags)
}

func TestInstaContainerClient_ExecuteTransactionalBatch(t *testing.T) {

	ctx, recorder, cc, a := prepareContainerClient(t)

	spanID := fmt.Sprintf("span-%s", ID6)
	pk := cc.NewPartitionKeyString(spanID)

	batch := cc.NewTransactionalBatch(pk)
	data := Span{
		ID:          ID6,
		SpanID:      spanID,
		Type:        EntrySpan,
		Description: "sample-description",
	}

	jsonData, err := json.Marshal(data)
	a.NoError(err)

	batch.CreateItem(jsonData, nil)
	batch.ReadItem(ID1, nil)
	batch.DeleteItem(ID7, nil)

	resp, err := cc.ExecuteTransactionalBatch(ctx, batch, &azcosmos.TransactionalBatchOptions{})
	a.NoError(err)
	a.NotEmpty(resp)

	span := getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CosmosSpanData{}, span.Data)
	spData := span.Data.(instana.CosmosSpanData)
	a.Equal(instana.CosmosSpanTags{
		ConnectionURL: endpoint,
		Database:      databaseID + ":" + container,
		Type:          Execute,
		Sql:           "Execute",
		Object:        container,
		PartitionKey:  spanID,
		ReturnCode:    fmt.Sprintf("%d", resp.RawResponse.StatusCode),
		Error:         "",
	}, spData.Tags)
}

func TestInstaContainerClient_Read(t *testing.T) {

	ctx, recorder, cc, a := prepareContainerClient(t)

	resp, err := cc.Read(ctx, &azcosmos.ReadContainerOptions{})
	a.NoError(err)
	a.NotEmpty(resp)

	span := getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CosmosSpanData{}, span.Data)
	spData := span.Data.(instana.CosmosSpanData)
	a.Equal(instana.CosmosSpanTags{
		ConnectionURL: endpoint,
		Database:      databaseID + ":" + container,
		Type:          Query,
		Sql:           "Read",
		Object:        container,
		PartitionKey:  "",
		ReturnCode:    fmt.Sprintf("%d", resp.RawResponse.StatusCode),
		Error:         "",
	}, spData.Tags)
}

func TestInstaContainerClient_ReadItem(t *testing.T) {

	ctx, recorder, cc, a := prepareContainerClient(t)
	spanID := fmt.Sprintf("span-%s", ID1)
	pk := cc.NewPartitionKeyString(spanID)

	resp, err := cc.ReadItem(ctx, pk, ID1, &azcosmos.ItemOptions{})
	a.NoError(err)
	a.NotEmpty(resp)

	span := getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CosmosSpanData{}, span.Data)
	spData := span.Data.(instana.CosmosSpanData)
	a.Equal(instana.CosmosSpanTags{
		ConnectionURL: endpoint,
		Database:      databaseID + ":" + container,
		Type:          Query,
		Sql:           "SELECT " + ID1 + " FROM " + container,
		Object:        container,
		PartitionKey:  spanID,
		ReturnCode:    fmt.Sprintf("%d", resp.RawResponse.StatusCode),
		Error:         "",
	}, spData.Tags)
}

func TestInstaContainerClient_ReadThroughput(t *testing.T) {

	ctx, recorder, cc, a := prepareContainerClient(t)

	resp, err := cc.ReadThroughput(ctx, &azcosmos.ThroughputOptions{})
	a.NoError(err)
	a.NotEmpty(resp)

	span := getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CosmosSpanData{}, span.Data)
	spData := span.Data.(instana.CosmosSpanData)
	a.Equal(instana.CosmosSpanTags{
		ConnectionURL: endpoint,
		Database:      databaseID + ":" + container,
		Type:          Query,
		Sql:           "ReadThroughput",
		Object:        container,
		PartitionKey:  "",
		ReturnCode:    fmt.Sprintf("%d", resp.RawResponse.StatusCode),
		Error:         "",
	}, spData.Tags)
}

func TestInstaContainerClient_Replace(t *testing.T) {

	ctx, recorder, cc, a := prepareContainerClient(t)

	containerResponse, err := cc.Read(context.Background(), nil)
	a.NoError(err)
	a.NotEmpty(containerResponse)

	// Changing the indexing policy
	containerResponse.ContainerProperties.IndexingPolicy = &azcosmos.IndexingPolicy{
		IncludedPaths: []azcosmos.IncludedPath{},
		ExcludedPaths: []azcosmos.ExcludedPath{},
		Automatic:     false,
		IndexingMode:  azcosmos.IndexingModeNone,
	}

	resp, err := cc.Replace(ctx, *containerResponse.ContainerProperties, &azcosmos.ReplaceContainerOptions{})
	a.NoError(err)
	a.NotEmpty(resp)

	span := getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CosmosSpanData{}, span.Data)
	spData := span.Data.(instana.CosmosSpanData)
	a.Equal(instana.CosmosSpanTags{
		ConnectionURL: endpoint,
		Database:      databaseID + ":" + container,
		Type:          Replace,
		Sql:           "Replace",
		Object:        container,
		PartitionKey:  "",
		ReturnCode:    fmt.Sprintf("%d", resp.RawResponse.StatusCode),
		Error:         "",
	}, spData.Tags)
}

func TestInstaContainerClient_ReplaceItem(t *testing.T) {

	ctx, recorder, cc, a := prepareContainerClient(t)

	spanID := fmt.Sprintf("span-%s", ID2)
	pk := cc.NewPartitionKeyString(spanID)

	data := Span{
		ID:          ID2,
		SpanID:      spanID,
		Type:        ExitSpan,
		Description: "updated-description",
	}

	jsonData, err := json.Marshal(data)
	a.NoError(err)

	resp, err := cc.ReplaceItem(ctx, pk, ID2, jsonData, &azcosmos.ItemOptions{})
	a.NoError(err)
	a.NotEmpty(resp)

	span := getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CosmosSpanData{}, span.Data)
	spData := span.Data.(instana.CosmosSpanData)
	a.Equal(instana.CosmosSpanTags{
		ConnectionURL: endpoint,
		Database:      databaseID + ":" + container,
		Type:          Replace,
		Sql:           "ReplaceItem",
		Object:        container,
		PartitionKey:  spanID,
		ReturnCode:    fmt.Sprintf("%d", resp.RawResponse.StatusCode),
		Error:         "",
	}, spData.Tags)
}

func TestInstaContainerClient_ReplaceThroughput(t *testing.T) {

	ctx, recorder, cc, a := prepareContainerClient(t)

	throughputResponse, err := cc.ReadThroughput(context.Background(), nil)
	a.NoError(err)

	_, hasManual := throughputResponse.ThroughputProperties.ManualThroughput()
	a.True(hasManual)

	// Replace manual throughput
	newScale := azcosmos.NewManualThroughputProperties(500)

	resp, err := cc.ReplaceThroughput(ctx, newScale, &azcosmos.ThroughputOptions{})
	a.NoError(err)
	a.NotEmpty(resp)

	span := getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CosmosSpanData{}, span.Data)
	spData := span.Data.(instana.CosmosSpanData)
	a.Equal(instana.CosmosSpanTags{
		ConnectionURL: endpoint,
		Database:      databaseID + ":" + container,
		Type:          Replace,
		Sql:           "ReplaceThroughput",
		Object:        container,
		PartitionKey:  "",
		ReturnCode:    fmt.Sprintf("%d", resp.RawResponse.StatusCode),
		Error:         "",
	}, spData.Tags)
}

func TestInstaContainerClient_UpsertItem(t *testing.T) {

	ctx, recorder, cc, a := prepareContainerClient(t)

	spanID := fmt.Sprintf("span-%s", ID5)
	pk := cc.NewPartitionKeyString(spanID)

	data := Span{
		ID:          ID2,
		SpanID:      spanID,
		Type:        ExitSpan,
		Description: "updated-description",
	}

	jsonData, err := json.Marshal(data)
	a.NoError(err)

	resp, err := cc.UpsertItem(ctx, pk, jsonData, &azcosmos.ItemOptions{})
	a.NoError(err)
	a.NotEmpty(resp)

	span := getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CosmosSpanData{}, span.Data)
	spData := span.Data.(instana.CosmosSpanData)
	a.Equal(instana.CosmosSpanTags{
		ConnectionURL: endpoint,
		Database:      databaseID + ":" + container,
		Type:          Upsert,
		Sql:           "UpsertItem",
		Object:        container,
		PartitionKey:  spanID,
		ReturnCode:    fmt.Sprintf("%d", resp.RawResponse.StatusCode),
		Error:         "",
	}, spData.Tags)
}

func validateAzureCreds() {
	endpoint, _ = os.LookupEnv(CONNECTION_URL)
	key, _ = os.LookupEnv(KEY)

	if endpoint == "" || key == "" {
		log.Fatalln("Azure credentials are not provided")
	}
}

func failOnErrorAndTearDown(collector instana.TracerLogger, err error) {
	if err != nil {
		fmt.Printf("instacosmos integration test failed : %s \n", err.Error())
		shutdown(collector, exitNotOk)
	}
}

func prepare(t *testing.T) (context.Context, *instana.Recorder, instacosmos.Client, *assert.Assertions) {
	a := assert.New(t)
	rec = getInstaRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, rec)
	sensor := instana.NewSensorWithTracer(tracer)

	pSpan := sensor.Tracer().StartSpan("parent-span")
	ctx := context.Background()
	if pSpan != nil {
		ctx = instana.ContextWithSpan(ctx, pSpan)
	}

	client, err := getInstaClient(sensor)
	a.NoError(err)

	return ctx, rec, client, a

}

func getInstaClient(collector instana.TracerLogger) (instacosmos.Client, error) {
	var err error
	syncInstaClient.Do(func() {
		cred, e := instacosmos.NewKeyCredential(key)
		if e != nil {
			err = e
		}

		client, e = instacosmos.NewClientWithKey(collector, endpoint, cred, &azcosmos.ClientOptions{})
		if e != nil {
			err = e
		}
	})

	return client, err
}

func getInstaRecorder() *instana.Recorder {
	syncInstaRecorder.Do(func() {
		rec = instana.NewTestRecorder()
	})
	return rec
}

func prepareContainerClient(t *testing.T) (context.Context, *instana.Recorder, instacosmos.ContainerClient, *assert.Assertions) {
	ctx, rec, client, a := prepare(t)
	containerClient, err := client.NewContainer(databaseID, container)
	a.NoError(err)
	return ctx, rec, containerClient, a
}

func getLatestSpan(recorder *instana.Recorder) instana.Span {
	spans := recorder.GetQueuedSpans()
	span := spans[len(spans)-1]
	return span
}

func prepareTestData(client instacosmos.ContainerClient) (err error) {
	data := []Span{
		{
			ID:          ID1,
			SpanID:      "span-" + ID1,
			Type:        EntrySpan,
			Description: "sample-description",
		},
		{
			ID:          ID2,
			SpanID:      "span-" + ID2,
			Type:        EntrySpan,
			Description: "sample-description",
		},
		{
			ID:          ID3,
			SpanID:      "span-" + ID3,
			Type:        EntrySpan,
			Description: "sample-description",
		},
		{
			ID:          ID4,
			SpanID:      "span-" + ID4,
			Type:        EntrySpan,
			Description: "sample-description",
		},
		{
			ID:          ID5,
			SpanID:      "span-" + ID5,
			Type:        EntrySpan,
			Description: "sample-description",
		},
		{
			ID:          ID7,
			SpanID:      "span-" + ID7,
			Type:        EntrySpan,
			Description: "sample-description",
		},
	}

	for _, item := range data {
		pk := client.NewPartitionKeyString(item.SpanID)
		jsonData, err := json.Marshal(item)
		if err != nil {
			return err
		}
		_, err = client.CreateItem(context.TODO(), pk, jsonData, &azcosmos.ItemOptions{})
		if err != nil {
			return err
		}
	}

	return
}

func Test_instaContainerClient_DatabaseID(t *testing.T) {
	_, _, cc, _ := prepareContainerClient(t)
	tests := []struct {
		name string
		want string
	}{
		{
			name: "success",
			want: databaseID,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if got := cc.DatabaseID(); got != tt.want {
				t.Errorf("instaContainerClient.DatabaseID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_instaContainerClient_ID(t *testing.T) {
	_, _, cc, _ := prepareContainerClient(t)
	tests := []struct {
		name string
		want string
	}{
		{
			name: "success",
			want: containerName,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cc.ID(); got != tt.want {
				t.Errorf("instaContainerClient.ID() = %v, want %v", got, tt.want)
			}
		})
	}
}
