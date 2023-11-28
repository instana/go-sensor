// (c) Copyright IBM Corp. 2023

package instagocb

import (
	"github.com/couchbase/gocb/v2"
)

type TransactionAttemptContext interface {
	Query(statement string, options *gocb.TransactionQueryOptions) (*gocb.TransactionQueryResult, error)
	Get(collection *gocb.Collection, id string) (*gocb.TransactionGetResult, error)
	Replace(collection *gocb.Collection, doc *gocb.TransactionGetResult, value interface{}) (*gocb.TransactionGetResult, error)
	Insert(collection *gocb.Collection, id string, value interface{}) (*gocb.TransactionGetResult, error)
	Remove(collection *gocb.Collection, doc *gocb.TransactionGetResult) error

	Unwrap() *gocb.TransactionAttemptContext
}

type instaTransactionAttemptContext struct {
	iTracer requestTracer
	*gocb.TransactionAttemptContext

	// Need this here to get the parent child relationship for transaction spans.
	parentSpan gocb.RequestSpan
}

// Query executes the query statement on the server.
func (itac *instaTransactionAttemptContext) Query(statement string, options *gocb.TransactionQueryOptions) (*gocb.TransactionQueryResult, error) {
	var tracectx gocb.RequestSpanContext
	if itac.parentSpan != nil {
		tracectx = itac.parentSpan.Context()
	}

	span := itac.iTracer.RequestSpan(tracectx, "TRANSACTION_QUERY")
	span.SetAttribute(operationSpanTag, statement)

	res, err := itac.TransactionAttemptContext.Query(statement, options)

	span.(*Span).err = err

	defer span.End()

	return res, err
}

// Get will attempt to fetch a document, and fail the transaction if it does not exist.
func (itac *instaTransactionAttemptContext) Get(collection *gocb.Collection, id string) (*gocb.TransactionGetResult, error) {
	var tracectx gocb.RequestSpanContext
	if itac.parentSpan != nil {
		tracectx = itac.parentSpan.Context()
	}

	span := itac.iTracer.RequestSpan(tracectx, "TRANSACTION_GET")
	span.SetAttribute(bucketNameSpanTag, collection.Bucket().Name())

	res, err := itac.TransactionAttemptContext.Get(collection, id)

	span.(*Span).err = err

	defer span.End()

	return res, err
}

// Replace will replace the contents of a document, failing if the document does not already exist.
func (itac *instaTransactionAttemptContext) Replace(collection *gocb.Collection, doc *gocb.TransactionGetResult, value interface{}) (*gocb.TransactionGetResult, error) {
	var tracectx gocb.RequestSpanContext
	if itac.parentSpan != nil {
		tracectx = itac.parentSpan.Context()
	}

	span := itac.iTracer.RequestSpan(tracectx, "TRANSACTION_REPLACE")
	span.SetAttribute(bucketNameSpanTag, collection.Bucket().Name())

	res, err := itac.TransactionAttemptContext.Replace(doc, value)

	span.(*Span).err = err

	defer span.End()

	return res, err
}

// Insert will insert a new document, failing if the document already exists.
func (itac *instaTransactionAttemptContext) Insert(collection *gocb.Collection, id string, value interface{}) (*gocb.TransactionGetResult, error) {
	var tracectx gocb.RequestSpanContext
	if itac.parentSpan != nil {
		tracectx = itac.parentSpan.Context()
	}

	span := itac.iTracer.RequestSpan(tracectx, "TRANSACTION_INSERT")
	span.SetAttribute(bucketNameSpanTag, collection.Bucket().Name())

	res, err := itac.TransactionAttemptContext.Insert(collection, id, value)

	span.(*Span).err = err

	defer span.End()

	return res, err
}

// Remove will delete a document.
func (itac *instaTransactionAttemptContext) Remove(collection *gocb.Collection, doc *gocb.TransactionGetResult) error {
	var tracectx gocb.RequestSpanContext
	if itac.parentSpan != nil {
		tracectx = itac.parentSpan.Context()
	}

	span := itac.iTracer.RequestSpan(tracectx, "TRANSACTION_INSERT")
	span.SetAttribute(bucketNameSpanTag, collection.Bucket().Name())

	err := itac.TransactionAttemptContext.Remove(doc)

	span.(*Span).err = err

	defer span.End()

	return err
}

// Unwrap returns the original *gocb.TransactionAttemptContext instance.
// Note: It is not advisable to use this directly, as Instana tracing will not be enabled if you directly utilize this instance.
func (itac *instaTransactionAttemptContext) Unwrap() *gocb.TransactionAttemptContext {
	return itac.TransactionAttemptContext
}

// Helper functions

// createTransactionAttemptContext will wrap *gocb.TransactionAttemptContext in to instaTransactionAttemptContext and will return it as TransactionAttemptContext interface
func createTransactionAttemptContext(tracer requestTracer, ctx *gocb.TransactionAttemptContext, parentSpan gocb.RequestSpan) TransactionAttemptContext {
	return &instaTransactionAttemptContext{
		iTracer:                   tracer,
		TransactionAttemptContext: ctx,

		parentSpan: parentSpan,
	}
}
