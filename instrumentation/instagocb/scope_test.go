// (c) Copyright IBM Corp. 2023

//go:build integration
// +build integration

package instagocb_test

import (
	"testing"

	"github.com/couchbase/gocb/v2"
	instana "github.com/instana/go-sensor"
)

func TestScope(t *testing.T) {
	defer instana.ShutdownSensor()
	recorder, _, cluster, a, _ := prepareWithATestDocumentInCollection(t, "scope")

	scope := cluster.Bucket(testBucketName).Scope(testScope)

	// Query
	q := "SELECT count(*) FROM `" + testBucketName + "`." + testScope + "." + testCollection + ";"
	_, err := scope.Query(q, &gocb.QueryOptions{})
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
		SQL:    q,
		Error:  "",
	}, data.Tags)

}
