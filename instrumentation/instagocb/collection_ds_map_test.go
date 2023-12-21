// (c) Copyright IBM Corp. 2023

//go:build integration
// +build integration

package instagocb_test

import (
	"testing"

	"github.com/couchbase/gocb/v2"
	instana "github.com/instana/go-sensor"
)

func TestCollection_DS_Map(t *testing.T) {
	defer instana.ShutdownSensor()
	recorder, _, cluster, a, _ := prepareWithATestDocumentInCollection(t, "ds_map")

	collection := cluster.Bucket(testBucketName).Scope(testScope).Collection(testCollection)

	// Map
	m := collection.Map(testDocumentID)

	// Iterator
	_, err := m.Iterator()
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
		SQL:    "MAP_ITERATOR",
		Error:  "",
	}, data.Tags)

	// At
	var result string
	err = m.At("test-key", &result)
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
		SQL:    "MAP_AT",
		Error:  "",
	}, data.Tags)

	// Add
	err = m.Add("test-key-1", "test-value-1")
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
		SQL:    "MAP_ADD",
		Error:  "",
	}, data.Tags)

	// Remove
	err = m.Remove("test-key-1")
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
		SQL:    "MAP_REMOVE",
		Error:  "",
	}, data.Tags)

	// Exists
	isExists, err := m.Exists("test-key")
	a.NoError(err)
	a.True(isExists)

	span = getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CouchbaseSpanData{}, span.Data)
	data = span.Data.(instana.CouchbaseSpanData)
	a.Equal(instana.CouchbaseSpanTags{
		Bucket: testBucketName,
		Host:   "localhost",
		Type:   string(gocb.CouchbaseBucketType),
		SQL:    "MAP_EXISTS",
		Error:  "",
	}, data.Tags)

	// Size
	c, err := m.Size()
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
		SQL:    "MAP_SIZE",
		Error:  "",
	}, data.Tags)

	// Keys
	keys, err := m.Keys()
	a.NoError(err)
	a.Equal([]string{"test-key"}, keys)

	span = getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CouchbaseSpanData{}, span.Data)
	data = span.Data.(instana.CouchbaseSpanData)
	a.Equal(instana.CouchbaseSpanTags{
		Bucket: testBucketName,
		Host:   "localhost",
		Type:   string(gocb.CouchbaseBucketType),
		SQL:    "MAP_KEYS",
		Error:  "",
	}, data.Tags)

	// Values
	values, err := m.Values()
	a.NoError(err)
	a.Equal("test-value-string", values[0])

	span = getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CouchbaseSpanData{}, span.Data)
	data = span.Data.(instana.CouchbaseSpanData)
	a.Equal(instana.CouchbaseSpanTags{
		Bucket: testBucketName,
		Host:   "localhost",
		Type:   string(gocb.CouchbaseBucketType),
		SQL:    "MAP_VALUES",
		Error:  "",
	}, data.Tags)

	// Clear
	err = m.Clear()
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
		SQL:    "MAP_CLEAR",
		Error:  "",
	}, data.Tags)
	a.NoError(cluster.Close(&gocb.ClusterCloseOptions{}))
}
