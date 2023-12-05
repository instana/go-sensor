// (c) Copyright IBM Corp. 2023

//go:build integration
// +build integration

package instagocb_test

import (
	"testing"

	"github.com/couchbase/gocb/v2"
	instana "github.com/instana/go-sensor"
)

func TestCollection_DS_Queue(t *testing.T) {
	defer instana.ShutdownSensor()
	recorder, _, cluster, a, _ := prepareWithATestDocumentInCollection(t, "ds_queue")

	collection := cluster.Bucket(testBucketName).Scope(testScope).Collection(testCollection)

	// Queue
	q := collection.Queue(testDocumentID)

	// Iterator
	_, err := q.Iterator()
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
		SQL:    "QUEUE_ITERATOR",
		Error:  "",
	}, data.Tags)

	// Push
	err = q.Push("test-value-2")
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
		SQL:    "QUEUE_PUSH",
		Error:  "",
	}, data.Tags)

	// Pop
	var result string
	err = q.Pop(&result)
	a.NoError(err)
	a.Equal("test-value-string", result)

	span = getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CouchbaseSpanData{}, span.Data)
	data = span.Data.(instana.CouchbaseSpanData)
	a.Equal(instana.CouchbaseSpanTags{
		Bucket: testBucketName,
		Host:   "localhost",
		Type:   string(gocb.CouchbaseBucketType),
		SQL:    "QUEUE_POP",
		Error:  "",
	}, data.Tags)

	// Size
	c, err := q.Size()
	a.NoError(err)
	a.Equal(1, c)

	span = getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CouchbaseSpanData{}, span.Data)
	data = span.Data.(instana.CouchbaseSpanData)
	a.Equal(instana.CouchbaseSpanTags{
		Bucket: testBucketName,
		Host:   "localhost",
		Type:   string(gocb.CouchbaseBucketType),
		SQL:    "QUEUE_SIZE",
		Error:  "",
	}, data.Tags)

	// Clear
	err = q.Clear()
	a.NoError(err)
	a.Equal(1, c)

	span = getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CouchbaseSpanData{}, span.Data)
	data = span.Data.(instana.CouchbaseSpanData)
	a.Equal(instana.CouchbaseSpanTags{
		Bucket: testBucketName,
		Host:   "localhost",
		Type:   string(gocb.CouchbaseBucketType),
		SQL:    "QUEUE_CLEAR",
		Error:  "",
	}, data.Tags)

}
