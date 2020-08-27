package storage

import (
	"context"

	"cloud.google.com/go/storage"
	"github.com/instana/go-sensor/instrumentation/cloud.google.com/go/internal"
	ot "github.com/opentracing/opentracing-go"
	"google.golang.org/api/iterator"
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

// HMACKeysIterator is an instrumented wrapper for cloud.google.com/go/storage.HMACKeysIterator
// that traces calls made to Google Cloud Storage API
type HMACKeysIterator struct {
	*storage.HMACKeysIterator
	ProjectID string
	ctx       context.Context
}

// ListHMACKeys returns an instrumented object iterator that traces and proxies requests to
// the underlying cloud.google.com/go/storage.HMACKeysIterator
func (c *Client) ListHMACKeys(ctx context.Context, projectID string, opts ...storage.HMACKeyOption) *HMACKeysIterator {
	return &HMACKeysIterator{
		HMACKeysIterator: c.Client.ListHMACKeys(ctx, projectID, opts...),
		ProjectID:        projectID,
		ctx:              ctx,
	}
}

// Next calls the Next() method of the wrapped iterator and creates a span for each call
// that results in an API request
func (it *HMACKeysIterator) Next() (hk *storage.HMACKey, err error) {
	// don't trace calls returning buffered data
	if it.HMACKeysIterator.PageInfo().Remaining() > 0 {
		return it.HMACKeysIterator.Next()
	}

	ctx := internal.StartExitSpan(it.ctx, "gcs", ot.Tags{
		"gcs.op":        "hmacKeys.list",
		"gcs.projectId": it.ProjectID,
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

	return it.HMACKeysIterator.Next()
}
