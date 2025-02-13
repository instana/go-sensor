// (c) Copyright IBM Corp. 2023

//go:build integration
// +build integration

package instagocb_test

import (
	"testing"

	"github.com/couchbase/gocb/v2"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instagocb"
)

func TestCollectionManager(t *testing.T) {
	defer instana.ShutdownCollector()

	recorder, ctx, cluster, a := prepareWithBucket(t, testBucketName)
	defer cluster.Close(&gocb.ClusterCloseOptions{})

	bucket := cluster.Bucket(testBucketName)
	cm := bucket.Collections()

	// create scope
	err := cm.CreateScope(testScope, &gocb.CreateScopeOptions{ParentSpan: instagocb.GetParentSpanFromContext(ctx)})
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
		SQL:    "CREATE_SCOPE",
		Error:  "",
	}, data.Tags)

	// create collection
	err = cm.CreateCollection(gocb.CollectionSpec{
		Name:      testCollection,
		ScopeName: testScope,
	}, &gocb.CreateCollectionOptions{ParentSpan: instagocb.GetParentSpanFromContext(ctx)})
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
		SQL:    "CREATE_COLLECTION",
		Error:  "",
	}, data.Tags)

	// Drop collection
	err = cm.DropCollection(gocb.CollectionSpec{
		Name:      testCollection,
		ScopeName: testScope,
	}, &gocb.DropCollectionOptions{ParentSpan: instagocb.GetParentSpanFromContext(ctx)})
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
		SQL:    "DROP_COLLECTION",
		Error:  "",
	}, data.Tags)

	// Drop scope
	err = cm.DropScope(testScope, &gocb.DropScopeOptions{ParentSpan: instagocb.GetParentSpanFromContext(ctx)})
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
		SQL:    "DROP_SCOPE",
		Error:  "",
	}, data.Tags)

	// Checking error
	err = cm.DropScope(testScope, &gocb.DropScopeOptions{ParentSpan: instagocb.GetParentSpanFromContext(ctx)})
	a.Error(err)

	spans := recorder.GetQueuedSpans()
	span, logSpan := spans[0], spans[1]
	a.NotEqual(0, span.Ec)
	a.Equal(span.TraceID, logSpan.TraceID)
	a.Equal(span.SpanID, logSpan.ParentID)
	a.Equal("log.go", logSpan.Name)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CouchbaseSpanData{}, span.Data)
	data = span.Data.(instana.CouchbaseSpanData)
	a.Contains(data.Tags.Error, "scope not found")

}
