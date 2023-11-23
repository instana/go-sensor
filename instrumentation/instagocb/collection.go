// (c) Copyright IBM Corp. 2023

package instagocb

import (
	"time"

	"github.com/couchbase/gocb/v2"
)

type Collection interface {
	Bucket() Bucket
	Name() string
	QueryIndexes() *gocb.CollectionQueryIndexManager

	Do(ops []gocb.BulkOp, opts *gocb.BulkOpOptions) error

	Insert(id string, val interface{}, opts *gocb.InsertOptions) (mutOut *gocb.MutationResult, errOut error)
	Upsert(id string, val interface{}, opts *gocb.UpsertOptions) (mutOut *gocb.MutationResult, errOut error)
	Replace(id string, val interface{}, opts *gocb.ReplaceOptions) (mutOut *gocb.MutationResult, errOut error)
	Get(id string, opts *gocb.GetOptions) (docOut *gocb.GetResult, errOut error)
	Exists(id string, opts *gocb.ExistsOptions) (docOut *gocb.ExistsResult, errOut error)
	GetAllReplicas(id string, opts *gocb.GetAllReplicaOptions) (docOut *gocb.GetAllReplicasResult, errOut error)
	GetAnyReplica(id string, opts *gocb.GetAnyReplicaOptions) (docOut *gocb.GetReplicaResult, errOut error)
	Remove(id string, opts *gocb.RemoveOptions) (mutOut *gocb.MutationResult, errOut error)
	GetAndTouch(id string, expiry time.Duration, opts *gocb.GetAndTouchOptions) (docOut *gocb.GetResult, errOut error)
	GetAndLock(id string, lockTime time.Duration, opts *gocb.GetAndLockOptions) (docOut *gocb.GetResult, errOut error)
	Unlock(id string, cas gocb.Cas, opts *gocb.UnlockOptions) (errOut error)
	Touch(id string, expiry time.Duration, opts *gocb.TouchOptions) (mutOut *gocb.MutationResult, errOut error)
	Binary() BinaryCollection

	List(id string) CouchbaseList
	Map(id string) CouchbaseMap
	Set(id string) CouchbaseSet
	Queue(id string) CouchbaseQueue

	LookupIn(id string, ops []gocb.LookupInSpec, opts *gocb.LookupInOptions) (docOut *gocb.LookupInResult, errOut error)
	MutateIn(id string, ops []gocb.MutateInSpec, opts *gocb.MutateInOptions) (mutOut *gocb.MutateInResult, errOut error)
	ScopeName() string

	Unwrap() *gocb.Collection
}

type instaCollection struct {
	*gocb.Collection
	iTracer gocb.RequestTracer
}

// Bucket returns the bucket to which this collection belongs.
func (ic *instaCollection) Bucket() Bucket {
	bucket := ic.Collection.Bucket()
	return createBucket(ic.iTracer, bucket)
}

// Insert creates a new document in the Collection.
func (ic *instaCollection) Insert(id string, val interface{}, opts *gocb.InsertOptions) (mutOut *gocb.MutationResult, errOut error) {
	var tracectx gocb.RequestSpanContext
	if opts.ParentSpan != nil {
		tracectx = opts.ParentSpan.Context()
	}

	span := ic.iTracer.RequestSpan(tracectx, "INSERT")
	span.SetAttribute(bucketNameSpanTag, ic.Bucket().Name())

	// calling the original Insert
	mutOut, errOut = ic.Collection.Insert(id, val, opts)

	// setting error to span
	span.(*Span).err = errOut

	defer span.End()
	return
}

// Upsert creates a new document in the Collection if it does not exist, if it does exist then it updates it.
func (ic *instaCollection) Upsert(id string, val interface{}, opts *gocb.UpsertOptions) (mutOut *gocb.MutationResult, errOut error) {
	var tracectx gocb.RequestSpanContext
	if opts.ParentSpan != nil {
		tracectx = opts.ParentSpan.Context()
	}

	span := ic.iTracer.RequestSpan(tracectx, "UPSERT")
	span.SetAttribute(bucketNameSpanTag, ic.Bucket().Name())

	// calling the original Upsert
	mutOut, errOut = ic.Collection.Upsert(id, val, opts)

	// setting error to span
	span.(*Span).err = errOut

	defer span.End()
	return
}

// Replace updates a document in the collection.
func (ic *instaCollection) Replace(id string, val interface{}, opts *gocb.ReplaceOptions) (mutOut *gocb.MutationResult, errOut error) {
	var tracectx gocb.RequestSpanContext
	if opts.ParentSpan != nil {
		tracectx = opts.ParentSpan.Context()
	}

	span := ic.iTracer.RequestSpan(tracectx, "REPLACE")
	span.SetAttribute(bucketNameSpanTag, ic.Bucket().Name())

	mutOut, errOut = ic.Collection.Replace(id, val, opts)

	span.(*Span).err = errOut

	defer span.End()
	return
}

// Get performs a fetch operation against the collection. This can take 3 paths, a standard full document
// fetch, a subdocument full document fetch also fetching document expiry (when WithExpiry is set),
// or a subdocument fetch (when Project is used).
func (ic *instaCollection) Get(id string, opts *gocb.GetOptions) (docOut *gocb.GetResult, errOut error) {
	var tracectx gocb.RequestSpanContext
	if opts.ParentSpan != nil {
		tracectx = opts.ParentSpan.Context()
	}

	span := ic.iTracer.RequestSpan(tracectx, "GET")
	span.SetAttribute(bucketNameSpanTag, ic.Bucket().Name())

	// calling the original Get
	docOut, errOut = ic.Collection.Get(id, opts)

	// setting error to span
	span.(*Span).err = errOut

	defer span.End()
	return
}

// Exists checks if a document exists for the given id.
func (ic *instaCollection) Exists(id string, opts *gocb.ExistsOptions) (docOut *gocb.ExistsResult, errOut error) {
	var tracectx gocb.RequestSpanContext
	if opts.ParentSpan != nil {
		tracectx = opts.ParentSpan.Context()
	}

	span := ic.iTracer.RequestSpan(tracectx, "EXISTS")
	span.SetAttribute(bucketNameSpanTag, ic.Bucket().Name())

	docOut, errOut = ic.Collection.Exists(id, opts)

	span.(*Span).err = errOut

	defer span.End()
	return
}

// GetAllReplicas returns the value of a particular document from all replica servers. This will return an iterable
// which streams results one at a time.
func (ic *instaCollection) GetAllReplicas(id string, opts *gocb.GetAllReplicaOptions) (docOut *gocb.GetAllReplicasResult, errOut error) {
	var tracectx gocb.RequestSpanContext
	if opts.ParentSpan != nil {
		tracectx = opts.ParentSpan.Context()
	}

	span := ic.iTracer.RequestSpan(tracectx, "GET_ALL_REPLICAS")
	span.SetAttribute(bucketNameSpanTag, ic.Bucket().Name())

	docOut, errOut = ic.Collection.GetAllReplicas(id, opts)

	span.(*Span).err = errOut

	defer span.End()
	return
}

// GetAnyReplica returns the value of a particular document from a replica server.
func (ic *instaCollection) GetAnyReplica(id string, opts *gocb.GetAnyReplicaOptions) (docOut *gocb.GetReplicaResult, errOut error) {
	var tracectx gocb.RequestSpanContext
	if opts.ParentSpan != nil {
		tracectx = opts.ParentSpan.Context()
	}

	span := ic.iTracer.RequestSpan(tracectx, "GET_ANY_REPLICA")
	span.SetAttribute(bucketNameSpanTag, ic.Bucket().Name())

	docOut, errOut = ic.Collection.GetAnyReplica(id, opts)

	span.(*Span).err = errOut

	defer span.End()
	return
}

// Remove removes a document from the collection.
func (ic *instaCollection) Remove(id string, opts *gocb.RemoveOptions) (mutOut *gocb.MutationResult, errOut error) {
	var tracectx gocb.RequestSpanContext
	if opts.ParentSpan != nil {
		tracectx = opts.ParentSpan.Context()
	}

	span := ic.iTracer.RequestSpan(tracectx, "REMOVE")
	span.SetAttribute(bucketNameSpanTag, ic.Bucket().Name())

	mutOut, errOut = ic.Collection.Remove(id, opts)

	span.(*Span).err = errOut

	defer span.End()
	return
}

// GetAndTouch retrieves a document and simultaneously updates its expiry time.
func (ic *instaCollection) GetAndTouch(id string, expiry time.Duration, opts *gocb.GetAndTouchOptions) (docOut *gocb.GetResult, errOut error) {
	var tracectx gocb.RequestSpanContext
	if opts.ParentSpan != nil {
		tracectx = opts.ParentSpan.Context()
	}

	span := ic.iTracer.RequestSpan(tracectx, "GET_AND_TOUCH")
	span.SetAttribute(bucketNameSpanTag, ic.Bucket().Name())

	docOut, errOut = ic.Collection.GetAndTouch(id, expiry, opts)

	span.(*Span).err = errOut

	defer span.End()
	return
}

// GetAndLock locks a document for a period of time, providing exclusive RW access to it.
// A lockTime value of over 30 seconds will be treated as 30 seconds. The resolution used to send this value to
// the server is seconds and is calculated using uint32(lockTime/time.Second).
func (ic *instaCollection) GetAndLock(id string, lockTime time.Duration, opts *gocb.GetAndLockOptions) (docOut *gocb.GetResult, errOut error) {
	var tracectx gocb.RequestSpanContext
	if opts.ParentSpan != nil {
		tracectx = opts.ParentSpan.Context()
	}

	span := ic.iTracer.RequestSpan(tracectx, "GET_AND_LOCK")
	span.SetAttribute(bucketNameSpanTag, ic.Bucket().Name())

	docOut, errOut = ic.Collection.GetAndLock(id, lockTime, opts)

	span.(*Span).err = errOut

	defer span.End()
	return
}

// Unlock unlocks a document which was locked with GetAndLock.
func (ic *instaCollection) Unlock(id string, cas gocb.Cas, opts *gocb.UnlockOptions) (errOut error) {
	var tracectx gocb.RequestSpanContext
	if opts.ParentSpan != nil {
		tracectx = opts.ParentSpan.Context()
	}

	span := ic.iTracer.RequestSpan(tracectx, "UNLOCK")
	span.SetAttribute(bucketNameSpanTag, ic.Bucket().Name())

	errOut = ic.Collection.Unlock(id, cas, opts)

	span.(*Span).err = errOut

	defer span.End()
	return
}

// Touch touches a document, specifying a new expiry time for it.
func (ic *instaCollection) Touch(id string, expiry time.Duration, opts *gocb.TouchOptions) (mutOut *gocb.MutationResult, errOut error) {
	var tracectx gocb.RequestSpanContext
	if opts.ParentSpan != nil {
		tracectx = opts.ParentSpan.Context()
	}

	span := ic.iTracer.RequestSpan(tracectx, "TOUCH")
	span.SetAttribute(bucketNameSpanTag, ic.Bucket().Name())

	mutOut, errOut = ic.Collection.Touch(id, expiry, opts)

	span.(*Span).err = errOut

	defer span.End()
	return
}

// Binary creates and returns a BinaryCollection.
func (ic *instaCollection) Binary() BinaryCollection {
	return createBinaryCollection(ic)
}

// LookupIn performs a set of subdocument lookup operations on the document identified by id.
func (ic *instaCollection) LookupIn(id string, ops []gocb.LookupInSpec, opts *gocb.LookupInOptions) (docOut *gocb.LookupInResult, errOut error) {
	var tracectx gocb.RequestSpanContext
	if opts.ParentSpan != nil {
		tracectx = opts.ParentSpan.Context()
	}

	span := ic.iTracer.RequestSpan(tracectx, "LOOKUP_IN")
	span.SetAttribute(bucketNameSpanTag, ic.Bucket().Name())

	docOut, errOut = ic.Collection.LookupIn(id, ops, opts)

	span.(*Span).err = errOut

	defer span.End()
	return
}

// MutateIn performs a set of subdocument mutations on the document specified by id.
func (ic *instaCollection) MutateIn(id string, ops []gocb.MutateInSpec, opts *gocb.MutateInOptions) (mutOut *gocb.MutateInResult, errOut error) {
	var tracectx gocb.RequestSpanContext
	if opts.ParentSpan != nil {
		tracectx = opts.ParentSpan.Context()
	}

	span := ic.iTracer.RequestSpan(tracectx, "MUTATE_IN")
	span.SetAttribute(bucketNameSpanTag, ic.Bucket().Name())

	mutOut, errOut = ic.Collection.MutateIn(id, ops, opts)

	span.(*Span).err = errOut

	defer span.End()
	return
}

// List returns a new CouchbaseList for the document specified by id.
func (ic *instaCollection) List(id string) CouchbaseList {
	return createList(ic, id)
}

// Map returns a new CouchbaseMap.
func (ic *instaCollection) Map(id string) CouchbaseMap {
	return createMap(ic, id)
}

// Set returns a new CouchbaseSet.
func (ic *instaCollection) Set(id string) CouchbaseSet {
	return createSet(ic, id)
}

// Queue returns a new CouchbaseQueue.
func (ic *instaCollection) Queue(id string) CouchbaseQueue {
	return createQueue(ic, id)
}

// Do execute one or more `BulkOp` items in parallel.
func (ic *instaCollection) Do(ops []gocb.BulkOp, opts *gocb.BulkOpOptions) error {
	var tracectx gocb.RequestSpanContext
	if opts.ParentSpan != nil {
		tracectx = opts.ParentSpan.Context()
	}

	span := ic.iTracer.RequestSpan(tracectx, "BULK")
	span.SetAttribute(bucketNameSpanTag, ic.Bucket().Name())

	errOut := ic.Collection.Do(ops, opts)

	span.(*Span).err = errOut

	defer span.End()
	return errOut
}

// Unwrap returns the original *gocb.Collection instance.
// Note: It is not advisable to use this directly, as Instana tracing will not be enabled if you directly utilize this instance.
func (ic *instaCollection) Unwrap() *gocb.Collection {
	return ic.Collection
}

// helper functions

// createCollection will wrap *gocb.Collection in to instanaCollection and will return it as Collection interface
func createCollection(tracer gocb.RequestTracer, collection *gocb.Collection) Collection {
	return &instaCollection{
		iTracer:    tracer,
		Collection: collection,
	}
}
