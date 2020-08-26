package storage

import (
	"context"

	"cloud.google.com/go/storage"
)

// CopierFrom creates a Copier that can copy src to dst.
// You can immediately call Run on the returned Copier, or
// you can configure it first.
//
// For Requester Pays buckets, the user project of dst is billed, unless it is empty,
// in which case the user project of src is billed.
func (dst *ObjectHandle) CopierFrom(src *ObjectHandle) *Copier {
	return &Copier{dst.ObjectHandle.CopierFrom(src.ObjectHandle)}
}

// A Copier copies a source object to a destination.
type Copier struct {
	*storage.Copier
}

// Run performs the copy.
//
// INSTRUMENT
func (c *Copier) Run(ctx context.Context) (attrs *storage.ObjectAttrs, err error) {
	return c.Copier.Run(ctx)
}

// ComposerFrom creates a Composer that can compose srcs into dst.
// You can immediately call Run on the returned Composer, or you can
// configure it first.
//
// The encryption key for the destination object will be used to decrypt all
// source objects and encrypt the destination object. It is an error
// to specify an encryption key for any of the source objects.
func (dst *ObjectHandle) ComposerFrom(srcs ...*ObjectHandle) *Composer {
	srcsCopy := make([]*storage.ObjectHandle, len(srcs))
	for i := range srcs {
		srcsCopy[i] = srcs[i].ObjectHandle
	}

	return &Composer{dst.ObjectHandle.ComposerFrom(srcsCopy...)}
}

// A Composer composes source objects into a destination object.
//
// For Requester Pays buckets, the user project of dst is billed.
type Composer struct {
	*storage.Composer
}

// Run performs the compose operation.
//
// INSTRUMENT
func (c *Composer) Run(ctx context.Context) (attrs *storage.ObjectAttrs, err error) {
	return c.Composer.Run(ctx)
}
