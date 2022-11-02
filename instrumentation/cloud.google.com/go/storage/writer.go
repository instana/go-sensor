// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package storage

import (
	"context"
	"sync"

	"github.com/instana/go-sensor/instrumentation/cloud.google.com/go/internal/tags"

	"cloud.google.com/go/storage"
	"github.com/instana/go-sensor/instrumentation/cloud.google.com/go/internal"
	ot "github.com/opentracing/opentracing-go"
)

// Writer is an instrumented wrapper for cloud.google.com/go/storage.Writer
// that traces calls made to Google Cloud Storage API.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#Writer for further details on wrapped type.
type Writer struct {
	*storage.Writer
	Bucket string

	ctx context.Context

	mu       sync.Mutex
	writeCtx context.Context
}

// Write calls the Write() method of the wrapped cloud.google.com/go/storage.Writer and initiates an exit span.
// Note that this span will be finished only when Close() is called, since writes are performed
// asynchronously and only guaranteed to be finished upon close. Thus each created span represents
// a single object insertion operation regardless of the number of Write() calls before the Writer
// is closed.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#Writer.Write for further details on wrapped method.
func (w *Writer) Write(p []byte) (n int, err error) {
	bucket := w.Writer.ObjectAttrs.Bucket
	if bucket == "" {
		bucket = w.Bucket
	}

	w.mu.Lock()
	if w.writeCtx == nil {
		w.writeCtx = internal.StartExitSpan(w.ctx, "gcs", ot.Tags{
			tags.GcsOp:     "objects.insert",
			tags.GcsBucket: bucket,
			tags.GcsObject: w.Writer.ObjectAttrs.Name,
		})
	}
	w.mu.Unlock()

	return w.Writer.Write(p)
}

// Close closes the underlying cloud.google.com/go/storage.Writer and finalizes current exit span.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#Writer.Close for further details on wrapped method.
func (w *Writer) Close() error {
	err := w.Writer.Close()

	w.mu.Lock()
	if w.writeCtx != nil {
		internal.FinishSpan(w.writeCtx, err)
		w.writeCtx = nil
	}
	w.mu.Unlock()

	return err
}

// CloseWithError terminates any writes performed by this Writer with an error and finalizes current exit span.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#Writer.CloseWithError for further details on wrapped method.
//
// Deprecated: this method is added for compatibility with the cloud.google.com/go/storage.Writer interface, however
// it is recommended to cancel the write operation using the context passed to NewWriter instead.
func (w *Writer) CloseWithError(err error) error {
	defer func() {
		w.mu.Lock()
		if w.writeCtx != nil {
			internal.FinishSpan(w.writeCtx, err)
			w.writeCtx = nil
		}
		w.mu.Unlock()
	}()

	return w.Writer.CloseWithError(err)
}
