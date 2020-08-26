package storage

import (
	"context"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

// Client is a client for interacting with Google Cloud Storage.
//
// Clients should be reused instead of created as needed.
// The methods of Client are safe for concurrent use by multiple goroutines.
type Client struct {
	*storage.Client
}

// NewClient creates a new Google Cloud Storage client.
// The default scope is ScopeFullControl. To use a different scope, like
// ScopeReadOnly, use option.WithScopes.
//
// Clients should be reused instead of created as needed. The methods of Client
// are safe for concurrent use by multiple goroutines.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	c, err := storage.NewClient(ctx, opts...)
	return &Client{
		Client: c,
	}, err
}

// ObjectHandle provides operations on an object in a Google Cloud Storage bucket.
// Use BucketHandle.Object to get a handle.
type ObjectHandle struct {
	*storage.ObjectHandle
}

// ACL provides access to the object's access control list.
// This controls who can read and write this object.
// This call does not perform any network operations.
func (o *ObjectHandle) ACL() *ACLHandle {
	return &ACLHandle{o.ObjectHandle.ACL()}
}

// Generation returns a new ObjectHandle that operates on a specific generation
// of the object.
// By default, the handle operates on the latest generation. Not
// all operations work when given a specific generation; check the API
// endpoints at https://cloud.google.com/storage/docs/json_api/ for details.
func (o *ObjectHandle) Generation(gen int64) *ObjectHandle {
	return &ObjectHandle{o.ObjectHandle.Generation(gen)}
}

// If returns a new ObjectHandle that applies a set of preconditions.
// Preconditions already set on the ObjectHandle are ignored.
// Operations on the new handle will return an error if the preconditions are not
// satisfied. See https://cloud.google.com/storage/docs/generations-preconditions
// for more details.
func (o *ObjectHandle) If(conds storage.Conditions) *ObjectHandle {
	return &ObjectHandle{o.ObjectHandle.If(conds)}
}

// Key returns a new ObjectHandle that uses the supplied encryption
// key to encrypt and decrypt the object's contents.
//
// Encryption key must be a 32-byte AES-256 key.
// See https://cloud.google.com/storage/docs/encryption for details.
func (o *ObjectHandle) Key(encryptionKey []byte) *ObjectHandle {
	return &ObjectHandle{o.ObjectHandle.Key(encryptionKey)}
}

// Attrs returns meta information about the object.
// ErrObjectNotExist will be returned if the object is not found.
//
// INSTRUMENT
func (o *ObjectHandle) Attrs(ctx context.Context) (attrs *storage.ObjectAttrs, err error) {
	return o.ObjectHandle.Attrs(ctx)
}

// Update updates an object with the provided attributes.
// All zero-value attributes are ignored.
// ErrObjectNotExist will be returned if the object is not found.
//
// INSTRUMENT
func (o *ObjectHandle) Update(ctx context.Context, uattrs storage.ObjectAttrsToUpdate) (oa *storage.ObjectAttrs, err error) {
	return o.ObjectHandle.Update(ctx, uattrs)
}

// Delete deletes the single specified object.
//
// INSTRUMENT
func (o *ObjectHandle) Delete(ctx context.Context) error {
	return o.ObjectHandle.Delete(ctx)
}

// ReadCompressed when true causes the read to happen without decompressing.
func (o *ObjectHandle) ReadCompressed(compressed bool) *ObjectHandle {
	return &ObjectHandle{o.ObjectHandle.ReadCompressed(compressed)}
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
