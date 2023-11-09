// (c) Copyright IBM Corp. 2023

package instagocb

import (
	"time"

	"github.com/couchbase/gocb/v2"
)

type Bucket interface {
	Name() string
	Scope(scopeName string) Scope
	DefaultScope() *gocb.Scope
	Collection(collectionName string) *gocb.Collection
	DefaultCollection() *gocb.Collection
	ViewIndexes() *gocb.ViewIndexManager
	Collections() *gocb.CollectionManager
	WaitUntilReady(timeout time.Duration, opts *gocb.WaitUntilReadyOptions) error
}

type InstanaBucket struct {
	iTracer gocb.RequestTracer
	*gocb.Bucket
}

func (ib *InstanaBucket) Scope(s string) Scope {
	scope := ib.Bucket.Scope(s)

	return &InstanaScope{
		iTracer: ib.iTracer,
		Scope:   scope,
	}
}
