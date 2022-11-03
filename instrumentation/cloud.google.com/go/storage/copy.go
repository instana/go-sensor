// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package storage

import (
	"context"
	"strings"

	"github.com/instana/go-sensor/instrumentation/cloud.google.com/go/internal/tags"

	"cloud.google.com/go/storage"
	"github.com/instana/go-sensor/instrumentation/cloud.google.com/go/internal"
	ot "github.com/opentracing/opentracing-go"
)

// CopierFrom returns an instrumented cloud.google.com/go/storage.Copier.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#ObjectHandle.CopierFrom for further details on wrapped method.
func (o *ObjectHandle) CopierFrom(src *ObjectHandle) *Copier {
	return &Copier{
		Copier:            o.ObjectHandle.CopierFrom(src.ObjectHandle),
		SourceBucket:      src.Bucket,
		SourceName:        src.Name,
		DestinationBucket: o.Bucket,
		DestinationName:   o.Name,
	}
}

// Copier is an instrumented wrapper for cloud.google.com/go/storage.Copier
// that traces calls made to Google Cloud Storage API.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#Copier for further details on wrapped type.
type Copier struct {
	*storage.Copier
	SourceBucket, SourceName           string
	DestinationBucket, DestinationName string
}

// Run calls and traces the Run() method of the wrapped Copier.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#Copier.Run for further details on wrapped method.
func (c *Copier) Run(ctx context.Context) (attrs *storage.ObjectAttrs, err error) {
	ctx = internal.StartExitSpan(ctx, "gcs", ot.Tags{
		tags.GcsOp:                "objects.copy",
		tags.GcsSourceBucket:      c.SourceBucket,
		tags.GcsSourceObject:      c.SourceName,
		tags.GcsDestinationBucket: c.DestinationBucket,
		tags.GcsDestinationObject: c.DestinationName,
	})

	defer func() { internal.FinishSpan(ctx, err) }()

	return c.Copier.Run(ctx)
}

// ComposerFrom returns an instrumented cloud.google.com/go/storage.Composer.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#ObjectHandle.ComposerFrom for further details on wrapped method.
func (o *ObjectHandle) ComposerFrom(srcs ...*ObjectHandle) *Composer {
	srcsCopy := make([]*storage.ObjectHandle, len(srcs))
	sourceObjects := make([]string, len(srcs))
	for i := range srcs {
		srcsCopy[i] = srcs[i].ObjectHandle
		sourceObjects[i] = srcs[i].Bucket + "/" + srcs[i].Name
	}

	return &Composer{
		Composer:          o.ObjectHandle.ComposerFrom(srcsCopy...),
		DestinationBucket: o.Bucket,
		DestinationName:   o.Name,
		SourceObjects:     sourceObjects,
	}
}

// Composer is an instrumented wrapper for cloud.google.com/go/storage.Composer.
// that traces calls made to Google Cloud Storage API
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#Composer for further details on wrapped type.
type Composer struct {
	*storage.Composer
	DestinationBucket, DestinationName string
	SourceObjects                      []string
}

// Run calls and traces the Run() method of the wrapped cloud.google.com/go/storage.Composer.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#Composer.Run for further details on wrapped method.
func (c *Composer) Run(ctx context.Context) (attrs *storage.ObjectAttrs, err error) {
	ctx = internal.StartExitSpan(ctx, "gcs", ot.Tags{
		tags.GcsOp:                "objects.compose",
		tags.GcsDestinationBucket: c.DestinationBucket,
		tags.GcsDestinationObject: c.DestinationName,
		tags.GcsSourceObjects:     strings.Join(c.SourceObjects, ","),
	})

	defer func() { internal.FinishSpan(ctx, err) }()

	return c.Composer.Run(ctx)
}
