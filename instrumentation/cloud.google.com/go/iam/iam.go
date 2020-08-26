package iam

import (
	"context"

	"cloud.google.com/go/iam"
)

// A Handle provides IAM operations for a resource.
type Handle struct {
	*iam.Handle
}

// A Handle3 provides IAM operations for a resource. It is similar to a Handle, but provides access to newer IAM features (e.g., conditions).
type Handle3 struct {
	*iam.Handle3
}

// InternalNewHandle is for use by the Google Cloud Libraries only.
//
// InternalNewHandle returns a Handle for resource.
// The conn parameter refers to a server that must support the IAMPolicy service.
func WrapInternalHandle(h *iam.Handle) *Handle {
	return &Handle{h}
}

// V3 returns a Handle3, which is like Handle except it sets
// requestedPolicyVersion to 3 when retrieving a policy and policy.version to 3
// when storing a policy.
func (h *Handle) V3() *Handle3 {
	return &Handle3{h.Handle.V3()}
}

// Policy retrieves the IAM policy for the resource.
//
// INSTRUMENT
func (h *Handle) Policy(ctx context.Context) (*iam.Policy, error) {
	return h.Handle.Policy(ctx)
}

// SetPolicy replaces the resource's current policy with the supplied Policy.
//
// If policy was created from a prior call to Get, then the modification will
// only succeed if the policy has not changed since the Get.
//
// INSTRUMENT
func (h *Handle) SetPolicy(ctx context.Context, policy *iam.Policy) error {
	return h.Handle.SetPolicy(ctx, policy)
}

// TestPermissions returns the subset of permissions that the caller has on the resource.
//
// INSTRUMENT
func (h *Handle) TestPermissions(ctx context.Context, permissions []string) ([]string, error) {
	return h.Handle.TestPermissions(ctx, permissions)
}

// Policy retrieves the IAM policy for the resource.
//
// requestedPolicyVersion is always set to 3.
//
// INSTRUMENT
func (h *Handle3) Policy(ctx context.Context) (*iam.Policy3, error) {
	return h.Handle3.Policy(ctx)
}

// SetPolicy replaces the resource's current policy with the supplied Policy.
//
// If policy was created from a prior call to Get, then the modification will
// only succeed if the policy has not changed since the Get.
//
// INSTRUMENT
func (h *Handle3) SetPolicy(ctx context.Context, policy *iam.Policy3) error {
	return h.Handle3.SetPolicy(ctx, policy)
}

// TestPermissions returns the subset of permissions that the caller has on the resource.
//
// INSTRUMENT
func (h *Handle3) TestPermissions(ctx context.Context, permissions []string) ([]string, error) {
	return h.Handle3.TestPermissions(ctx, permissions)
}
