// (c) Copyright IBM Corp. 2023

//go:build integration
// +build integration

package instagocb_test

import (
	"testing"
	"time"

	"github.com/couchbase/gocb/v2"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instagocb"
)

func TestCollection_CRUD(t *testing.T) {
	testDocumentValue := getTestDocumentValue()
	defer instana.ShutdownSensor()
	recorder, ctx, cluster, a := prepareWithCollection(t)

	collection := cluster.Bucket(testBucketName).Scope(testScope).Collection(testCollection)

	// Insert
	_, err := collection.Insert(testDocumentID, testDocumentValue, &gocb.InsertOptions{ParentSpan: instagocb.GetParentSpanFromContext(ctx)})
	a.NoError(err)

	span := getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CouchbaseSpanData{}, span.Data)
	data := span.Data.(instana.CouchbaseSpanData)
	a.Equal(instana.CouchbaseSpanTags{
		Bucket: testBucketName,
		Host:   "localhost",
		Type:   string(gocb.CouchbaseBucketType),
		SQL:    "INSERT",
		Error:  "",
	}, data.Tags)

	// Get
	var result myDoc
	res, err := collection.Get(testDocumentID, &gocb.GetOptions{ParentSpan: instagocb.GetParentSpanFromContext(ctx)})
	a.NoError(err)
	res.Content(&result)
	a.Equal(testDocumentValue, result)
	span = getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CouchbaseSpanData{}, span.Data)
	data = span.Data.(instana.CouchbaseSpanData)
	a.Equal(instana.CouchbaseSpanTags{
		Bucket: testBucketName,
		Host:   "localhost",
		Type:   string(gocb.CouchbaseBucketType),
		SQL:    "GET",
		Error:  "",
	}, data.Tags)

	// Upsert
	_, err = collection.Upsert(testDocumentID, &myDoc{}, &gocb.UpsertOptions{ParentSpan: instagocb.GetParentSpanFromContext(ctx)})
	a.NoError(err)

	span = getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CouchbaseSpanData{}, span.Data)
	data = span.Data.(instana.CouchbaseSpanData)
	a.Equal(instana.CouchbaseSpanTags{
		Bucket: testBucketName,
		Host:   "localhost",
		Type:   string(gocb.CouchbaseBucketType),
		SQL:    "UPSERT",
		Error:  "",
	}, data.Tags)

	// Replace
	_, err = collection.Replace(testDocumentID, "newValue2", &gocb.ReplaceOptions{ParentSpan: instagocb.GetParentSpanFromContext(ctx)})
	a.NoError(err)

	span = getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CouchbaseSpanData{}, span.Data)
	data = span.Data.(instana.CouchbaseSpanData)
	a.Equal(instana.CouchbaseSpanTags{
		Bucket: testBucketName,
		Host:   "localhost",
		Type:   string(gocb.CouchbaseBucketType),
		SQL:    "REPLACE",
		Error:  "",
	}, data.Tags)

	// Exists
	res1, err := collection.Exists(testDocumentID, &gocb.ExistsOptions{ParentSpan: instagocb.GetParentSpanFromContext(ctx)})
	a.NoError(err)
	a.True(res1.Exists())

	span = getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CouchbaseSpanData{}, span.Data)
	data = span.Data.(instana.CouchbaseSpanData)
	a.Equal(instana.CouchbaseSpanTags{
		Bucket: testBucketName,
		Host:   "localhost",
		Type:   string(gocb.CouchbaseBucketType),
		SQL:    "EXISTS",
		Error:  "",
	}, data.Tags)

	// GetAllReplicas
	_, err = collection.GetAllReplicas(testDocumentID, &gocb.GetAllReplicaOptions{ParentSpan: instagocb.GetParentSpanFromContext(ctx)})
	a.NoError(err)

	span = getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CouchbaseSpanData{}, span.Data)
	data = span.Data.(instana.CouchbaseSpanData)
	a.Equal(instana.CouchbaseSpanTags{
		Bucket: testBucketName,
		Host:   "localhost",
		Type:   string(gocb.CouchbaseBucketType),
		SQL:    "GET_ALL_REPLICAS",
		Error:  "",
	}, data.Tags)

	// GetAnyReplica
	_, err = collection.GetAnyReplica(testDocumentID, &gocb.GetAnyReplicaOptions{ParentSpan: instagocb.GetParentSpanFromContext(ctx)})
	a.NoError(err)

	span = getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CouchbaseSpanData{}, span.Data)
	data = span.Data.(instana.CouchbaseSpanData)
	a.Equal(instana.CouchbaseSpanTags{
		Bucket: testBucketName,
		Host:   "localhost",
		Type:   string(gocb.CouchbaseBucketType),
		SQL:    "GET_ANY_REPLICA",
		Error:  "",
	}, data.Tags)

	// GetAndTouch
	_, err = collection.GetAndTouch(testDocumentID, time.Minute*20, &gocb.GetAndTouchOptions{ParentSpan: instagocb.GetParentSpanFromContext(ctx)})
	a.NoError(err)

	span = getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CouchbaseSpanData{}, span.Data)
	data = span.Data.(instana.CouchbaseSpanData)
	a.Equal(instana.CouchbaseSpanTags{
		Bucket: testBucketName,
		Host:   "localhost",
		Type:   string(gocb.CouchbaseBucketType),
		SQL:    "GET_AND_TOUCH",
		Error:  "",
	}, data.Tags)

	// GetAndLock
	ress, err := collection.GetAndLock(testDocumentID, time.Minute*20, &gocb.GetAndLockOptions{ParentSpan: instagocb.GetParentSpanFromContext(ctx)})
	a.NoError(err)

	span = getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CouchbaseSpanData{}, span.Data)
	data = span.Data.(instana.CouchbaseSpanData)
	a.Equal(instana.CouchbaseSpanTags{
		Bucket: testBucketName,
		Host:   "localhost",
		Type:   string(gocb.CouchbaseBucketType),
		SQL:    "GET_AND_LOCK",
		Error:  "",
	}, data.Tags)

	// Unlock
	err = collection.Unlock(testDocumentID, ress.Cas(), &gocb.UnlockOptions{ParentSpan: instagocb.GetParentSpanFromContext(ctx)})
	a.NoError(err)

	span = getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CouchbaseSpanData{}, span.Data)
	data = span.Data.(instana.CouchbaseSpanData)
	a.Equal(instana.CouchbaseSpanTags{
		Bucket: testBucketName,
		Host:   "localhost",
		Type:   string(gocb.CouchbaseBucketType),
		SQL:    "UNLOCK",
		Error:  "",
	}, data.Tags)

	// Touch
	_, err = collection.Touch(testDocumentID, time.Minute*20, &gocb.TouchOptions{ParentSpan: instagocb.GetParentSpanFromContext(ctx)})
	a.NoError(err)

	span = getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CouchbaseSpanData{}, span.Data)
	data = span.Data.(instana.CouchbaseSpanData)
	a.Equal(instana.CouchbaseSpanTags{
		Bucket: testBucketName,
		Host:   "localhost",
		Type:   string(gocb.CouchbaseBucketType),
		SQL:    "TOUCH",
		Error:  "",
	}, data.Tags)

	// LookupIn
	_, err = collection.LookupIn(testDocumentID, []gocb.LookupInSpec{
		gocb.GetSpec("test", &gocb.GetSpecOptions{}),
	}, &gocb.LookupInOptions{ParentSpan: instagocb.GetParentSpanFromContext(ctx)})
	a.NoError(err)

	span = getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CouchbaseSpanData{}, span.Data)
	data = span.Data.(instana.CouchbaseSpanData)
	a.Equal(instana.CouchbaseSpanTags{
		Bucket: testBucketName,
		Host:   "localhost",
		Type:   string(gocb.CouchbaseBucketType),
		SQL:    "LOOKUP_IN",
		Error:  "",
	}, data.Tags)

	// MutateIn
	_, err = collection.Upsert(testDocumentID, testDocumentValue, &gocb.UpsertOptions{ParentSpan: instagocb.GetParentSpanFromContext(ctx)})
	a.NoError(err)
	_, err = collection.MutateIn(testDocumentID, []gocb.MutateInSpec{
		gocb.UpsertSpec("foo", "311-555-0151", &gocb.UpsertSpecOptions{}),
	}, &gocb.MutateInOptions{})
	a.NoError(err)

	span = getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CouchbaseSpanData{}, span.Data)
	data = span.Data.(instana.CouchbaseSpanData)
	a.Equal(instana.CouchbaseSpanTags{
		Bucket: testBucketName,
		Host:   "localhost",
		Type:   string(gocb.CouchbaseBucketType),
		SQL:    "MUTATE_IN",
		Error:  "",
	}, data.Tags)

	//Binary Operations
	bc := collection.Binary()

	// Append
	_, err = bc.Append(testDocumentID, []byte{23}, &gocb.AppendOptions{ParentSpan: instagocb.GetParentSpanFromContext(ctx)})
	a.NoError(err)

	span = getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CouchbaseSpanData{}, span.Data)
	data = span.Data.(instana.CouchbaseSpanData)
	a.Equal(instana.CouchbaseSpanTags{
		Bucket: testBucketName,
		Host:   "localhost",
		Type:   string(gocb.CouchbaseBucketType),
		SQL:    "APPEND",
		Error:  "",
	}, data.Tags)

	// Prepend
	_, err = bc.Prepend(testDocumentID, []byte{23}, &gocb.PrependOptions{ParentSpan: instagocb.GetParentSpanFromContext(ctx)})
	a.NoError(err)

	span = getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CouchbaseSpanData{}, span.Data)
	data = span.Data.(instana.CouchbaseSpanData)
	a.Equal(instana.CouchbaseSpanTags{
		Bucket: testBucketName,
		Host:   "localhost",
		Type:   string(gocb.CouchbaseBucketType),
		SQL:    "PREPEND",
		Error:  "",
	}, data.Tags)

	// Remove
	_, err = collection.Remove(testDocumentID, &gocb.RemoveOptions{ParentSpan: instagocb.GetParentSpanFromContext(ctx)})
	a.NoError(err)

	span = getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CouchbaseSpanData{}, span.Data)
	data = span.Data.(instana.CouchbaseSpanData)
	a.Equal(instana.CouchbaseSpanTags{
		Bucket: testBucketName,
		Host:   "localhost",
		Type:   string(gocb.CouchbaseBucketType),
		SQL:    "REMOVE",
		Error:  "",
	}, data.Tags)

	// Increment
	_, err = bc.Increment(testDocumentID, &gocb.IncrementOptions{
		Initial:    2,
		ParentSpan: instagocb.GetParentSpanFromContext(ctx),
	})
	a.NoError(err)

	span = getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CouchbaseSpanData{}, span.Data)
	data = span.Data.(instana.CouchbaseSpanData)
	a.Equal(instana.CouchbaseSpanTags{
		Bucket: testBucketName,
		Host:   "localhost",
		Type:   string(gocb.CouchbaseBucketType),
		SQL:    "INCREMENT",
		Error:  "",
	}, data.Tags)

	// Decrement
	_, err = bc.Decrement(testDocumentID, &gocb.DecrementOptions{
		Initial:    2,
		ParentSpan: instagocb.GetParentSpanFromContext(ctx),
	})
	a.NoError(err)

	span = getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CouchbaseSpanData{}, span.Data)
	data = span.Data.(instana.CouchbaseSpanData)
	a.Equal(instana.CouchbaseSpanTags{
		Bucket: testBucketName,
		Host:   "localhost",
		Type:   string(gocb.CouchbaseBucketType),
		SQL:    "DECREMENT",
		Error:  "",
	}, data.Tags)

	// Bulk operations
	_, err = collection.Insert("test-bulk-1", "test-bulk-value-1", &gocb.InsertOptions{ParentSpan: instagocb.GetParentSpanFromContext(ctx)})
	a.NoError(err)
	_, err = collection.Insert("test-bulk-2", "test-bulk-value-2", &gocb.InsertOptions{ParentSpan: instagocb.GetParentSpanFromContext(ctx)})
	a.NoError(err)

	var get1, get2 gocb.GetResult
	var str1, str2 string
	var itemsGet []gocb.BulkOp
	itemsGet = append(itemsGet, &gocb.GetOp{ID: "test-bulk-1", Result: &get1})
	itemsGet = append(itemsGet, &gocb.GetOp{ID: "test-bulk-2", Result: &get2})

	err = collection.Do(itemsGet, &gocb.BulkOpOptions{ParentSpan: instagocb.GetParentSpanFromContext(ctx)})
	a.NoError(err)
	item1 := itemsGet[0].(*gocb.GetOp)
	a.NoError(item1.Err)
	a.NoError(item1.Result.Content(&str1))

	item2 := itemsGet[1].(*gocb.GetOp)
	a.NoError(item2.Err)
	a.NoError(item2.Result.Content(&str2))

	a.Equal("test-bulk-value-1", str1)
	a.Equal("test-bulk-value-2", str2)

}
