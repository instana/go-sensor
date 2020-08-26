package storage

import (
	"context"

	"cloud.google.com/go/storage"
	"github.com/instana/go-sensor/instrumentation/cloud.google.com/go/internal"
	ot "github.com/opentracing/opentracing-go"
	"google.golang.org/api/option"
)

// Client is a client for interacting with Google Cloud Storage.
//
// Clients should be reused instead of created as needed.
// The methods of Client are safe for concurrent use by multiple goroutines.
type Client struct {
	*storage.Client
}

// NewClient returns a new wrapped cloud.google.com/go/storage.Client
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	c, err := storage.NewClient(ctx, opts...)
	return &Client{Client: c}, err
}

// ObjectHandle is an instrumented wrapper for cloud.google.com/go/storage.ObjectHandle
// that traces calls made to Google Cloud Storage API.
// Use BucketHandle.Object to get a handle.
type ObjectHandle struct {
	*storage.ObjectHandle
	Bucket string
	Name   string
}

// ACL returns an instrumented cloud.google.com/go/storage.ACLHandle that provides access to the
// object's access control list
func (o *ObjectHandle) ACL() *ACLHandle {
	return &ACLHandle{
		ACLHandle: o.ObjectHandle.ACL(),
		Bucket:    o.Bucket,
		Object:    o.Name,
	}
}

// Generation returns an instrumented ObjectHandle that operates on a specific generation
// of the object
func (o *ObjectHandle) Generation(gen int64) *ObjectHandle {
	return &ObjectHandle{
		ObjectHandle: o.ObjectHandle.Generation(gen),
		Bucket:       o.Bucket,
		Name:         o.Name,
	}
}

// If returns an instrumented ObjectHandle that applies a set of preconditions
func (o *ObjectHandle) If(conds storage.Conditions) *ObjectHandle {
	return &ObjectHandle{
		ObjectHandle: o.ObjectHandle.If(conds),
		Bucket:       o.Bucket,
		Name:         o.Name,
	}
}

// Key returns an instrumented ObjectHandle that uses the supplied encryption
// key to encrypt and decrypt the object's contents
func (o *ObjectHandle) Key(encryptionKey []byte) *ObjectHandle {
	return &ObjectHandle{
		ObjectHandle: o.ObjectHandle.Key(encryptionKey),
		Bucket:       o.Bucket,
		Name:         o.Name,
	}
}

// Attrs calls and traces the Attrs() method of the wrapped ObjectHandle
func (o *ObjectHandle) Attrs(ctx context.Context) (attrs *storage.ObjectAttrs, err error) {
	ctx = internal.StartExitSpan(ctx, "gcs", ot.Tags{
		"gcs.op":     "objects.get",
		"gcs.bucket": o.Bucket,
		"gcs.object": o.Name,
	})

	defer func() { internal.FinishSpan(ctx, err) }()

	return o.ObjectHandle.Attrs(ctx)
}

// Update calls and traces the Update() method of the wrapped ObjectHandle
func (o *ObjectHandle) Update(ctx context.Context, uattrs storage.ObjectAttrsToUpdate) (oa *storage.ObjectAttrs, err error) {
	ctx = internal.StartExitSpan(ctx, "gcs", ot.Tags{
		"gcs.op":     "objects.patch",
		"gcs.bucket": o.Bucket,
		"gcs.object": o.Name,
	})

	defer func() { internal.FinishSpan(ctx, err) }()

	return o.ObjectHandle.Update(ctx, uattrs)
}

// Delete calls and traces the Delete() method of the wrapped ObjectHandle
func (o *ObjectHandle) Delete(ctx context.Context) (err error) {
	ctx = internal.StartExitSpan(ctx, "gcs", ot.Tags{
		"gcs.op":     "objects.delete",
		"gcs.bucket": o.Bucket,
		"gcs.object": o.Name,
	})

	defer func() { internal.FinishSpan(ctx, err) }()

	return o.ObjectHandle.Delete(ctx)
}

// ReadCompressed returns an instrumented ObjectHandle that performs reads without
// decompressing when given true as an argument
func (o *ObjectHandle) ReadCompressed(compressed bool) *ObjectHandle {
	return &ObjectHandle{
		ObjectHandle: o.ObjectHandle.ReadCompressed(compressed),
		Bucket:       o.Bucket,
		Name:         o.Name,
	}
}

// NewWriter returns a storage Writer that writes to the GCS object
// associated with this ObjectHandle.
//
// A new object will be created unless an object with this name already exists.
// Otherwise any previous object with the same name will be replaced.
// The object will not be available (and any previous object will remain)
// until Close has been called.
//
// Attributes can be set on the object by modifying the returned Writer's
// ObjectAttrs field before the first call to Write. If no ContentType
// attribute is specified, the content type will be automatically sniffed
// using net/http.DetectContentType.
//
// It is the caller's responsibility to call Close when writing is done. To
// stop writing without saving the data, cancel the context.
//
// INSTRUMENT
func (o *ObjectHandle) NewWriter(ctx context.Context) *Writer {
	return &Writer{o.ObjectHandle.NewWriter(ctx)}
}

// ServiceAccount fetches the email address of the given project's Google Cloud Storage service account.
//
// INSTRUMENT
func (c *Client) ServiceAccount(ctx context.Context, projectID string) (string, error) {
	return c.Client.ServiceAccount(ctx, projectID)
}
