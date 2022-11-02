// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/instana/go-sensor/instrumentation/cloud.google.com/go/internal/tags"

	"cloud.google.com/go/storage"
	"github.com/instana/go-sensor/instrumentation/cloud.google.com/go/internal"
	ot "github.com/opentracing/opentracing-go"
)

// NewReader returns an instrumented wrapper for cloud.google.com/go/storage.Reader for an object.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#ObjectHandle.NewReader for further details on wrapped method.
func (o *ObjectHandle) NewReader(ctx context.Context) (*Reader, error) {
	return o.NewRangeReader(ctx, 0, -1)
}

// NewRangeReader returns an instrumented wrapper for cloud.google.com/go/storage.Reader that reads
// the object partially.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#ObjectHandle.NewRangeReader for further details on wrapped method.
func (o *ObjectHandle) NewRangeReader(ctx context.Context, offset, length int64) (r *Reader, err error) {
	attrsCtx := internal.StartExitSpan(ctx, "gcs", ot.Tags{
		tags.GcsOp:     "objects.get",
		tags.GcsBucket: o.Bucket,
		tags.GcsObject: o.Name,
	})
	defer func() { internal.FinishSpan(attrsCtx, err) }()

	rdr, err := o.ObjectHandle.NewRangeReader(ctx, offset, length)
	return &Reader{
		Reader: rdr,
		ctx:    ctx,
		Bucket: o.Bucket,
		Name:   o.Name,
	}, err
}

// Reader is an instrumented wrapper for cloud.google.com/go/storage.Reader
// that traces calls made to Google Cloud Storage API.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#Reader for further details on wrapped type.
type Reader struct {
	*storage.Reader
	ctx context.Context

	Bucket string
	Name   string
}

// Read calls and traces the Read() method of the wrapped cloud.google.com/go.Reader.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#Reader.Read for further details on wrapped method.
func (r *Reader) Read(p []byte) (n int, err error) {
	tags := ot.Tags{
		tags.GcsOp:     "objects.get",
		tags.GcsBucket: r.Bucket,
		tags.GcsObject: r.Name,
		tags.GcsRange:  fmt.Sprintf("%d-%d", r.Attrs.StartOffset, r.Attrs.Size),
	}

	ctx := internal.StartExitSpan(r.ctx, "gcs", tags)
	defer func() {
		if err == io.EOF {
			internal.FinishSpan(ctx, nil)
			return
		}

		internal.FinishSpan(ctx, err)
	}()

	return r.Reader.Read(p)
}
