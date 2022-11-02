// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package storage

import (
	"context"

	"cloud.google.com/go/storage"
	"github.com/instana/go-sensor/instrumentation/cloud.google.com/go/internal"
	"github.com/instana/go-sensor/instrumentation/cloud.google.com/go/internal/tags"
	ot "github.com/opentracing/opentracing-go"
	"google.golang.org/api/iterator"
)

// BucketHandle is an instrumented wrapper for cloud.google.com/go/storage.BucketHandle
// that traces calls made to Google Cloud Storage API.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#BucketHandle for further details on wrapped type.
type BucketHandle struct {
	*storage.BucketHandle
	Name string
}

// Bucket returns an instrumented cloud.google.com/go/storage.BucketHandle.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#Client.Bucket for further details on wrapped method.
func (c *Client) Bucket(name string) *BucketHandle {
	return &BucketHandle{
		BucketHandle: c.Client.Bucket(name),
		Name:         name,
	}
}

// Create calls and traces the Create() method of the wrapped cloud.google.com/go/storage.BucketHandle.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#BucketHandle.Create for further details on wrapped method.
func (b *BucketHandle) Create(ctx context.Context, projectID string, attrs *storage.BucketAttrs) (err error) {
	ctx = internal.StartExitSpan(ctx, "gcs", ot.Tags{
		tags.GcsOp:        "buckets.insert",
		tags.GcsBucket:    b.Name,
		tags.GcsProjectId: projectID,
	})

	defer func() { internal.FinishSpan(ctx, err) }()

	return b.BucketHandle.Create(ctx, projectID, attrs)
}

// Delete calls and traces the Delete() method of the wrapped cloud.google.com/go/storage.BucketHandle.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#BucketHandle.Delete for further details on wrapped method.
func (b *BucketHandle) Delete(ctx context.Context) (err error) {
	ctx = internal.StartExitSpan(ctx, "gcs", ot.Tags{
		tags.GcsOp:     "buckets.delete",
		tags.GcsBucket: b.Name,
	})

	defer func() { internal.FinishSpan(ctx, err) }()

	return b.BucketHandle.Delete(ctx)
}

// ACL returns an instrumented cloud.google.com/go/storage.ACLHandle.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#BucketHandle.ACL for further details on wrapped method.
func (b *BucketHandle) ACL() *ACLHandle {
	return &ACLHandle{
		ACLHandle: b.BucketHandle.ACL(),
		Bucket:    b.Name,
	}
}

// DefaultObjectACL returns an instrumented cloud.google.com/go/storage.ACLHandle, which provides
// access to the bucket's default object ACLs.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#BucketHandle.DefaultObjectACL for further details on wrapped method.
func (b *BucketHandle) DefaultObjectACL() *ACLHandle {
	return &ACLHandle{
		ACLHandle: b.BucketHandle.DefaultObjectACL(),
		Bucket:    b.Name,
		Default:   true,
	}
}

// Object returns an instrumented cloud.google.com/go/storage.ObjectHandle, which provides operations
// on the named object.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#BucketHandle.Object for further details on wrapped method.
func (b *BucketHandle) Object(name string) *ObjectHandle {
	return &ObjectHandle{
		ObjectHandle: b.BucketHandle.Object(name),
		Bucket:       b.Name,
		Name:         name,
	}
}

// Attrs calls and traces the Attrs() method of the wrapped cloud.google.com/go/storage.BucketHandle.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#BucketHandle.Attrs for further details on wrapped method.
func (b *BucketHandle) Attrs(ctx context.Context) (attrs *storage.BucketAttrs, err error) {
	ctx = internal.StartExitSpan(ctx, "gcs", ot.Tags{
		tags.GcsOp:     "buckets.get",
		tags.GcsBucket: b.Name,
	})

	defer func() { internal.FinishSpan(ctx, err) }()

	return b.BucketHandle.Attrs(ctx)
}

// Update calls and traces the Update() method of the wrapped cloud.google.com/go/storage.BucketHandle.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#BucketHandle.Update for further details on wrapped method.
func (b *BucketHandle) Update(ctx context.Context, uattrs storage.BucketAttrsToUpdate) (attrs *storage.BucketAttrs, err error) {
	ctx = internal.StartExitSpan(ctx, "gcs", ot.Tags{
		tags.GcsOp:     "buckets.patch",
		tags.GcsBucket: b.Name,
	})

	defer func() { internal.FinishSpan(ctx, err) }()

	return b.BucketHandle.Update(ctx, uattrs)
}

// If returns an instrumented cloud.google.com/go/storage.BucketHandle that applies a set of preconditions.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#BucketHandle.If for further details on wrapped method.
func (b *BucketHandle) If(conds storage.BucketConditions) *BucketHandle {
	return &BucketHandle{
		BucketHandle: b.BucketHandle.If(conds),
		Name:         b.Name,
	}
}

// UserProject returns an instrumented cloud.google.com/go/storage.BucketHandle that passes the project ID as the user
// project for all subsequent calls.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#BucketHandle.UserProject for further details on wrapped method.
func (b *BucketHandle) UserProject(projectID string) *BucketHandle {
	return &BucketHandle{
		BucketHandle: b.BucketHandle.UserProject(projectID),
		Name:         b.Name,
	}
}

// LockRetentionPolicy calls and traces the LockRetentionPolicy() method of the wrapped cloud.google.com/go/storage.BucketHandle.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#BucketHandle.LockRetentionPolicy for further details on wrapped method.
func (b *BucketHandle) LockRetentionPolicy(ctx context.Context) (err error) {
	ctx = internal.StartExitSpan(ctx, "gcs", ot.Tags{
		tags.GcsOp:     "buckets.lockRetentionPolicy",
		tags.GcsBucket: b.Name,
	})

	defer func() { internal.FinishSpan(ctx, err) }()

	return b.BucketHandle.LockRetentionPolicy(ctx)
}

// Objects returns an instrumented object iterator that traces and proxies requests to
// the underlying cloud.google.com/go/storage.ObjectIterator.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#BucketHandle.Objects for further details on wrapped method.
func (b *BucketHandle) Objects(ctx context.Context, q *storage.Query) *ObjectIterator {
	return &ObjectIterator{
		ObjectIterator: b.BucketHandle.Objects(ctx, q),
		Bucket:         b.Name,
		ctx:            ctx,
	}
}

// ObjectIterator is an instrumented wrapper for cloud.google.com/go/storage.ObjectIterator
// that traces calls made to Google Cloud Storage API.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#ObjectIterator for further details on wrapped type.
type ObjectIterator struct {
	*storage.ObjectIterator
	Bucket string
	ctx    context.Context
}

// Next calls the Next() method of the wrapped iterator and creates a span for each call
// that results in an API request.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#ObjectIterator.Next for further details on wrapped method.
func (it *ObjectIterator) Next() (attrs *storage.ObjectAttrs, err error) {
	// don't trace calls returning buffered data
	if it.ObjectIterator.PageInfo().Remaining() > 0 {
		return it.ObjectIterator.Next()
	}

	ctx := internal.StartExitSpan(it.ctx, "gcs", ot.Tags{
		tags.GcsOp:     "objects.list",
		tags.GcsBucket: it.Bucket,
	})

	defer func() {
		if err == iterator.Done {
			// the last iterator call only meant for signalling
			// that all items have been processed, we don't need
			// a separate span for this
			return
		}

		internal.FinishSpan(ctx, err)
	}()

	return it.ObjectIterator.Next()
}

// Buckets returns an instrumented bucket iterator that traces and proxies requests to
// the underlying cloud.google.com/go/storage.BucketIterator.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#Client.Buckets for further details on wrapped method.
func (c *Client) Buckets(ctx context.Context, projectID string) *BucketIterator {
	return &BucketIterator{
		BucketIterator: c.Client.Buckets(ctx, projectID),
		projectID:      projectID,
		ctx:            ctx,
	}
}

// BucketIterator is an instrumented wrapper for cloud.google.com/go/storage.BucketIterator.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#BucketIterator for further details on wrapped type.
type BucketIterator struct {
	*storage.BucketIterator
	projectID string
	ctx       context.Context
}

// Next calls the Next() method of the wrapped iterator and creates a span for each call
// that results in an API request.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#BucketIterator.Next for further details on wrapped method.
func (it *BucketIterator) Next() (attrs *storage.BucketAttrs, err error) {
	// don't trace calls returning buffered data
	if it.BucketIterator.PageInfo().Remaining() > 0 {
		return it.BucketIterator.Next()
	}

	ctx := internal.StartExitSpan(it.ctx, "gcs", ot.Tags{
		tags.GcsOp:        "buckets.list",
		tags.GcsProjectId: it.projectID,
	})

	defer func() {
		if err == iterator.Done {
			// the last iterator call only meant for signalling
			// that all items have been processed, we don't need
			// a separate span for this
			return
		}

		internal.FinishSpan(ctx, err)
	}()

	return it.BucketIterator.Next()
}
