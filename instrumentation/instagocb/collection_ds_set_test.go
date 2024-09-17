// (c) Copyright IBM Corp. 2023

//go:build integration
// +build integration

package instagocb_test

import (
	"testing"

	"github.com/couchbase/gocb/v2"
	instana "github.com/instana/go-sensor"
)

func TestCollection_DS_Set(t *testing.T) {

	// setting this environment variable as the operation doesn't support parent span
	setAllowRootExitSpanEnv()
	defer unsetAllowRootExitSpanEnv()

	defer instana.ShutdownSensor()

	recorder, _, cluster, a := prepare(t)
	defer cluster.Close(&gocb.ClusterCloseOptions{})

	collection := cluster.Bucket(cbTestBucket).Scope(cbTestScope).Collection(cbTestCollection)

	// Set
	s := collection.Set(dsSetTestDocumentID)

	// Iterator
	_, err := s.Iterator()
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
		SQL:    "SET_ITERATOR",
		Error:  "",
	}, data.Tags)

	// Add
	err = s.Add("test-value-2")
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
		SQL:    "SET_ADD",
		Error:  "",
	}, data.Tags)

	// Remove
	err = s.Remove("test-value-2")
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
		SQL:    "SET_REMOVE",
		Error:  "",
	}, data.Tags)

	// Contains
	r, err := s.Contains("test-value-2")
	a.NoError(err)
	a.False(r)

	span = getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CouchbaseSpanData{}, span.Data)
	data = span.Data.(instana.CouchbaseSpanData)
	a.Equal(instana.CouchbaseSpanTags{
		Bucket: cbTestBucket,
		Host:   "localhost",
		Type:   string(gocb.CouchbaseBucketType),
		SQL:    "SET_CONTAINS",
		Error:  "",
	}, data.Tags)

	// Values
	res, err := s.Values()
	a.NoError(err)
	a.Equal("test-value-string", res[0])

	span = getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CouchbaseSpanData{}, span.Data)
	data = span.Data.(instana.CouchbaseSpanData)
	a.Equal(instana.CouchbaseSpanTags{
		Bucket: cbTestBucket,
		Host:   "localhost",
		Type:   string(gocb.CouchbaseBucketType),
		SQL:    "SET_VALUES",
		Error:  "",
	}, data.Tags)

	// Size
	c, err := s.Size()
	a.NoError(err)
	a.Equal(1, c)

	span = getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CouchbaseSpanData{}, span.Data)
	data = span.Data.(instana.CouchbaseSpanData)
	a.Equal(instana.CouchbaseSpanTags{
		Bucket: cbTestBucket,
		Host:   "localhost",
		Type:   string(gocb.CouchbaseBucketType),
		SQL:    "SET_SIZE",
		Error:  "",
	}, data.Tags)

	// Clear
	err = s.Clear()
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
		SQL:    "SET_CLEAR",
		Error:  "",
	}, data.Tags)
}
