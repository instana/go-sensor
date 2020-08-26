package storage

import (
	"context"

	"cloud.google.com/go/storage"
	"github.com/instana/go-sensor/instrumentation/cloud.google.com/go/internal"
	ot "github.com/opentracing/opentracing-go"
	"google.golang.org/api/iterator"
)

// BucketHandle is an instrumented wrapper for cloud.google.com/go/storage.BucketHandle
// that traces calls made to Google Cloud Storage API
type BucketHandle struct {
	*storage.BucketHandle
	Name string
}

// Bucket returns an instrumented cloud.google.com/go/storage.BucketHandle
func (c *Client) Bucket(name string) *BucketHandle {
	return &BucketHandle{
		BucketHandle: c.Client.Bucket(name),
		Name:         name,
	}
}

// Create calls and traces the Create() method of the wrapped BucketHandle
func (b *BucketHandle) Create(ctx context.Context, projectID string, attrs *storage.BucketAttrs) (err error) {
	ctx = internal.StartExitSpan(ctx, "gcs", ot.Tags{
		"gcs.op":        "buckets.insert",
		"gcs.bucket":    b.Name,
		"gcs.projectId": projectID,
	})

	defer func() { internal.FinishSpan(ctx, err) }()

	return b.BucketHandle.Create(ctx, projectID, attrs)
}

// Delete calls and traces the Delete() method of the wrapped BucketHandle
func (b *BucketHandle) Delete(ctx context.Context) (err error) {
	ctx = internal.StartExitSpan(ctx, "gcs", ot.Tags{
		"gcs.op":     "buckets.delete",
		"gcs.bucket": b.Name,
	})

	defer func() { internal.FinishSpan(ctx, err) }()

	return b.BucketHandle.Delete(ctx)
}

// ACL returns an ACLHandle, which provides access to the bucket's access control list.
// This controls who can list, create or overwrite the objects in a bucket.
// This call does not perform any network operations.
func (b *BucketHandle) ACL() *ACLHandle {
	return &ACLHandle{b.BucketHandle.ACL()}
}

// DefaultObjectACL returns an ACLHandle, which provides access to the bucket's default object ACLs.
// These ACLs are applied to newly created objects in this bucket that do not have a defined ACL.
// This call does not perform any network operations.
func (b *BucketHandle) DefaultObjectACL() *ACLHandle {
	return &ACLHandle{b.BucketHandle.DefaultObjectACL()}
}

// Object returns an ObjectHandle, which provides operations on the named object.
// This call does not perform any network operations.
//
// name must consist entirely of valid UTF-8-encoded runes. The full specification
// for valid object names can be found at:
//   https://cloud.google.com/storage/docs/bucket-naming
func (b *BucketHandle) Object(name string) *ObjectHandle {
	return &ObjectHandle{b.BucketHandle.Object(name)}
}

// Attrs calls and traces the Attrs() method of the wrapped BucketHandle
func (b *BucketHandle) Attrs(ctx context.Context) (attrs *storage.BucketAttrs, err error) {
	ctx = internal.StartExitSpan(ctx, "gcs", ot.Tags{
		"gcs.op":     "buckets.get",
		"gcs.bucket": b.Name,
	})

	defer func() { internal.FinishSpan(ctx, err) }()

	return b.BucketHandle.Attrs(ctx)
}

// Update calls and traces the Update() method of the wrapped BucketHandle
func (b *BucketHandle) Update(ctx context.Context, uattrs storage.BucketAttrsToUpdate) (attrs *storage.BucketAttrs, err error) {
	ctx = internal.StartExitSpan(ctx, "gcs", ot.Tags{
		"gcs.op":     "buckets.patch",
		"gcs.bucket": b.Name,
	})

	defer func() { internal.FinishSpan(ctx, err) }()

	return b.BucketHandle.Update(ctx, uattrs)
}

// If returns an instrumented BucketHandle that applies set of preconditions
func (b *BucketHandle) If(conds storage.BucketConditions) *BucketHandle {
	return &BucketHandle{
		BucketHandle: b.BucketHandle.If(conds),
		Name:         b.Name,
	}
}

// UserProject returns an instrumented cloud.google.com/go/storage.BucketHandle that passes the project ID as the user
// project for all subsequent calls
func (b *BucketHandle) UserProject(projectID string) *BucketHandle {
	return &BucketHandle{
		BucketHandle: b.BucketHandle.UserProject(projectID),
		Name:         b.Name,
	}
}

// LockRetentionPolicy calls and traces the LockRetentionPolicy() method of the wrapped BucketHandle
func (b *BucketHandle) LockRetentionPolicy(ctx context.Context) (err error) {
	ctx = internal.StartExitSpan(ctx, "gcs", ot.Tags{
		"gcs.op":     "buckets.lockRetentionPolicy",
		"gcs.bucket": b.Name,
	})

	defer func() { internal.FinishSpan(ctx, err) }()

	return b.BucketHandle.LockRetentionPolicy(ctx)
}

// Objects returns an instrumented object iterator that traces and proxies requests to
// the underlying cloud.google.com/go/storage.ObjectIterator
func (b *BucketHandle) Objects(ctx context.Context, q *storage.Query) *ObjectIterator {
	return &ObjectIterator{
		ObjectIterator: b.BucketHandle.Objects(ctx, q),
		Bucket:         b.Name,
		ctx:            ctx,
	}
}

// ObjectIterator is an instrumented wrapper for cloud.google.com/go/storage.ObjectIterator
// that traces calls made to Google Cloud Storage API
type ObjectIterator struct {
	*storage.ObjectIterator
	Bucket string
	ctx    context.Context
}

// Next calls the Next() method of the wrapped iterator and creates a span for each call
// that results in an API request
func (it *ObjectIterator) Next() (attrs *storage.ObjectAttrs, err error) {
	// don't trace calls returning buffered data
	if it.ObjectIterator.PageInfo().Remaining() > 0 {
		return it.ObjectIterator.Next()
	}

	ctx := internal.StartExitSpan(it.ctx, "gcs", ot.Tags{
		"gcs.op":     "objects.list",
		"gcs.bucket": it.Bucket,
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
// the underlying cloud.google.com/go/storage.BucketIterator
func (c *Client) Buckets(ctx context.Context, projectID string) *BucketIterator {
	return &BucketIterator{
		BucketIterator: c.Client.Buckets(ctx, projectID),
		projectID:      projectID,
		ctx:            ctx,
	}
}

// BucketIterator is an instrumented wrapper for cloud.google.com/go/storage.BucketIterator
type BucketIterator struct {
	*storage.BucketIterator
	projectID string
	ctx       context.Context
}

// Next calls the Next() method of the wrapped iterator and creates a span for each call
// that results in an API request
func (it *BucketIterator) Next() (attrs *storage.BucketAttrs, err error) {
	// don't trace calls returning buffered data
	if it.BucketIterator.PageInfo().Remaining() > 0 {
		return it.BucketIterator.Next()
	}

	ctx := internal.StartExitSpan(it.ctx, "gcs", ot.Tags{
		"gcs.op":        "buckets.list",
		"gcs.projectId": it.projectID,
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
