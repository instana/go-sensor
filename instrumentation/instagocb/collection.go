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

	// bulk operation
	Do(ops []gocb.BulkOp, opts *gocb.BulkOpOptions) error

	// crud
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
	Binary() *gocb.BinaryCollection

	// ds
	List(id string) *gocb.CouchbaseList
	Map(id string) *gocb.CouchbaseMap
	Set(id string) *gocb.CouchbaseSet
	Queue(id string) *gocb.CouchbaseQueue

	// sub doc
	LookupIn(id string, ops []gocb.LookupInSpec, opts *gocb.LookupInOptions) (docOut *gocb.LookupInResult, errOut error)
	MutateIn(id string, ops []gocb.MutateInSpec, opts *gocb.MutateInOptions) (mutOut *gocb.MutateInResult, errOut error)
	ScopeName() string
}

type InstanaCollection struct {
	*gocb.Collection
	iTracer gocb.RequestTracer
}

// Bucket returns the bucket to which this collection belongs.
func (ic *InstanaCollection) Bucket() Bucket {
	bucket := ic.Collection.Bucket()
	return createBucket(ic.iTracer, bucket)
}

// Upsert creates a new document in the Collection if it does not exist, if it does exist then it updates it.
func (ic *InstanaCollection) Upsert(id string, val interface{}, opts *gocb.UpsertOptions) (mutOut *gocb.MutationResult, errOut error) {
	span := ic.iTracer.RequestSpan(opts.ParentSpan.Context(), "UPSERT")
	span.SetAttribute(bucketNameSpanTag, ic.Bucket().Name())

	// calling the original Upsert
	mutOut, errOut = ic.Collection.Upsert(id, val, opts)

	// setting error to span
	span.(*Span).err = errOut

	defer span.End()
	return
}

// Get performs a fetch operation against the collection. This can take 3 paths, a standard full document
// fetch, a subdocument full document fetch also fetching document expiry (when WithExpiry is set),
// or a subdocument fetch (when Project is used).
func (ic *InstanaCollection) Get(id string, opts *gocb.GetOptions) (docOut *gocb.GetResult, errOut error) {
	span := ic.iTracer.RequestSpan(opts.ParentSpan.Context(), "GET")
	span.SetAttribute(bucketNameSpanTag, ic.Bucket().Name())

	// calling the original Get
	docOut, errOut = ic.Collection.Get(id, opts)

	// setting error to span
	span.(*Span).err = errOut

	defer span.End()
	return
}

// helper functions

// createCollection will wrap *gocb.Collection in to instanaCollection and will return it as Collection interface
func createCollection(tracer gocb.RequestTracer, collection *gocb.Collection) Collection {
	return &InstanaCollection{
		iTracer:    tracer,
		Collection: collection,
	}
}
