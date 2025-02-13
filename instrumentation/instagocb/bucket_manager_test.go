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

func TestBucketManager(t *testing.T) {
	defer instana.ShutdownCollector()
	recorder, ctx, cluster, a := prepare(t)
	defer cluster.Close(&gocb.ClusterCloseOptions{})

	bucketMgr := cluster.Buckets()

	bs := gocb.BucketSettings{
		Name:                 testBucketName,
		FlushEnabled:         true,
		ReplicaIndexDisabled: true,
		RAMQuotaMB:           150,
		NumReplicas:          1,
		BucketType:           gocb.CouchbaseBucketType,
	}

	// creation
	err := bucketMgr.CreateBucket(gocb.CreateBucketSettings{
		BucketSettings:         bs,
		ConflictResolutionType: gocb.ConflictResolutionTypeSequenceNumber,
	}, &gocb.CreateBucketOptions{
		Context:    ctx,
		ParentSpan: instagocb.GetParentSpanFromContext(ctx),
	})
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
		SQL:    "CREATE_BUCKET",
		Error:  "",
	}, data.Tags)

	// Get
	bsRes, err := bucketMgr.GetBucket(testBucketName, &gocb.GetBucketOptions{ParentSpan: instagocb.GetParentSpanFromContext(ctx)})
	a.NoError(err)
	a.Equal(bs.Name, bsRes.Name)
	a.Equal(bs.BucketType, bsRes.BucketType)
	span = getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CouchbaseSpanData{}, span.Data)
	data = span.Data.(instana.CouchbaseSpanData)
	a.Equal(instana.CouchbaseSpanTags{
		Bucket: testBucketName,
		Host:   "localhost",
		Type:   string(gocb.CouchbaseBucketType),
		SQL:    "GET_BUCKET",
		Error:  "",
	}, data.Tags)

	// Flush
	err = bucketMgr.FlushBucket(testBucketName, &gocb.FlushBucketOptions{ParentSpan: instagocb.GetParentSpanFromContext(ctx)})
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
		SQL:    "FLUSH_BUCKET",
		Error:  "",
	}, data.Tags)

	// Update
	bs.RAMQuotaMB = 200
	err = bucketMgr.UpdateBucket(bs, &gocb.UpdateBucketOptions{ParentSpan: instagocb.GetParentSpanFromContext(ctx)})
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
		SQL:    "UPDATE_BUCKET",
		Error:  "",
	}, data.Tags)
	bsRes, err = bucketMgr.GetBucket(testBucketName, &gocb.GetBucketOptions{})
	a.Equal(bs.RAMQuotaMB, bsRes.RAMQuotaMB)
	a.NoError(err)

	// Drop
	err = bucketMgr.DropBucket(testBucketName, &gocb.DropBucketOptions{ParentSpan: instagocb.GetParentSpanFromContext(ctx)})
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
		SQL:    "DROP_BUCKET",
		Error:  "",
	}, data.Tags)

	// Checking error
	err = bucketMgr.DropBucket(testBucketName, &gocb.DropBucketOptions{ParentSpan: instagocb.GetParentSpanFromContext(ctx)})
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
	a.Contains(data.Tags.Error, "bucket not found")

}
