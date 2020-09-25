// +build go1.11

package storage

import (
	"context"

	"cloud.google.com/go/storage"
	"github.com/instana/go-sensor/instrumentation/cloud.google.com/go/internal"
	ot "github.com/opentracing/opentracing-go"
)

// ACLHandle is an instrumented wrapper for cloud.google.com/go/storage.ACLHandle
// that traces calls made to Google Cloud Storage API.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#ACLHandle for furter details on wrapped type.
type ACLHandle struct {
	*storage.ACLHandle
	Bucket  string
	Object  string
	Default bool
}

// Delete calls and traces the Delete() method of the wrapped cloud.google.com/go/storage.ACLHandle.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#ACLHandle.Delete for furter details on wrapped method.
func (a *ACLHandle) Delete(ctx context.Context, entity storage.ACLEntity) (err error) {
	ctx = internal.StartExitSpan(ctx, "gcs", ot.Tags{
		"gcs.op":     aclOpPrefix(a) + ".delete",
		"gcs.bucket": a.Bucket,
		"gcs.object": a.Object,
		"gcs.entity": string(entity),
	})

	defer func() { internal.FinishSpan(ctx, err) }()

	return a.ACLHandle.Delete(ctx, entity)
}

// Set calls and traces the Set() method of the wrapped cloud.google.com/go/storage.ACLHandle.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#ACLHandle.Set for furter details on wrapped method.
func (a *ACLHandle) Set(ctx context.Context, entity storage.ACLEntity, role storage.ACLRole) (err error) {
	ctx = internal.StartExitSpan(ctx, "gcs", ot.Tags{
		"gcs.op":     aclOpPrefix(a) + ".update",
		"gcs.bucket": a.Bucket,
		"gcs.object": a.Object,
		"gcs.entity": string(entity),
	})

	defer func() { internal.FinishSpan(ctx, err) }()

	return a.ACLHandle.Set(ctx, entity, role)
}

// List calls and traces the List() method of the wrapped cloud.google.com/go/storage.ACLHandle.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#ACLHandle.List for furter details on wrapped method.
func (a *ACLHandle) List(ctx context.Context) (rules []storage.ACLRule, err error) {
	ctx = internal.StartExitSpan(ctx, "gcs", ot.Tags{
		"gcs.op":     aclOpPrefix(a) + ".list",
		"gcs.bucket": a.Bucket,
		"gcs.object": a.Object,
	})

	defer func() { internal.FinishSpan(ctx, err) }()

	return a.ACLHandle.List(ctx)
}

func aclOpPrefix(a *ACLHandle) string {
	switch {
	case a.Object != "": // object-specific ACL
		return "objectAcls"
	case a.Default: // default object ACL for a bucket
		return "defaultAcls"
	default: // bucket ACL
		return "bucketAcls"
	}
}
