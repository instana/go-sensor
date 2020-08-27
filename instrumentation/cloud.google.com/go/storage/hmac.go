package storage

import (
	"context"

	"cloud.google.com/go/storage"
	"github.com/instana/go-sensor/instrumentation/cloud.google.com/go/internal"
	ot "github.com/opentracing/opentracing-go"
)

// HMACKeyHandle is an instrumented wrapper for cloud.google.com/go/storage.HMACKeyHandle
// that traces calls made to Google Cloud Storage API
type HMACKeyHandle struct {
	*storage.HMACKeyHandle
	ProjectID string
	AccessID  string
}

// HMACKeyHandle returns an instrumented cloud.google.com/go/storage.HMACKeyHandle
func (c *Client) HMACKeyHandle(projectID, accessID string) *HMACKeyHandle {
	return &HMACKeyHandle{
		HMACKeyHandle: c.Client.HMACKeyHandle(projectID, accessID),
		ProjectID:     projectID,
		AccessID:      accessID,
	}
}

// Get calls and traces the Get() method of the wrapped HMACKeyHandle
func (hkh *HMACKeyHandle) Get(ctx context.Context, opts ...storage.HMACKeyOption) (hk *storage.HMACKey, err error) {
	ctx = internal.StartExitSpan(ctx, "gcs", ot.Tags{
		"gcs.op":        "hmacKeys.get",
		"gcs.projectId": hkh.ProjectID,
		"gcs.accessId":  hkh.AccessID,
	})

	defer func() { internal.FinishSpan(ctx, err) }()

	return hkh.HMACKeyHandle.Get(ctx, opts...)
}

// Delete calls and traces the Delete() method of the wrapped HMACKeyHandle
func (hkh *HMACKeyHandle) Delete(ctx context.Context, opts ...storage.HMACKeyOption) (err error) {
	ctx = internal.StartExitSpan(ctx, "gcs", ot.Tags{
		"gcs.op":        "hmacKeys.delete",
		"gcs.projectId": hkh.ProjectID,
		"gcs.accessId":  hkh.AccessID,
	})

	defer func() { internal.FinishSpan(ctx, err) }()

	return hkh.HMACKeyHandle.Delete(ctx, opts...)
}

// CreateHMACKey calls and traces the CreateHMACKey() method of the wrapped Client
func (c *Client) CreateHMACKey(ctx context.Context, projectID, serviceAccountEmail string, opts ...storage.HMACKeyOption) (hk *storage.HMACKey, err error) {
	ctx = internal.StartExitSpan(ctx, "gcs", ot.Tags{
		"gcs.op":        "hmacKeys.create",
		"gcs.projectId": projectID,
	})

	defer func() { internal.FinishSpan(ctx, err) }()

	return c.Client.CreateHMACKey(ctx, projectID, serviceAccountEmail, opts...)
}

// Update calls and traces the Update() method of the wrapped HMACKeyHandle
func (h *HMACKeyHandle) Update(ctx context.Context, au storage.HMACKeyAttrsToUpdate, opts ...storage.HMACKeyOption) (hk *storage.HMACKey, err error) {
	ctx = internal.StartExitSpan(ctx, "gcs", ot.Tags{
		"gcs.op":        "hmacKeys.update",
		"gcs.projectId": h.ProjectID,
		"gcs.accessId":  h.AccessID,
	})

	defer func() { internal.FinishSpan(ctx, err) }()

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
