// +build go1.11

package storage

import (
	"context"

	"cloud.google.com/go/storage"
	"github.com/instana/go-sensor/instrumentation/cloud.google.com/go/internal"
	ot "github.com/opentracing/opentracing-go"
)

// CopierFrom returns an instrumented cloud.google.com/go/storage.Copier.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#ObjectHandle.CopierFrom for furter details on wrapped method.
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
// that traces calls made to Google Cloud Storage API.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#Copier for furter details on wrapped type.
type Copier struct {
	*storage.Copier
	SourceBucket, SourceName           string
	DestinationBucket, DestinationName string
}

// Run calls and traces the Run() method of the wrapped Copier.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#Copier.Run for furter details on wrapped method.
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

// ComposerFrom returns an instrumented cloud.google.com/go/storage.Composer.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#ObjectHandle.ComposerFrom for furter details on wrapped method.
func (dst *ObjectHandle) ComposerFrom(srcs ...*ObjectHandle) *Composer {
	srcsCopy := make([]*storage.ObjectHandle, len(srcs))
	for i := range srcs {
		srcsCopy[i] = srcs[i].ObjectHandle
	}

	return &Composer{
		Composer:          dst.ObjectHandle.ComposerFrom(srcsCopy...),
		DestinationBucket: dst.Bucket,
		DestinationName:   dst.Name,
	}
}

// Composer is an instrumented wrapper for cloud.google.com/go/storage.Composer.
// that traces calls made to Google Cloud Storage API
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#Composer for furter details on wrapped type.
type Composer struct {
	*storage.Composer
	DestinationBucket, DestinationName string
}

// Run calls and traces the Run() method of the wrapped cloud.google.com/go/storage.Composer.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#Composer.Run for furter details on wrapped method.
func (c *Composer) Run(ctx context.Context) (attrs *storage.ObjectAttrs, err error) {
	ctx = internal.StartExitSpan(ctx, "gcs", ot.Tags{
		"gcs.op":                "objects.compose",
		"gcs.destinationBucket": c.DestinationBucket,
		"gcs.destinationObject": c.DestinationName,
	})

	defer func() { internal.FinishSpan(ctx, err) }()

	return c.Composer.Run(ctx)
}
