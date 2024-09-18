// (c) Copyright IBM Corp. 2023

//go:build integration
// +build integration

package instagocb_test

import (
	"testing"

	"github.com/couchbase/gocb/v2"
	instana "github.com/instana/go-sensor"
)

func TestCollection_DS_List(t *testing.T) {

	// setting this environment variable as the operation doesn't support parent span
	setAllowRootExitSpanEnv()
	defer unsetAllowRootExitSpanEnv()

	defer instana.ShutdownSensor()

	recorder, _, cluster, a := prepare(t)
	defer cluster.Close(&gocb.ClusterCloseOptions{})

	collection := cluster.Bucket(cbTestBucket).Scope(cbTestScope).Collection(cbTestCollection)

	// List
	l := collection.List(dsListTestDocumentID)

	// Iterator
	_, err := l.Iterator()
	a.NoError(err)

	span := getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CouchbaseSpanData{}, span.Data)
	data := span.Data.(instana.CouchbaseSpanData)
	a.Equal(instana.CouchbaseSpanTags{
		Bucket: cbTestBucket,
		Host:   "localhost",
		Type:   string(gocb.CouchbaseBucketType),
		SQL:    "LIST_ITERATOR",
		Error:  "",
	}, data.Tags)

	// At
	var result string
	err = l.At(0, &result)
	a.NoError(err)

	span = getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CouchbaseSpanData{}, span.Data)
	data = span.Data.(instana.CouchbaseSpanData)
	a.Equal(instana.CouchbaseSpanTags{
		Bucket: cbTestBucket,
		Host:   "localhost",
		Type:   string(gocb.CouchbaseBucketType),
		SQL:    "LIST_AT",
		Error:  "",
	}, data.Tags)

	// Append
	err = l.Append("test-foo-1")
	a.NoError(err)

	span = getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CouchbaseSpanData{}, span.Data)
	data = span.Data.(instana.CouchbaseSpanData)
	a.Equal(instana.CouchbaseSpanTags{
		Bucket: cbTestBucket,
		Host:   "localhost",
		Type:   string(gocb.CouchbaseBucketType),
		SQL:    "LIST_APPEND",
		Error:  "",
	}, data.Tags)

	// Prepend
	err = l.Prepend("test-bar-1")
	a.NoError(err)

	span = getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CouchbaseSpanData{}, span.Data)
	data = span.Data.(instana.CouchbaseSpanData)
	a.Equal(instana.CouchbaseSpanTags{
		Bucket: cbTestBucket,
		Host:   "localhost",
		Type:   string(gocb.CouchbaseBucketType),
		SQL:    "LIST_PREPEND",
		Error:  "",
	}, data.Tags)

	// IndexOf
	i, err := l.IndexOf("test-value-string")
	a.NoError(err)
	a.Equal(1, i)

	span = getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CouchbaseSpanData{}, span.Data)
	data = span.Data.(instana.CouchbaseSpanData)
	a.Equal(instana.CouchbaseSpanTags{
		Bucket: cbTestBucket,
		Host:   "localhost",
		Type:   string(gocb.CouchbaseBucketType),
		SQL:    "LIST_INDEX_OF",
		Error:  "",
	}, data.Tags)

	// Size
	c, err := l.Size()
	a.NoError(err)
	a.Equal(3, c)

	span = getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CouchbaseSpanData{}, span.Data)
	data = span.Data.(instana.CouchbaseSpanData)
	a.Equal(instana.CouchbaseSpanTags{
		Bucket: cbTestBucket,
		Host:   "localhost",
		Type:   string(gocb.CouchbaseBucketType),
		SQL:    "LIST_SIZE",
		Error:  "",
	}, data.Tags)

	// RemoveAt
	err = l.RemoveAt(0)
	a.NoError(err)

	span = getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CouchbaseSpanData{}, span.Data)
	data = span.Data.(instana.CouchbaseSpanData)
	a.Equal(instana.CouchbaseSpanTags{
		Bucket: cbTestBucket,
		Host:   "localhost",
		Type:   string(gocb.CouchbaseBucketType),
		SQL:    "LIST_REMOVE_AT",
		Error:  "",
	}, data.Tags)

	// Clear
	err = l.Clear()
	a.NoError(err)

	span = getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CouchbaseSpanData{}, span.Data)
	data = span.Data.(instana.CouchbaseSpanData)
	a.Equal(instana.CouchbaseSpanTags{
		Bucket: cbTestBucket,
		Host:   "localhost",
		Type:   string(gocb.CouchbaseBucketType),
		SQL:    "LIST_CLEAR",
		Error:  "",
	}, data.Tags)

}
