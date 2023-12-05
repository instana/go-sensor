// (c) Copyright IBM Corp. 2023

package instagocb

import (
	"github.com/couchbase/gocb/v2"
)

type BinaryCollection interface {
	Append(id string, val []byte, opts *gocb.AppendOptions) (mutOut *gocb.MutationResult, errOut error)
	Prepend(id string, val []byte, opts *gocb.PrependOptions) (mutOut *gocb.MutationResult, errOut error)
	Increment(id string, opts *gocb.IncrementOptions) (countOut *gocb.CounterResult, errOut error)
	Decrement(id string, opts *gocb.DecrementOptions) (countOut *gocb.CounterResult, errOut error)

	Unwrap() *gocb.BinaryCollection
}

type instaBinaryCollection struct {
	*gocb.BinaryCollection
	iTracer gocb.RequestTracer

	// *gocb.BinaryCollection.collection is not accessible as it is private to gocb.
	// Need this for getting bucket in the methods.
	collection Collection
}

// Append appends a byte value to a document.
func (ibc *instaBinaryCollection) Append(id string, val []byte, opts *gocb.AppendOptions) (mutOut *gocb.MutationResult, errOut error) {
	var tracectx gocb.RequestSpanContext
	if opts.ParentSpan != nil {
		tracectx = opts.ParentSpan.Context()
	}

	span := ibc.iTracer.RequestSpan(tracectx, "APPEND")
	span.SetAttribute(bucketNameSpanTag, ibc.collection.Bucket().Name())

	// calling the original Append
	mutOut, errOut = ibc.BinaryCollection.Append(id, val, opts)

	// setting error to span
	span.(*Span).err = errOut

	defer span.End()
	return
}

// Prepend prepends a byte value to a document.
func (ibc *instaBinaryCollection) Prepend(id string, val []byte, opts *gocb.PrependOptions) (mutOut *gocb.MutationResult, errOut error) {
	var tracectx gocb.RequestSpanContext
	if opts.ParentSpan != nil {
		tracectx = opts.ParentSpan.Context()
	}

	span := ibc.iTracer.RequestSpan(tracectx, "PREPEND")
	span.SetAttribute(bucketNameSpanTag, ibc.collection.Bucket().Name())

	mutOut, errOut = ibc.BinaryCollection.Prepend(id, val, opts)

	span.(*Span).err = errOut

	defer span.End()
	return
}

// Increment performs an atomic addition for an integer document. Passing a
// non-negative `initial` value will cause the document to be created if it did not
// already exist.
func (ibc *instaBinaryCollection) Increment(id string, opts *gocb.IncrementOptions) (countOut *gocb.CounterResult, errOut error) {
	var tracectx gocb.RequestSpanContext
	if opts.ParentSpan != nil {
		tracectx = opts.ParentSpan.Context()
	}

	span := ibc.iTracer.RequestSpan(tracectx, "INCREMENT")
	span.SetAttribute(bucketNameSpanTag, ibc.collection.Bucket().Name())

	countOut, errOut = ibc.BinaryCollection.Increment(id, opts)

	span.(*Span).err = errOut

	defer span.End()
	return
}

// Decrement performs an atomic subtraction for an integer document. Passing a
// non-negative `initial` value will cause the document to be created if it did not
// already exist.
func (ibc *instaBinaryCollection) Decrement(id string, opts *gocb.DecrementOptions) (countOut *gocb.CounterResult, errOut error) {
	var tracectx gocb.RequestSpanContext
	if opts.ParentSpan != nil {
		tracectx = opts.ParentSpan.Context()
	}

	span := ibc.iTracer.RequestSpan(tracectx, "DECREMENT")
	span.SetAttribute(bucketNameSpanTag, ibc.collection.Bucket().Name())

	countOut, errOut = ibc.BinaryCollection.Decrement(id, opts)

	span.(*Span).err = errOut

	defer span.End()
	return
}

// Unwrap returns the original *gocb.BinaryCollection instance.
// Note: It is not advisable to use this directly, as Instana tracing will not be enabled if you directly utilize this instance.
func (ibc *instaBinaryCollection) Unwrap() *gocb.BinaryCollection {
	return ibc.BinaryCollection
}

// helper functions

// createBinaryCollection creates an instance of gocb.BinaryCollection and returns it as a BinaryCollection interface
func createBinaryCollection(ic *instaCollection) BinaryCollection {

	// creating a gocb.BinaryCollection object.
	bc := ic.Collection.Binary()

	return &instaBinaryCollection{
		iTracer:          ic.iTracer,
		BinaryCollection: bc,

		collection: ic,
	}
}
