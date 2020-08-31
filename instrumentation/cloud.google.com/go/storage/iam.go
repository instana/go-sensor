package storage

import (
	"github.com/instana/go-sensor/instrumentation/cloud.google.com/go/iam"
)

// IAM returns an instrumented wrapper for cloud.google.com/go/iam.Handle.
//
// See https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#BucketHandle.IAM for furter details on wrapped method.
func (b *BucketHandle) IAM() *iam.Handle {
	return iam.WrapInternalHandle(b.BucketHandle.IAM(), iam.Resource{
		Type: "bucket",
		Name: b.Name,
	})
}
