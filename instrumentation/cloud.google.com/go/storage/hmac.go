package storage

import (
	"context"

	"cloud.google.com/go/storage"
)

// HMACKeyHandle helps provide access and management for HMAC keys.
//
// This type is EXPERIMENTAL and subject to change or removal without notice.
type HMACKeyHandle struct {
	*storage.HMACKeyHandle
}

// HMACKeyHandle creates a handle that will be used for HMACKey operations.
//
// This method is EXPERIMENTAL and subject to change or removal without notice.
func (c *Client) HMACKeyHandle(projectID, accessID string) *HMACKeyHandle {
	return &HMACKeyHandle{c.Client.HMACKeyHandle(projectID, accessID)}
}

// Get invokes an RPC to retrieve the HMAC key referenced by the
// HMACKeyHandle's accessID.
//
// Options such as UserProjectForHMACKeys can be used to set the
// userProject to be billed against for operations.
//
// This method is EXPERIMENTAL and subject to change or removal without notice.
//
// INSTRUMENT
func (hkh *HMACKeyHandle) Get(ctx context.Context, opts ...storage.HMACKeyOption) (*storage.HMACKey, error) {
	return hkh.HMACKeyHandle.Get(ctx, opts...)
}

// Delete invokes an RPC to delete the key referenced by accessID, on Google Cloud Storage.
// Only inactive HMAC keys can be deleted.
// After deletion, a key cannot be used to authenticate requests.
//
// This method is EXPERIMENTAL and subject to change or removal without notice.
//
// INSTRUMENT
func (hkh *HMACKeyHandle) Delete(ctx context.Context, opts ...storage.HMACKeyOption) error {
	return hkh.HMACKeyHandle.Delete(ctx, opts...)
}

// CreateHMACKey invokes an RPC for Google Cloud Storage to create a new HMACKey.
//
// This method is EXPERIMENTAL and subject to change or removal without notice.
//
// INSTRUMENT
func (c *Client) CreateHMACKey(ctx context.Context, projectID, serviceAccountEmail string, opts ...storage.HMACKeyOption) (*storage.HMACKey, error) {
	return c.Client.CreateHMACKey(ctx, projectID, serviceAccountEmail, opts...)
}

// Update mutates the HMACKey referred to by accessID.
//
// This method is EXPERIMENTAL and subject to change or removal without notice.
//
// INSTRUMENT
func (h *HMACKeyHandle) Update(ctx context.Context, au storage.HMACKeyAttrsToUpdate, opts ...storage.HMACKeyOption) (*storage.HMACKey, error) {
	return h.HMACKeyHandle.Update(ctx, au, opts...)
}

// An HMACKeysIterator is an iterator over HMACKeys.
//
// Note: This iterator is not safe for concurrent operations without explicit synchronization.
//
// This type is EXPERIMENTAL and subject to change or removal without notice.
type HMACKeysIterator struct {
	*storage.HMACKeysIterator
}

// ListHMACKeys returns an iterator for listing HMACKeys.
//
// Note: This iterator is not safe for concurrent operations without explicit synchronization.
//
// This method is EXPERIMENTAL and subject to change or removal without notice.
func (c *Client) ListHMACKeys(ctx context.Context, projectID string, opts ...storage.HMACKeyOption) *HMACKeysIterator {
	return &HMACKeysIterator{c.Client.ListHMACKeys(ctx, projectID, opts...)}
}

// Next returns the next result. Its second return value is iterator.Done if
// there are no more results. Once Next returns iterator.Done, all subsequent
// calls will return iterator.Done.
//
// Note: This iterator is not safe for concurrent operations without explicit synchronization.
//
// This method is EXPERIMENTAL and subject to change or removal without notice.
//
// INSTRUMENT
func (it *HMACKeysIterator) Next() (*storage.HMACKey, error) {
	return it.HMACKeysIterator.Next()
}
