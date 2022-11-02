// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package storage

import (
	"context"

	"cloud.google.com/go/storage"
	"github.com/instana/go-sensor/instrumentation/cloud.google.com/go/internal"
	"github.com/instana/go-sensor/instrumentation/cloud.google.com/go/internal/tags"
	ot "github.com/opentracing/opentracing-go"
	"google.golang.org/api/option"
)

// Client is an instrumented wrapper for cloud.google.com/go/storage.Client
// that traces calls made to Google Cloud Storage API.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#Client for further details on wrapped type.
type Client struct {
	*storage.Client
}

// NewClient returns a new wrapped cloud.google.com/go/storage.Client.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#NewClient for further details on wrapped method.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	c, err := storage.NewClient(ctx, opts...)
	return &Client{Client: c}, err
}

// ObjectHandle is an instrumented wrapper for cloud.google.com/go/storage.ObjectHandle
// that traces calls made to Google Cloud Storage API.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#ObjectHandle for further details on wrapped type.
type ObjectHandle struct {
	*storage.ObjectHandle
	Bucket string
	Name   string
}

// ACL returns an instrumented cloud.google.com/go/storage.ACLHandle that provides access to the
// object's access control list.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#ObjectHandle.ACL for further details on wrapped method.
func (o *ObjectHandle) ACL() *ACLHandle {
	return &ACLHandle{
		ACLHandle: o.ObjectHandle.ACL(),
		Bucket:    o.Bucket,
		Object:    o.Name,
	}
}

// Generation returns an instrumented cloud.google.com/go/storage.ObjectHandle that operates on a specific generation
// of the object.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#ObjectHandle.Generation for further details on wrapped method.
func (o *ObjectHandle) Generation(gen int64) *ObjectHandle {
	return &ObjectHandle{
		ObjectHandle: o.ObjectHandle.Generation(gen),
		Bucket:       o.Bucket,
		Name:         o.Name,
	}
}

// If returns an instrumented cloud.google.com/go/storage.ObjectHandle that applies a set of preconditions.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#ObjectHandle.If for further details on wrapped method.
func (o *ObjectHandle) If(conds storage.Conditions) *ObjectHandle {
	return &ObjectHandle{
		ObjectHandle: o.ObjectHandle.If(conds),
		Bucket:       o.Bucket,
		Name:         o.Name,
	}
}

// Key returns an instrumented cloud.google.com/go/storage.ObjectHandle that uses the supplied encryption
// key to encrypt and decrypt the object's contents.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#ObjectHandle.Key for further details on wrapped method.
func (o *ObjectHandle) Key(encryptionKey []byte) *ObjectHandle {
	return &ObjectHandle{
		ObjectHandle: o.ObjectHandle.Key(encryptionKey),
		Bucket:       o.Bucket,
		Name:         o.Name,
	}
}

// Attrs calls and traces the Attrs() method of the wrapped cloud.google.com/go/storage.ObjectHandle.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#ObjectHandle.Attrs for further details on wrapped method.
func (o *ObjectHandle) Attrs(ctx context.Context) (attrs *storage.ObjectAttrs, err error) {
	ctx = internal.StartExitSpan(ctx, "gcs", ot.Tags{
		tags.GcsOp:     "objects.attrs",
		tags.GcsBucket: o.Bucket,
		tags.GcsObject: o.Name,
	})

	defer func() { internal.FinishSpan(ctx, err) }()

	return o.ObjectHandle.Attrs(ctx)
}

// Update calls and traces the Update() method of the wrapped cloud.google.com/go/storage.ObjectHandle.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#ObjectHandle.Update for further details on wrapped method.
func (o *ObjectHandle) Update(ctx context.Context, uattrs storage.ObjectAttrsToUpdate) (oa *storage.ObjectAttrs, err error) {
	ctx = internal.StartExitSpan(ctx, "gcs", ot.Tags{
		tags.GcsOp:     "objects.patch",
		tags.GcsBucket: o.Bucket,
		tags.GcsObject: o.Name,
	})

	defer func() { internal.FinishSpan(ctx, err) }()

	return o.ObjectHandle.Update(ctx, uattrs)
}

// Delete calls and traces the Delete() method of the wrapped cloud.google.com/go/storage.ObjectHandle.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#ObjectHandle.Delete for further details on wrapped method.
func (o *ObjectHandle) Delete(ctx context.Context) (err error) {
	ctx = internal.StartExitSpan(ctx, "gcs", ot.Tags{
		tags.GcsOp:     "objects.delete",
		tags.GcsBucket: o.Bucket,
		tags.GcsObject: o.Name,
	})

	defer func() { internal.FinishSpan(ctx, err) }()

	return o.ObjectHandle.Delete(ctx)
}

// ReadCompressed returns an instrumented cloud.google.com/go/storage.ObjectHandle that performs reads without
// decompressing when given true as an argument.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#ObjectHandle.ReadCompressed for further details on wrapped method.
func (o *ObjectHandle) ReadCompressed(compressed bool) *ObjectHandle {
	return &ObjectHandle{
		ObjectHandle: o.ObjectHandle.ReadCompressed(compressed),
		Bucket:       o.Bucket,
		Name:         o.Name,
	}
}

// NewWriter returns an instrumented cloud.google.com/go/storage.Writer
// that traces calls made to Google Cloud Storage API.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#ObjectHandle.NewWriter for further details on wrapped method.
func (o *ObjectHandle) NewWriter(ctx context.Context) *Writer {
	return &Writer{
		Writer: o.ObjectHandle.NewWriter(ctx),
		ctx:    ctx,
		Bucket: o.Bucket,
	}
}

// ServiceAccount calls and traces the ServiceAccount() method of the wrapped cloud.google.com/go/storage.Client.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#Client.ServiceAccount for further details on wrapped method.
func (c *Client) ServiceAccount(ctx context.Context, projectID string) (email string, err error) {
	ctx = internal.StartExitSpan(ctx, "gcs", ot.Tags{
		tags.GcsOp:        "serviceAccount.get",
		tags.GcsProjectId: projectID,
	})

	defer func() { internal.FinishSpan(ctx, err) }()

	return c.Client.ServiceAccount(ctx, projectID)
}
