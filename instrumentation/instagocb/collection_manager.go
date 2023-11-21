// (c) Copyright IBM Corp. 2023

package instagocb

import (
	"github.com/couchbase/gocb/v2"
)

type CollectionManager interface {
	GetAllScopes(opts *gocb.GetAllScopesOptions) ([]gocb.ScopeSpec, error)
	CreateCollection(spec gocb.CollectionSpec, opts *gocb.CreateCollectionOptions) error
	DropCollection(spec gocb.CollectionSpec, opts *gocb.DropCollectionOptions) error
	CreateScope(scopeName string, opts *gocb.CreateScopeOptions) error
	DropScope(scopeName string, opts *gocb.DropScopeOptions) error

	Unwrap() *gocb.CollectionManager
}

type instaCollectionManager struct {
	iTracer gocb.RequestTracer
	*gocb.CollectionManager

	bucketName string
}

// CreateCollection creates a new collection on the bucket.
func (icm *instaCollectionManager) CreateCollection(spec gocb.CollectionSpec, opts *gocb.CreateCollectionOptions) error {
	var tracectx gocb.RequestSpanContext
	if opts.ParentSpan != nil {
		tracectx = opts.ParentSpan.Context()
	}

	span := icm.iTracer.RequestSpan(tracectx, "CREATE_COLLECTION")
	span.SetAttribute(bucketNameSpanTag, icm.bucketName)

	// calling the original CreateCollection
	errOut := icm.CollectionManager.CreateCollection(spec, opts)

	// setting error to span
	span.(*Span).err = errOut

	defer span.End()
	return errOut
}

// DropCollection removes a collection.
func (icm *instaCollectionManager) DropCollection(spec gocb.CollectionSpec, opts *gocb.DropCollectionOptions) error {
	var tracectx gocb.RequestSpanContext
	if opts.ParentSpan != nil {
		tracectx = opts.ParentSpan.Context()
	}

	span := icm.iTracer.RequestSpan(tracectx, "DROP_COLLECTION")
	span.SetAttribute(bucketNameSpanTag, icm.bucketName)

	errOut := icm.CollectionManager.DropCollection(spec, opts)

	span.(*Span).err = errOut

	defer span.End()
	return errOut
}

// CreateScope creates a new scope on the bucket.
func (icm *instaCollectionManager) CreateScope(scopeName string, opts *gocb.CreateScopeOptions) error {
	var tracectx gocb.RequestSpanContext
	if opts.ParentSpan != nil {
		tracectx = opts.ParentSpan.Context()
	}

	span := icm.iTracer.RequestSpan(tracectx, "CREATE_SCOPE")
	span.SetAttribute(bucketNameSpanTag, icm.bucketName)

	errOut := icm.CollectionManager.CreateScope(scopeName, opts)

	span.(*Span).err = errOut

	defer span.End()
	return errOut
}

// DropScope removes a scope.
func (icm *instaCollectionManager) DropScope(scopeName string, opts *gocb.DropScopeOptions) error {
	var tracectx gocb.RequestSpanContext
	if opts.ParentSpan != nil {
		tracectx = opts.ParentSpan.Context()
	}

	span := icm.iTracer.RequestSpan(tracectx, "DROP_SCOPE")
	span.SetAttribute(bucketNameSpanTag, icm.bucketName)

	errOut := icm.CollectionManager.DropScope(scopeName, opts)

	span.(*Span).err = errOut

	defer span.End()
	return errOut
}

// Unwrap returns the original *gocb.CollectionManager instance.
// Note: It is not advisable to use this directly, as Instana tracing will not be enabled if you directly utilize this instance.
func (icm *instaCollectionManager) Unwrap() *gocb.CollectionManager {
	return icm.CollectionManager
}

// helper functions

// createCollectionManager will wrap *gocb.CollectionManager in to instaCollectionManager and will return it as CollectionManager interface
func createCollectionManager(tracer gocb.RequestTracer, cm *gocb.CollectionManager, bucketName string) CollectionManager {
	return &instaCollectionManager{
		iTracer:           tracer,
		CollectionManager: cm,

		bucketName: bucketName,
	}
}
