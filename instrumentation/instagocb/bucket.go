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
	Collections() *gocb.CollectionManager
	WaitUntilReady(timeout time.Duration, opts *gocb.WaitUntilReadyOptions) error

	// view query
	ViewQuery(designDoc string, viewName string, opts *gocb.ViewOptions) (*gocb.ViewResult, error)

	// ping
	Ping(opts *gocb.PingOptions) (*gocb.PingResult, error)

	// internal
	Internal() *gocb.InternalBucket
}

type InstanaBucket struct {
	iTracer gocb.RequestTracer
	*gocb.Bucket
}

// Scope returns an instance of a Scope.
func (ib *InstanaBucket) Scope(s string) Scope {
	scope := ib.Bucket.Scope(s)
	return createScope(ib.iTracer, scope)
}

// DefaultScope returns an instance of the default scope.
func (ib *InstanaBucket) DefaultScope() Scope {
	ds := ib.Bucket.DefaultScope()
	return createScope(ib.iTracer, ds)
}

// Collection returns an instance of a collection from within the default scope.
func (ib *InstanaBucket) Collection(collectionName string) Collection {
	collection := ib.Bucket.Collection(collectionName)
	return createCollection(ib.iTracer, collection)
}

// DefaultCollection returns an instance of the default collection.
func (ib *InstanaBucket) DefaultCollection() Collection {
	dc := ib.Bucket.DefaultCollection()
	return createCollection(ib.iTracer, dc)
}

// helper functions

// createCollection will wrap *gocb.Collection in to instanaCollection and will return it as Collection interface
func createBucket(tracer gocb.RequestTracer, bucket *gocb.Bucket) Bucket {
	return &InstanaBucket{
		iTracer: tracer,
		Bucket:  bucket,
	}
}
