// (c) Copyright IBM Corp. 2023

package instagocb

import (
	"github.com/couchbase/gocb/v2"
)

type BucketManager interface {
	GetBucket(bucketName string, opts *gocb.GetBucketOptions) (*gocb.BucketSettings, error)
	GetAllBuckets(opts *gocb.GetAllBucketsOptions) (map[string]gocb.BucketSettings, error)
	CreateBucket(settings gocb.CreateBucketSettings, opts *gocb.CreateBucketOptions) error
	UpdateBucket(settings gocb.BucketSettings, opts *gocb.UpdateBucketOptions) error
	DropBucket(name string, opts *gocb.DropBucketOptions) error
	FlushBucket(name string, opts *gocb.FlushBucketOptions) error
}

type InstanaBucketManager struct {
	iTracer gocb.RequestTracer
	*gocb.BucketManager
}

// GetBucket returns settings for a bucket on the cluster.
func (ibm *InstanaBucketManager) GetBucket(bucketName string, opts *gocb.GetBucketOptions) (*gocb.BucketSettings, error) {
	var tracectx gocb.RequestSpanContext
	if opts.ParentSpan != nil {
		tracectx = opts.ParentSpan.Context()
	}

	span := ibm.iTracer.RequestSpan(tracectx, "GET_BUCKET")
	span.SetAttribute(bucketNameSpanTag, bucketName)

	// calling the original GetBucket
	res, errOut := ibm.BucketManager.GetBucket(bucketName, opts)

	// setting error to span
	span.(*Span).err = errOut

	defer span.End()
	return res, errOut
}

// CreateBucket creates a bucket on the cluster.
func (ibm *InstanaBucketManager) CreateBucket(settings gocb.CreateBucketSettings, opts *gocb.CreateBucketOptions) error {
	var tracectx gocb.RequestSpanContext
	if opts.ParentSpan != nil {
		tracectx = opts.ParentSpan.Context()
	}

	span := ibm.iTracer.RequestSpan(tracectx, "CREATE_BUCKET")
	span.SetAttribute(bucketNameSpanTag, settings.Name)
	span.SetAttribute(bucketTypeSpanTag, string(settings.BucketType))

	errOut := ibm.BucketManager.CreateBucket(settings, opts)

	span.(*Span).err = errOut

	defer span.End()
	return errOut
}

// UpdateBucket updates a bucket on the cluster.
func (ibm *InstanaBucketManager) UpdateBucket(settings gocb.BucketSettings, opts *gocb.UpdateBucketOptions) error {
	var tracectx gocb.RequestSpanContext
	if opts.ParentSpan != nil {
		tracectx = opts.ParentSpan.Context()
	}

	span := ibm.iTracer.RequestSpan(tracectx, "UPDATE_BUCKET")
	span.SetAttribute(bucketNameSpanTag, settings.Name)

	errOut := ibm.BucketManager.UpdateBucket(settings, opts)

	span.(*Span).err = errOut

	defer span.End()
	return errOut
}

// DropBucket will delete a bucket from the cluster by name.
func (ibm *InstanaBucketManager) DropBucket(name string, opts *gocb.DropBucketOptions) error {
	var tracectx gocb.RequestSpanContext
	if opts.ParentSpan != nil {
		tracectx = opts.ParentSpan.Context()
	}

	span := ibm.iTracer.RequestSpan(tracectx, "DROP_BUCKET")
	span.SetAttribute(bucketNameSpanTag, name)

	errOut := ibm.BucketManager.DropBucket(name, opts)

	span.(*Span).err = errOut

	defer span.End()
	return errOut
}

// FlushBucket will delete all the of the data from a bucket.
// Keep in mind that you must have flushing enabled in the buckets configuration.
func (ibm *InstanaBucketManager) FlushBucket(name string, opts *gocb.FlushBucketOptions) error {
	var tracectx gocb.RequestSpanContext
	if opts.ParentSpan != nil {
		tracectx = opts.ParentSpan.Context()
	}

	span := ibm.iTracer.RequestSpan(tracectx, "FLUSH_BUCKET")
	span.SetAttribute(bucketNameSpanTag, name)

	errOut := ibm.BucketManager.FlushBucket(name, opts)

	span.(*Span).err = errOut

	defer span.End()
	return errOut
}

// helper functions

// createBucketManager will wrap *gocb.BucketManager in to InstanaBucketManager and will return it as BucketManager interface
func createBucketManager(tracer gocb.RequestTracer, bm *gocb.BucketManager) BucketManager {
	return &InstanaBucketManager{
		iTracer:       tracer,
		BucketManager: bm,
	}
}
