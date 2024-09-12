// (c) Copyright IBM Corp. 2023

//go:build integration
// +build integration

package instagocb_test

import (
	"testing"
	"time"

	"github.com/couchbase/gocb/v2"
	"github.com/couchbase/gocb/v2/search"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instagocb"
)

func TestCluster(t *testing.T) {
	defer instana.ShutdownSensor()

	recorder, ctx, cluster, a := prepare(t)
	defer cluster.Close(&gocb.ClusterCloseOptions{})

	// Query
	q := "SELECT count(*) FROM `" + cbTestBucket + "`." + cbTestScope + "." + cbTestCollection + ";"
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

	// Search Query
	err = cluster.SearchIndexes().UpsertIndex(
		gocb.SearchIndex{
			Type:       "fulltext-index",
			Name:       "sample-index" + testID,
			SourceName: cbTestBucket,
			SourceType: "couchbase",
			PlanParams: map[string]interface{}{
				"maxPartitionsPerPIndex": 171,
			},
			Params: map[string]interface{}{
				"doc_config": map[string]interface{}{
					"mode":       "type_field",
					"type_field": "type",
				},
				"mapping": map[string]interface{}{
					"default_analyzer":        "standard",
					"default_datetime_parser": "dateTimeOptional",
					"default_field":           "_all",
					"default_mapping": map[string]interface{}{
						"dynamic": true,
						"enabled": true,
					},
					"default_type":  "_default",
					"index_dynamic": true,
					"store_dynamic": false,
				},
				"store": map[string]interface{}{
					"kvStoreName": "mossStore",
				},
			},
			SourceParams: map[string]interface{}{},
		},
		&gocb.UpsertSearchIndexOptions{
			ParentSpan: instagocb.GetParentSpanFromContext(ctx),
		},
	)

	a.NoError(err)
	time.Sleep(4 * time.Second)

	matchResult, err := cluster.SearchQuery(
		"sample-index"+testID,
		search.NewMatchQuery("test"),
		&gocb.SearchOptions{
			Limit:      10,
			Fields:     []string{"foo", "bar"},
			ParentSpan: instagocb.GetParentSpanFromContext(ctx),
		},
	)
	a.NoError(err)

	for matchResult.Next() {
		id := matchResult.Row().ID
		a.Equal("test-doc-id", id)

	}

	a.NoError(matchResult.Err())

	span = getLatestSpan(recorder)
	a.Equal(0, span.Ec)
	a.EqualValues(instana.ExitSpanKind, span.Kind)
	a.IsType(instana.CouchbaseSpanData{}, span.Data)
	data = span.Data.(instana.CouchbaseSpanData)
	a.Equal(instana.CouchbaseSpanTags{
		Bucket: "",
		Host:   "localhost",
		Type:   "",
		SQL:    "SEARCH sample-index" + testID,
		Error:  "",
	}, data.Tags)

}
