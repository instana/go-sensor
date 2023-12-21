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
	defer instana.ShutdownSensor()
	recorder, ctx, cluster, a := prepareWithCollection(t)

	scope := cluster.Bucket(testBucketName).Scope(testScope)

	collection := scope.Collection(testCollection)

	transaction := cluster.Transactions()

	// Just to clear all spans
	recorder.GetQueuedSpans()

	_, err := transaction.Run(
		func(tac *gocb.TransactionAttemptContext) error {
			// Insert
			c := cluster.WrapTransactionAttemptContext(tac, instagocb.GetParentSpanFromContext(ctx))
			_, err := c.Insert(collection.Unwrap(), testDocumentID, testDocumentValue)
			a.NoError(err)

			//Get
			var result myDoc
			res, err := c.Get(collection.Unwrap(), testDocumentID)
			a.NoError(err)
			res.Content(&result)
			a.Equal(testDocumentValue, result)

			// Replace
			_, err = c.Replace(collection.Unwrap(), res, "new-value")
			a.NoError(err)

			// Remove
			res, err = c.Get(collection.Unwrap(), testDocumentID)
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

	getFlag := 0
	insertFlag := 0
	replaceFlag := 0
	removeFlag := 0

	for _, span := range spans {

		a.Equal(0, span.Ec)
		a.EqualValues(instana.ExitSpanKind, span.Kind)
		a.IsType(instana.CouchbaseSpanData{}, span.Data)
		data := span.Data.(instana.CouchbaseSpanData)

		switch data.Tags.SQL {
		case "TRANSACTION_INSERT":
			a.Equalf(insertFlag, 0, "More than expected insert calls!")
			insertFlag += 1
			a.Equal(instana.CouchbaseSpanTags{
				Bucket: testBucketName,
				Host:   "localhost",
				Type:   string(gocb.CouchbaseBucketType),
				SQL:    "TRANSACTION_INSERT",
				Error:  "",
			}, data.Tags)

		case "TRANSACTION_GET":
			a.Containsf([]int{0, 1}, getFlag, "More than expected get calls!")
			getFlag += 1
			a.Equal(instana.CouchbaseSpanTags{
				Bucket: testBucketName,
				Host:   "localhost",
				Type:   string(gocb.CouchbaseBucketType),
				SQL:    "TRANSACTION_GET",
				Error:  "",
			}, data.Tags)

		case "TRANSACTION_REPLACE":
			a.Equalf(replaceFlag, 0, "More than expected replace calls!")
			replaceFlag += 1
			a.Equal(instana.CouchbaseSpanTags{
				Bucket: testBucketName,
				Host:   "localhost",
				Type:   string(gocb.CouchbaseBucketType),
				SQL:    "TRANSACTION_REPLACE",
				Error:  "",
			}, data.Tags)

		case "TRANSACTION_REMOVE":
			a.Equalf(removeFlag, 0, "More than expected remove calls!")
			removeFlag += 1
			a.Equal(instana.CouchbaseSpanTags{
				Bucket: testBucketName,
				Host:   "localhost",
				Type:   string(gocb.CouchbaseBucketType),
				SQL:    "TRANSACTION_REMOVE",
				Error:  "",
			}, data.Tags)
		}
	}

	recorder, ctx, cluster, a, _ = prepareWithATestDocumentInCollection(t, "scope")
	scope = cluster.Bucket(testBucketName).Scope(testScope)
	collection = scope.Collection(testCollection)
	transaction = cluster.Transactions()

	q := "SELECT count(*) FROM `" + testBucketName + "`." + testScope + "." + testCollection + ";"
	_, err = transaction.Run(
		func(tac *gocb.TransactionAttemptContext) error {
			// Query
			c := cluster.WrapTransactionAttemptContext(tac, instagocb.GetParentSpanFromContext(ctx))
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

}
