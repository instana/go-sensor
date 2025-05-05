// (c) Copyright IBM Corp. 2023

//go:build integration
// +build integration

package instagocb_test

// func TestScope(t *testing.T) {
// 	defer instana.ShutdownCollector()

// 	recorder, ctx, cluster, a := prepare(t)
// 	defer cluster.Close(&gocb.ClusterCloseOptions{})

// 	scope := cluster.Bucket(cbTestBucket).Scope(cbTestScope)

// 	// Query
// 	q := "SELECT count(*) FROM `" + cbTestBucket + "`." + cbTestScope + "." + cbTestCollection + ";"
// 	_, err := scope.Query(q, &gocb.QueryOptions{ParentSpan: instagocb.GetParentSpanFromContext(ctx)})
// 	a.NoError(err)

// 	span := getLatestSpan(recorder)
// 	a.Equal(0, span.Ec)
// 	a.EqualValues(instana.ExitSpanKind, span.Kind)
// 	a.IsType(instana.CouchbaseSpanData{}, span.Data)
// 	data := span.Data.(instana.CouchbaseSpanData)
// 	a.Equal(instana.CouchbaseSpanTags{
// 		Bucket: cbTestBucket,
// 		Host:   "localhost",
// 		Type:   string(gocb.CouchbaseBucketType),
// 		SQL:    q,
// 		Error:  "",
// 	}, data.Tags)

// }
