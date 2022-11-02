// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package iam

import (
	"context"
	"strings"

	"github.com/instana/go-sensor/instrumentation/cloud.google.com/go/internal/tags"

	"cloud.google.com/go/iam"
	"github.com/instana/go-sensor/instrumentation/cloud.google.com/go/internal"
	ot "github.com/opentracing/opentracing-go"
)

// Resource describes a Google Cloud IAM resource.
type Resource struct {
	Type string
	Name string
}

// Handle is an instrumented wrapper for cloud.google.com/go/iam.Handle
// that traces calls made to Google Cloud IAM API.
//
// See https://pkg.go.dev/cloud.google.com/go/iam?tab=doc#Handle for further details on wrapped type.
type Handle struct {
	*iam.Handle

	Resource Resource
}

// Handle3 is an instrumented wrapper for cloud.google.com/go/iam.Handle3.
// that traces calls made to Google Cloud IAM API.
//
// See https://pkg.go.dev/cloud.google.com/go/iam?tab=doc#Handle3 for further details on wrapped type.
type Handle3 struct {
	*iam.Handle3

	Resource Resource
}

// V3 returns an instrumented cloud.google.com/go/iam.Handle3.
// that traces requests to the Google Cloud API.
//
// See https://pkg.go.dev/cloud.google.com/go/iam?tab=doc#Handle.V3 for further details on wrapped method
func (h *Handle) V3() *Handle3 {
	return &Handle3{
		Handle3:  h.Handle.V3(),
		Resource: h.Resource,
	}
}

// Policy calls and traces the Policy() method of the wrapped cloud.google.com/go/iam.Handle.
//
// See https://pkg.go.dev/cloud.google.com/go/iam?tab=doc#Handle.Policy for further details on wrapped method
func (h *Handle) Policy(ctx context.Context) (p *iam.Policy, err error) {
	internal.StartExitSpan(ctx, "gcs", ot.Tags{
		tags.GcsOp:                 iamOpPrefix(h.Resource) + ".getIamPolicy",
		iamResourceTag(h.Resource): h.Resource.Name,
	})
	defer func() { internal.FinishSpan(ctx, err) }()

	return h.Handle.Policy(ctx)
}

// SetPolicy calls and traces the SetPolicy() method of the wrapped cloud.google.com/go/iam.Handle.
//
// See https://pkg.go.dev/cloud.google.com/go/iam?tab=doc#Handle.SetPolicy for further details on wrapped method
func (h *Handle) SetPolicy(ctx context.Context, policy *iam.Policy) (err error) {
	internal.StartExitSpan(ctx, "gcs", ot.Tags{
		tags.GcsOp:                 iamOpPrefix(h.Resource) + ".setIamPolicy",
		iamResourceTag(h.Resource): h.Resource.Name,
	})
	defer func() { internal.FinishSpan(ctx, err) }()

	return h.Handle.SetPolicy(ctx, policy)
}

// TestPermissions calls and traces the TestPermissions() method of the wrapped cloud.google.com/go/iam.Handle.
//
// See https://pkg.go.dev/cloud.google.com/go/iam?tab=doc#Handle.TestPermissions for further details on wrapped method
func (h *Handle) TestPermissions(ctx context.Context, permissions []string) (allowed []string, err error) {
	internal.StartExitSpan(ctx, "gcs", ot.Tags{
		tags.GcsOp:                 iamOpPrefix(h.Resource) + ".testIamPermissions",
		iamResourceTag(h.Resource): h.Resource.Name,
	})
	defer func() { internal.FinishSpan(ctx, err) }()

	return h.Handle.TestPermissions(ctx, permissions)
}

// Policy calls and traces the Policy() method of the wrapped cloud.google.com/go/iam.Handle3.
//
// See https://pkg.go.dev/cloud.google.com/go/iam?tab=doc#Handle3.Policy for further details on wrapped method
func (h *Handle3) Policy(ctx context.Context) (p *iam.Policy3, err error) {
	internal.StartExitSpan(ctx, "gcs", ot.Tags{
		tags.GcsOp:                 iamOpPrefix(h.Resource) + ".getIamPolicy",
		iamResourceTag(h.Resource): h.Resource.Name,
	})
	defer func() { internal.FinishSpan(ctx, err) }()

	return h.Handle3.Policy(ctx)
}

// SetPolicy calls and traces the SetPolicy() method of the wrapped cloud.google.com/go/iam.Handle3.
//
// See https://pkg.go.dev/cloud.google.com/go/iam?tab=doc#Handle3.SetPolicy for further details on wrapped method
func (h *Handle3) SetPolicy(ctx context.Context, policy *iam.Policy3) (err error) {
	internal.StartExitSpan(ctx, "gcs", ot.Tags{
		tags.GcsOp:                 iamOpPrefix(h.Resource) + ".setIamPolicy",
		iamResourceTag(h.Resource): h.Resource.Name,
	})
	defer func() { internal.FinishSpan(ctx, err) }()

	return h.Handle3.SetPolicy(ctx, policy)
}

// TestPermissions calls and traces the TestPermissions() method of the wrapped cloud.google.com/go/iam.Handle3.
//
// See https://pkg.go.dev/cloud.google.com/go/iam?tab=doc#Handle3.TestPermissions for further details on wrapped method
func (h *Handle3) TestPermissions(ctx context.Context, permissions []string) (allowed []string, err error) {
	internal.StartExitSpan(ctx, "gcs", ot.Tags{
		tags.GcsOp:                 iamOpPrefix(h.Resource) + ".testIamPermissions",
		iamResourceTag(h.Resource): h.Resource.Name,
	})
	defer func() { internal.FinishSpan(ctx, err) }()

	return h.Handle3.TestPermissions(ctx, permissions)
}

func iamOpPrefix(resource Resource) string {
	switch resource.Type {
	case "bucket":
		return "buckets"
	default:
		return strings.ToLower(resource.Type)
	}
}

func iamResourceTag(resource Resource) string {
	switch resource.Type {
	case "bucket":
		return "gcs.bucket"
	default:
		return "gcs." + strings.ToLower(resource.Type)
	}
}
