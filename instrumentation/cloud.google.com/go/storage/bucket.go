package storage

import (
	"context"

	"cloud.google.com/go/storage"
)

// BucketHandle provides operations on a Google Cloud Storage bucket.
// Use Client.Bucket to get a handle.
type BucketHandle struct {
	*storage.BucketHandle
}

// Bucket returns a BucketHandle, which provides operations on the named bucket.
// This call does not perform any network operations.
//
// The supplied name must contain only lowercase letters, numbers, dashes,
// underscores, and dots. The full specification for valid bucket names can be
// found at:
//   https://cloud.google.com/storage/docs/bucket-naming
func (c *Client) Bucket(name string) *BucketHandle {
	return &BucketHandle{c.Client.Bucket(name)}
}

// Create creates the Bucket in the project.
// If attrs is nil the API defaults will be used.
//
// INSTRUMENT
func (b *BucketHandle) Create(ctx context.Context, projectID string, attrs *storage.BucketAttrs) error {
	return b.BucketHandle.Create(ctx, projectID, attrs)
}

// Delete deletes the Bucket.
//
// INSTRUMENT
func (b *BucketHandle) Delete(ctx context.Context) error {
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

// Attrs returns the metadata for the bucket.
//
// INSTRUMENT
func (b *BucketHandle) Attrs(ctx context.Context) (attrs *storage.BucketAttrs, err error) {
	return b.BucketHandle.Attrs(ctx)
}

// Update updates a bucket's attributes.
//
// INSTRUMENT
func (b *BucketHandle) Update(ctx context.Context, uattrs storage.BucketAttrsToUpdate) (attrs *storage.BucketAttrs, err error) {
	return b.BucketHandle.Update(ctx, uattrs)
}

// If returns a new BucketHandle that applies a set of preconditions.
// Preconditions already set on the BucketHandle are ignored.
// Operations on the new handle will return an error if the preconditions are not
// satisfied. The only valid preconditions for buckets are MetagenerationMatch
// and MetagenerationNotMatch.
func (b *BucketHandle) If(conds storage.BucketConditions) *BucketHandle {
	return &BucketHandle{b.BucketHandle.If(conds)}
}

// UserProject returns a new BucketHandle that passes the project ID as the user
// project for all subsequent calls. Calls with a user project will be billed to that
// project rather than to the bucket's owning project.
//
// A user project is required for all operations on Requester Pays buckets.
func (b *BucketHandle) UserProject(projectID string) *BucketHandle {
	return &BucketHandle{b.BucketHandle.UserProject(projectID)}
}

// LockRetentionPolicy locks a bucket's retention policy until a previously-configured
// RetentionPeriod past the EffectiveTime. Note that if RetentionPeriod is set to less
// than a day, the retention policy is treated as a development configuration and locking
// will have no effect. The BucketHandle must have a metageneration condition that
// matches the bucket's metageneration. See BucketHandle.If.
//
// This feature is in private alpha release. It is not currently available to
// most customers. It might be changed in backwards-incompatible ways and is not
// subject to any SLA or deprecation policy.
//
// INSTRUMENT
func (b *BucketHandle) LockRetentionPolicy(ctx context.Context) error {
	return b.BucketHandle.LockRetentionPolicy(ctx)
}

// Objects returns an iterator over the objects in the bucket that match the Query q.
// If q is nil, no filtering is done.
//
// Note: The returned iterator is not safe for concurrent operations without explicit synchronization.
func (b *BucketHandle) Objects(ctx context.Context, q *storage.Query) *ObjectIterator {
	return &ObjectIterator{b.BucketHandle.Objects(ctx, q)}
}

// An ObjectIterator is an iterator over ObjectAttrs.
//
// Note: This iterator is not safe for concurrent operations without explicit synchronization.
type ObjectIterator struct {
	*storage.ObjectIterator
}

// Next returns the next result. Its second return value is iterator.Done if
// there are no more results. Once Next returns iterator.Done, all subsequent
// calls will return iterator.Done.
//
// If Query.Delimiter is non-empty, some of the ObjectAttrs returned by Next will
// have a non-empty Prefix field, and a zero value for all other fields. These
// represent prefixes.
//
// Note: This method is not safe for concurrent operations without explicit synchronization.
//
// INSTRUMENT
func (it *ObjectIterator) Next() (*storage.ObjectAttrs, error) {
	return it.ObjectIterator.Next()
}

// Buckets returns an iterator over the buckets in the project. You may
// optionally set the iterator's Prefix field to restrict the list to buckets
// whose names begin with the prefix. By default, all buckets in the project
// are returned.
//
// Note: The returned iterator is not safe for concurrent operations without explicit synchronization.
func (c *Client) Buckets(ctx context.Context, projectID string) *BucketIterator {
	return &BucketIterator{c.Client.Buckets(ctx, projectID)}
}

// A BucketIterator is an iterator over BucketAttrs.
//
// Note: This iterator is not safe for concurrent operations without explicit synchronization.
type BucketIterator struct {
	*storage.BucketIterator
}

// Next returns the next result. Its second return value is iterator.Done if
// there are no more results. Once Next returns iterator.Done, all subsequent
// calls will return iterator.Done.
//
// Note: This method is not safe for concurrent operations without explicit synchronization.
//
// INSTRUMENT
func (it *BucketIterator) Next() (*storage.BucketAttrs, error) {
	return it.BucketIterator.Next()
}
