package storage

import (
	"context"

	"cloud.google.com/go/storage"
)

// ACLHandle provides operations on an access control list for a Google Cloud Storage bucket or object.
type ACLHandle struct {
	*storage.ACLHandle
}

// Delete permanently deletes the ACL entry for the given entity.
//
// INSTRUMENT
func (a *ACLHandle) Delete(ctx context.Context, entity storage.ACLEntity) error {
	return a.ACLHandle.Delete(ctx, entity)
}

// Set sets the role for the given entity.
//
// INSTRUMENT
func (a *ACLHandle) Set(ctx context.Context, entity storage.ACLEntity, role storage.ACLRole) (err error) {
	return a.ACLHandle.Set(ctx, entity, role)
}

// List retrieves ACL entries.
//
// INSTRUMENT
func (a *ACLHandle) List(ctx context.Context) (rules []storage.ACLRule, err error) {
	return a.ACLHandle.List(ctx)
}
