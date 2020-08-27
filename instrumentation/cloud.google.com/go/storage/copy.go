package storage

import (
	"context"

	"cloud.google.com/go/storage"
	"github.com/instana/go-sensor/instrumentation/cloud.google.com/go/internal"
	ot "github.com/opentracing/opentracing-go"
)

// CopierFrom returns an instrumented cloud.google.com/go/storage.Copier
func (dst *ObjectHandle) CopierFrom(src *ObjectHandle) *Copier {
	return &Copier{
		Copier:            dst.ObjectHandle.CopierFrom(src.ObjectHandle),
		SourceBucket:      src.Bucket,
		SourceName:        src.Name,
		DestinationBucket: dst.Bucket,
		DestinationName:   dst.Name,
	}
}

// Copier is an instrumented wrapper for cloud.google.com/go/storage.Copier
// that traces calls made to Google Cloud Storage API
type Copier struct {
	*storage.Copier
	SourceBucket, SourceName           string
	DestinationBucket, DestinationName string
}

// Run calls and traces the Run() method of the wrapped Copier
func (c *Copier) Run(ctx context.Context) (attrs *storage.ObjectAttrs, err error) {
	ctx = internal.StartExitSpan(ctx, "gcs", ot.Tags{
		"gcs.op":                "objects.copy",
		"gcs.sourceBucket":      c.SourceBucket,
		"gcs.sourceObject":      c.SourceName,
		"gcs.destinationBucket": c.DestinationBucket,
		"gcs.destinationObject": c.DestinationName,
	})

	defer func() { internal.FinishSpan(ctx, err) }()

	return c.Copier.Run(ctx)
}

// ComposerFrom creates a Composer that can compose srcs into dst.
// You can immediately call Run on the returned Composer, or you can
// configure it first.
//
// The encryption key for the destination object will be used to decrypt all
// source objects and encrypt the destination object. It is an error
// to specify an encryption key for any of the source objects.
func (dst *ObjectHandle) ComposerFrom(srcs ...*ObjectHandle) *Composer {
	srcsCopy := make([]*storage.ObjectHandle, len(srcs))
	for i := range srcs {
		srcsCopy[i] = srcs[i].ObjectHandle
	}

	return &Composer{dst.ObjectHandle.ComposerFrom(srcsCopy...)}
}

// A Composer composes source objects into a destination object.
//
// For Requester Pays buckets, the user project of dst is billed.
type Composer struct {
	*storage.Composer
}

// Run performs the compose operation.
//
// INSTRUMENT
func (c *Composer) Run(ctx context.Context) (attrs *storage.ObjectAttrs, err error) {
	return c.Composer.Run(ctx)
}
