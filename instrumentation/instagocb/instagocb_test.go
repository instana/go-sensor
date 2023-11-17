// (c) Copyright IBM Corp. 2023

//go:build integration
// +build integration

package instagocb_test

import (
	"context"
	"testing"

	"github.com/couchbase/gocb/v2"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"
	"github.com/instana/go-sensor/instrumentation/instagocb"
	"github.com/stretchr/testify/assert"
)

var connStr = "couchbase://localhost"
var username = "Administrator"
var password = "password"
var testBucketName = "test-bucket"
var testScope = "test-scope"
var testCollection = "test-collection"
var testDocumentID string = "test-doc-id"

// Insert Document
type myDoc struct {
	Foo string `json:"foo"`
	Bar string `json:"bar"`
}

type alwaysReadyClient struct{}

func (alwaysReadyClient) Ready() bool                                       { return true }
func (alwaysReadyClient) SendMetrics(data acceptor.Metrics) error           { return nil }
func (alwaysReadyClient) SendEvent(event *instana.EventData) error          { return nil }
func (alwaysReadyClient) SendSpans(spans []instana.Span) error              { return nil }
func (alwaysReadyClient) SendProfiles(profiles []autoprofile.Profile) error { return nil }
func (alwaysReadyClient) Flush(context.Context) error                       { return nil }

func prepare(t *testing.T) (*instana.Recorder, context.Context, instagocb.Cluster, *assert.Assertions) {
	a := assert.New(t)
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	sensor := instana.NewSensorWithTracer(tracer)

	ctx := context.Background()
	conn, err := instagocb.Connect(sensor, connStr, gocb.ClusterOptions{
		Authenticator: gocb.PasswordAuthenticator{
			Username: username,
			Password: password,
		},
	})

	a.NoError(err)

	// clearing existing bucket
	bucketMgr := conn.Buckets()
	_ = bucketMgr.DropBucket(testBucketName, &gocb.DropBucketOptions{})

	return recorder, ctx, conn, a

}

func prepareWithBucket(t *testing.T) (*instana.Recorder, context.Context, instagocb.Cluster, *assert.Assertions) {
	recorder, ctx, cluster, a := prepare(t)

	bs := gocb.BucketSettings{
		Name:                 testBucketName,
		FlushEnabled:         true,
		ReplicaIndexDisabled: true,
		RAMQuotaMB:           150,
		NumReplicas:          1,
		BucketType:           gocb.CouchbaseBucketType,
	}
	bucketMgr := cluster.Buckets()
	err := bucketMgr.CreateBucket(gocb.CreateBucketSettings{
		BucketSettings:         bs,
		ConflictResolutionType: gocb.ConflictResolutionTypeSequenceNumber,
	}, &gocb.CreateBucketOptions{
		Context: ctx,
	})
	a.NoError(err)

	return recorder, ctx, cluster, a
}

func prepareWithScope(t *testing.T) (*instana.Recorder, context.Context, instagocb.Cluster, *assert.Assertions) {
	recorder, ctx, cluster, a := prepareWithBucket(t)
	b := cluster.Bucket(testBucketName)
	collections := b.Collections()
	err := collections.CreateScope(testScope, &gocb.CreateScopeOptions{})
	a.NoError(err)

	return recorder, ctx, cluster, a
}

func prepareWithCollection(t *testing.T) (*instana.Recorder, context.Context, instagocb.Cluster, *assert.Assertions) {
	recorder, ctx, cluster, a := prepareWithScope(t)
	b := cluster.Bucket(testBucketName)
	collections := b.Collections()
	err := collections.CreateCollection(gocb.CollectionSpec{
		Name:      testCollection,
		ScopeName: testScope,
	}, &gocb.CreateCollectionOptions{})
	a.NoError(err)

	return recorder, ctx, cluster, a
}

func prepareWithATestDocumentInCollection(t *testing.T, operation string) (*instana.Recorder, context.Context, instagocb.Cluster, *assert.Assertions, interface{}) {
	recorder, ctx, cluster, a := prepareWithCollection(t)
	collection := cluster.Bucket(testBucketName).Scope(testScope).Collection(testCollection)
	var err error
	var value interface{}

	switch operation {
	case "ds_list", "ds_set", "ds_queue":
		value = []interface{}{getTestStringValue()}
	case "ds_map":
		value = map[string]interface{}{
			"test-key": getTestStringValue(),
		}
	default:
		value = getTestStringValue()

	}
	_, err = collection.Insert(testDocumentID, value, &gocb.InsertOptions{})
	a.NoError(err)
	return recorder, ctx, cluster, a, value
}

func getLatestSpan(recorder *instana.Recorder) instana.Span {
	spans := recorder.GetQueuedSpans()
	span := spans[len(spans)-1]

	return span
}

func getTestDocumentValue() myDoc {
	return myDoc{
		Foo: "test-foo",
		Bar: "test-bar",
	}
}

func getTestStringValue() string {
	return "test-value-string"
}
