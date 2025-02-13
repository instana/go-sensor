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

func TestTransactions(t *testing.T) {
	testDocumentValue := getTestDocumentValue()
	defer instana.ShutdownCollector()
	recorder, ctx, cluster, a := prepare(t)

	scope := cluster.Bucket(cbTestBucket).Scope(cbTestScope)

	collection := scope.Collection(cbTestCollection)

	transaction := cluster.Transactions()

	// Just to clear all spans
	recorder.GetQueuedSpans()

	_, err := transaction.Run(
		func(tac *gocb.TransactionAttemptContext) error {
			// Insert
			c := cluster.WrapTransactionAttemptContext(tac, instagocb.GetParentSpanFromContext(ctx))

			//Get
			var result myDoc
			res, err := c.Get(collection.Unwrap(), transactionTestDocumentID)
			a.NoError(err)
			res.Content(&result)
			a.Equal(testDocumentValue, result)

			// Replace
			_, err = c.Replace(collection.Unwrap(), res, "new-value")
			a.NoError(err)

			// Remove
			res, err = c.Get(collection.Unwrap(), transactionTestDocumentID)
			a.NoError(err)
			err = c.Remove(collection.Unwrap(), res)
			a.NoError(err)

			return nil

		},
		&gocb.TransactionOptions{},
	)

	a.NoError(err)

	// asserting all spans recorded during transaction
	spans := recorder.GetQueuedSpans()
	for i, span := range spans {
		switch i {
		case 0, 2:
			a.Equal(0, span.Ec)
			a.EqualValues(instana.ExitSpanKind, span.Kind)
			a.IsType(instana.CouchbaseSpanData{}, span.Data)
			data := span.Data.(instana.CouchbaseSpanData)
			a.Equal(instana.CouchbaseSpanTags{
				Bucket: cbTestBucket,
				Host:   "localhost",
				Type:   string(gocb.CouchbaseBucketType),
				SQL:    "TRANSACTION_GET",
				Error:  "",
			}, data.Tags)

		case 1:
			a.Equal(0, span.Ec)
			a.EqualValues(instana.ExitSpanKind, span.Kind)
			a.IsType(instana.CouchbaseSpanData{}, span.Data)
			data := span.Data.(instana.CouchbaseSpanData)
			a.Equal(instana.CouchbaseSpanTags{
				Bucket: cbTestBucket,
				Host:   "localhost",
				Type:   string(gocb.CouchbaseBucketType),
				SQL:    "TRANSACTION_REPLACE",
				Error:  "",
			}, data.Tags)

		case 3:
			a.Equal(0, span.Ec)
			a.EqualValues(instana.ExitSpanKind, span.Kind)
			a.IsType(instana.CouchbaseSpanData{}, span.Data)
			data := span.Data.(instana.CouchbaseSpanData)
			a.Equal(instana.CouchbaseSpanTags{
				Bucket: cbTestBucket,
				Host:   "localhost",
				Type:   string(gocb.CouchbaseBucketType),
				SQL:    "TRANSACTION_REMOVE",
				Error:  "",
			}, data.Tags)
		}
	}

	cluster.Close(&gocb.ClusterCloseOptions{})

	recorder, ctx, conn, a := prepare(t)
	scope = conn.Bucket(cbTestBucket).Scope(cbTestScope)
	collection = scope.Collection(cbTestCollection)
	transaction = conn.Transactions()

	q := "SELECT count(*) FROM `" + cbTestBucket + "`." + cbTestScope + "." + cbTestCollection + ";"
	_, err = transaction.Run(
		func(tac *gocb.TransactionAttemptContext) error {
			// Query
			c := conn.WrapTransactionAttemptContext(tac, instagocb.GetParentSpanFromContext(ctx))
			_, err := c.Query(q, &gocb.TransactionQueryOptions{})
			a.NoError(err)

			return nil

		},
		&gocb.TransactionOptions{},
	)

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

	conn.Close(&gocb.ClusterCloseOptions{})

}
