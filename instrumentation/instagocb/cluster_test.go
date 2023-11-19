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

func TestCluster(t *testing.T) {
	defer instana.ShutdownSensor()
	recorder, ctx, cluster, a, _ := prepareWithATestDocumentInCollection(t, "cluster")

	// Query
	q := "SELECT count(*) FROM `" + testBucketName + "`." + testScope + "." + testCollection + ";"
	_, err := cluster.Query(q, &gocb.QueryOptions{ParentSpan: instagocb.GetParentSpanFromContext(ctx)})
	a.NoError(err)

	span := getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CouchbaseSpanData{}, span.Data)
	data := span.Data.(instana.CouchbaseSpanData)
	a.Equal(instana.CouchbaseSpanTags{
		Bucket: "",
		Host:   "localhost",
		Type:   "",
		SQL:    q,
		Error:  "",
	}, data.Tags)

}
