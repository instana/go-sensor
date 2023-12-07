// (c) Copyright IBM Corp. 2023

package instacosmos_test

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
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
	database  = "trace-data"
	container = "spans"
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
	endpoint string = ""
	key             = ""
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

func validateAzureCreds(t *testing.T) {
	endpoint, _ = os.LookupEnv(CONNECTION_URL)
	key, _ = os.LookupEnv(KEY)

	if endpoint == "" || key == "" {
		t.Error("Azure credentials are not provided")
		t.FailNow()
	}
}

func TestInstaContainerClient_CreateItem(t *testing.T) {

	validateAzureCreds(t)
	ctx, recorder, cc, a := prepareContainerClient(t)

	id := genRandomID()
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
		Database:      database,
		Type:          "Query",
		ReturnCode:    fmt.Sprintf("%d", resp.RawResponse.StatusCode),
		Error:         "",
	}, spData.Tags)
}

func genRandomID() string {
	size := 4
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	rand.New(rand.NewSource(time.Now().UnixNano()))

	b := make([]rune, size)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func prepare(t *testing.T) (context.Context, *instana.Recorder, instacosmos.Client, *instana.Sensor, *assert.Assertions) {
	a := assert.New(t)
	recorder := instana.NewRecorder()

	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	sensor := instana.NewSensorWithTracer(tracer)

	ctx := context.Background()

	cred, err := instacosmos.NewKeyCredential(key)
	a.NoError(err)

	client, err := instacosmos.NewClientWithKey(endpoint, cred, &azcosmos.ClientOptions{})
	a.NoError(err)

	return ctx, recorder, client, sensor, a

}

func prepareContainerClient(t *testing.T) (context.Context, *instana.Recorder, instacosmos.ContainerClient, *assert.Assertions) {
	ctx, rec, client, sensor, a := prepare(t)
	containerClient, err := client.NewContainer(sensor, database, container)
	a.NoError(err)
	return ctx, rec, containerClient, a
}

func getLatestSpan(recorder *instana.Recorder) instana.Span {
	spans := recorder.GetQueuedSpans()
	span := spans[len(spans)-1]
	return span
}
