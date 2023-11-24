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
var testScope = "test_scope"
var testCollection = "test_collection"
var testDocumentID string = "test-doc-id"

// Test Document to insert
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

func TestUnwrapForAll(t *testing.T) {
	defer instana.ShutdownSensor()
	_, _, cluster, a, _ := prepareWithATestDocumentInCollection(t, "ds_list")

	// Cluster
	c := cluster.Unwrap()
	a.IsType(&gocb.Cluster{}, c)
	a.NotNil(c)

	// Bucket Manager
	bm := cluster.Buckets().Unwrap()
	a.IsType(&gocb.BucketManager{}, bm)
	a.NotNil(bm)

	// Bucket
	b := cluster.Bucket(testBucketName).Unwrap()
	a.IsType(&gocb.Bucket{}, b)
	a.NotNil(b)

	//Scope
	s := cluster.Bucket(testBucketName).Scope(testScope)
	su := s.Unwrap()
	a.IsType(&gocb.Scope{}, su)
	a.NotNil(su)

	// Collection
	coll := s.Collection(testCollection)
	collU := coll.Unwrap()
	a.IsType(&gocb.Collection{}, collU)
	a.NotNil(collU)

	// Collection Manager
	cm := cluster.Bucket(testBucketName).Collections().Unwrap()
	a.IsType(&gocb.CollectionManager{}, cm)
	a.NotNil(cm)

	// Collection Map
	m := coll.Map("id").Unwrap()
	a.IsType(&gocb.CouchbaseMap{}, m)
	a.NotNil(m)

	// Collection List
	l := coll.List("id").Unwrap()
	a.IsType(&gocb.CouchbaseList{}, l)
	a.NotNil(l)

	// Collection Queue
	q := coll.Queue("id").Unwrap()
	a.IsType(&gocb.CouchbaseQueue{}, q)
	a.NotNil(q)

	// Collection Set
	st := coll.Set("id").Unwrap()
	a.IsType(&gocb.CouchbaseSet{}, st)
	a.NotNil(st)

	// Collection Binary
	cb := coll.Binary().Unwrap()
	a.IsType(&gocb.BinaryCollection{}, cb)
	a.NotNil(cb)

}

// helper functions

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
	b := cluster.Bucket(testBucketName)
	collection := b.Scope(testScope).Collection(testCollection)
	var err error
	var value interface{}

	switch operation {
	case "ds_list", "ds_set", "ds_queue":
		value = []interface{}{getTestStringValue()}
	case "ds_map":
		value = map[string]interface{}{
			"test-key": getTestStringValue(),
		}
	case "scope", "cluster":
		value = getTestDocumentValue()
	default:
		value = getTestDocumentValue()

	}
	_, err = collection.Insert(testDocumentID, value, &gocb.InsertOptions{})
	a.NoError(err)

	if operation == "cluster" {
		_, err = b.DefaultCollection().Insert(testDocumentID, value, &gocb.InsertOptions{})
		a.NoError(err)
	}

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
