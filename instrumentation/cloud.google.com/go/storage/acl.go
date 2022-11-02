// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package storage

import (
	"context"

	"cloud.google.com/go/storage"
	"github.com/instana/go-sensor/instrumentation/cloud.google.com/go/internal"
	"github.com/instana/go-sensor/instrumentation/cloud.google.com/go/internal/tags"
	ot "github.com/opentracing/opentracing-go"
)

// ACLHandle is an instrumented wrapper for cloud.google.com/go/storage.ACLHandle
// that traces calls made to Google Cloud Storage API.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#ACLHandle for further details on wrapped type.
type ACLHandle struct {
	*storage.ACLHandle
	Bucket  string
	Object  string
	Default bool
}

// Delete calls and traces the Delete() method of the wrapped cloud.google.com/go/storage.ACLHandle.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#ACLHandle.Delete for further details on wrapped method.
func (a *ACLHandle) Delete(ctx context.Context, entity storage.ACLEntity) (err error) {
	ctx = internal.StartExitSpan(ctx, "gcs", ot.Tags{
		tags.GcsOp:     aclOpPrefix(a) + ".delete",
		tags.GcsBucket: a.Bucket,
		tags.GcsObject: a.Object,
		tags.GcsEntity: string(entity),
	})

	defer func() { internal.FinishSpan(ctx, err) }()

	return a.ACLHandle.Delete(ctx, entity)
}

// Set calls and traces the Set() method of the wrapped cloud.google.com/go/storage.ACLHandle.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#ACLHandle.Set for further details on wrapped method.
func (a *ACLHandle) Set(ctx context.Context, entity storage.ACLEntity, role storage.ACLRole) (err error) {
	ctx = internal.StartExitSpan(ctx, "gcs", ot.Tags{
		tags.GcsOp:     aclOpPrefix(a) + ".update",
		tags.GcsBucket: a.Bucket,
		tags.GcsObject: a.Object,
		tags.GcsEntity: string(entity),
	})

	defer func() { internal.FinishSpan(ctx, err) }()

	return a.ACLHandle.Set(ctx, entity, role)
}

// List calls and traces the List() method of the wrapped cloud.google.com/go/storage.ACLHandle.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#ACLHandle.List for further details on wrapped method.
func (a *ACLHandle) List(ctx context.Context) (rules []storage.ACLRule, err error) {
	ctx = internal.StartExitSpan(ctx, "gcs", ot.Tags{
		tags.GcsOp:     aclOpPrefix(a) + ".list",
		tags.GcsBucket: a.Bucket,
		tags.GcsObject: a.Object,
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
