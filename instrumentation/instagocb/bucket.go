// (c) Copyright IBM Corp. 2023

package instagocb

import (
	"time"

	"github.com/couchbase/gocb/v2"
)

type Bucket interface {
	Name() string
	Scope(scopeName string) Scope
	DefaultScope() Scope
	Collection(collectionName string) Collection
	DefaultCollection() Collection
	ViewIndexes() *gocb.ViewIndexManager
	Collections() CollectionManager
	WaitUntilReady(timeout time.Duration, opts *gocb.WaitUntilReadyOptions) error

	ViewQuery(designDoc string, viewName string, opts *gocb.ViewOptions) (*gocb.ViewResult, error)

	Ping(opts *gocb.PingOptions) (*gocb.PingResult, error)

	Internal() *gocb.InternalBucket

	Unwrap() *gocb.Bucket
}

type instaBucket struct {
	iTracer gocb.RequestTracer
	*gocb.Bucket
}

// Scope returns an instance of a Scope.
func (ib *instaBucket) Scope(s string) Scope {
	scope := ib.Bucket.Scope(s)
	return createScope(ib.iTracer, scope)
}

// DefaultScope returns an instance of the default scope.
func (ib *instaBucket) DefaultScope() Scope {
	ds := ib.Bucket.DefaultScope()
	return createScope(ib.iTracer, ds)
}

// Collection returns an instance of a collection from within the default scope.
func (ib *instaBucket) Collection(collectionName string) Collection {
	collection := ib.Bucket.Collection(collectionName)
	return createCollection(ib.iTracer, collection)
}

// DefaultCollection returns an instance of the default collection.
func (ib *instaBucket) DefaultCollection() Collection {
	dc := ib.Bucket.DefaultCollection()
	return createCollection(ib.iTracer, dc)
}

// Collections provides functions for managing collections.
func (ib *instaBucket) Collections() CollectionManager {
	cm := ib.Bucket.Collections()
	return createCollectionManager(ib.iTracer, cm, ib.Name())
}

// Unwrap returns the original *gocb.Bucket instance.
// Note: It is not advisable to use this directly, as Instana tracing will not be enabled if you directly utilize this instance.
func (ib *instaBucket) Unwrap() *gocb.Bucket {
	return ib.Bucket
}

// helper functions

// createBucket will wrap *gocb.Bucket in to instaBucket and will return it as Bucket interface
func createBucket(tracer gocb.RequestTracer, bucket *gocb.Bucket) Bucket {
	return &instaBucket{
		iTracer: tracer,
		Bucket:  bucket,
	}
}
