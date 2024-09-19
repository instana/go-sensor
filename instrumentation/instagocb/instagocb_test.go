// (c) Copyright IBM Corp. 2023

//go:build integration
// +build integration

package instagocb_test

import (
	"context"
	"math/rand"
	"os"
	"testing"

	"github.com/couchbase/gocb/v2"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"
	"github.com/instana/go-sensor/instrumentation/instagocb"
	"github.com/stretchr/testify/assert"
)

var testID = genUniqueTestID()
var connStr = "couchbase://localhost"
var username = "Administrator"
var password = "password"

// cb resource names
var (
	cbTestBucket     = "cb-test-bucket_" + testID
	cbTestScope      = "cb_test_scope_" + testID
	cbTestCollection = "cb_test_collection_" + testID
	testBucketName   = "test-bucket_" + testID
	testScope        = "test_scope_" + testID
	testCollection   = "test_collection_" + testID
)

// test document IDs
var (
	testDocumentID            = "test-doc-id"
	clusterTestDocumentID     = "cluster-test-doc-id"
	dsListTestDocumentID      = "ds-list-test-doc-id"
	dsMapTestDocumentID       = "ds-map-test-doc-id"
	dsQueueTestDocumentID     = "ds-queue-test-doc-id"
	dsSetTestDocumentID       = "ds-set-test-doc-id"
	scopeTestDocumentID       = "scope-test-doc-id"
	crudTestDocumentID        = "crud-test-doc-id"
	transactionTestDocumentID = "transaction-test-doc-id"
)

var rec *instana.Recorder

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
	_, _, cluster, a := prepare(t)

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

	// Transactions
	ts := cluster.Transactions().Unwrap()
	a.IsType(&gocb.Transactions{}, ts)
	a.NotNil(ts)

	// TransactionAttemptContext
	tac := cluster.WrapTransactionAttemptContext(&gocb.TransactionAttemptContext{}, nil).Unwrap()
	a.IsType(&gocb.TransactionAttemptContext{}, tac)
	a.NotNil(tac)

}

// helper functions

func prepare(t *testing.T) (*instana.Recorder, context.Context, instagocb.Cluster, *assert.Assertions) {
	a := assert.New(t)
	var recorder *instana.Recorder
	if rec == nil {
		rec = instana.NewTestRecorder()
	}

	recorder = rec
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	sensor := instana.NewSensorWithTracer(tracer)

	pSpan := sensor.Tracer().StartSpan("parent-span")
	ctx := context.Background()
	if pSpan != nil {
		ctx = instana.ContextWithSpan(ctx, pSpan)
	}

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

func prepareWithBucket(t *testing.T, bucketName string) (*instana.Recorder, context.Context, instagocb.Cluster, *assert.Assertions) {
	recorder, ctx, cluster, a := prepare(t)

	bs := gocb.BucketSettings{
		Name:                 bucketName,
		FlushEnabled:         true,
		ReplicaIndexDisabled: true,
		RAMQuotaMB:           100,
		NumReplicas:          0,
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

func prepareWithScope(t *testing.T, bucketName, scope string) (*instana.Recorder, context.Context, instagocb.Cluster, *assert.Assertions) {
	recorder, ctx, cluster, a := prepareWithBucket(t, bucketName)
	b := cluster.Bucket(bucketName)
	collections := b.Collections()
	err := collections.CreateScope(scope, &gocb.CreateScopeOptions{})
	a.NoError(err)

	return recorder, ctx, cluster, a
}

func prepareWithCollection(t *testing.T, bucketName, scope, collection string) (*instana.Recorder, context.Context, instagocb.Cluster, *assert.Assertions) {
	recorder, ctx, cluster, a := prepareWithScope(t, bucketName, scope)
	b := cluster.Bucket(bucketName)
	collections := b.Collections()
	err := collections.CreateCollection(gocb.CollectionSpec{
		Name:      collection,
		ScopeName: scope,
	}, &gocb.CreateCollectionOptions{})
	a.NoError(err)

	return recorder, ctx, cluster, a
}

func setup() error {

	t := &testing.T{}

	_, _, cluster, a := prepareWithCollection(t, cbTestBucket, cbTestScope, cbTestCollection)
	b := cluster.Bucket(cbTestBucket)
	collection := b.Scope(cbTestScope).Collection(cbTestCollection)

	docs := map[string]interface{}{
		dsListTestDocumentID:  []interface{}{getTestStringValue()},
		dsSetTestDocumentID:   []interface{}{getTestStringValue()},
		dsQueueTestDocumentID: []interface{}{getTestStringValue()},
		dsMapTestDocumentID: map[string]interface{}{
			"test-key": getTestStringValue(),
		},
		scopeTestDocumentID:       getTestDocumentValue(),
		clusterTestDocumentID:     getTestDocumentValue(),
		crudTestDocumentID:        getTestDocumentValue(),
		transactionTestDocumentID: getTestDocumentValue(),
	}

	// Create an Array of BulkOps for Insert
	var items []gocb.BulkOp

	// Add docs to the array that will be performed as a bulk operation
	for key, val := range docs {
		items = append(items, &gocb.InsertOp{ID: key, Value: val})
	}

	// Perform the bulk operation
	err := collection.Do(items, &gocb.BulkOpOptions{})
	a.NoError(err)

	_, err = b.DefaultCollection().Insert(testDocumentID, getTestDocumentValue(), &gocb.InsertOptions{})
	a.NoError(err)

	return err
}

func TestMain(m *testing.M) {
	if err := setup(); err != nil {
		os.Exit(1)
	}

	exitCode := m.Run()

	clear()
	os.Exit(exitCode)

}

func clear() {

	_, _, cluster, _ := prepare(&testing.T{})
	bucketMgr := cluster.Buckets()

	// drop testBucket
	bucketMgr.DropBucket(testBucketName, &gocb.DropBucketOptions{})

	// drop cbTestBucket
	bucketMgr.DropBucket(cbTestBucket, &gocb.DropBucketOptions{})
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

func genUniqueTestID() string {

	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, 10)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}
