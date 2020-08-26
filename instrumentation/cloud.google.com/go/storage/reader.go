package storage

import (
	"context"

	"cloud.google.com/go/storage"
)

// NewReader creates a new Reader to read the contents of the
// object.
// ErrObjectNotExist will be returned if the object is not found.
//
// The caller must call Close on the returned Reader when done reading.
func (o *ObjectHandle) NewReader(ctx context.Context) (*Reader, error) {
	return o.NewRangeReader(ctx, 0, -1)
}

// NewRangeReader reads part of an object, reading at most length bytes
// starting at the given offset. If length is negative, the object is read
// until the end. If offset is negative, the object is read abs(offset) bytes
// from the end, and length must also be negative to indicate all remaining
// bytes will be read.
//
// If the object's metadata property "Content-Encoding" is set to "gzip" or satisfies
// decompressive transcoding per https://cloud.google.com/storage/docs/transcoding
// that file will be served back whole, regardless of the requested range as
// Google Cloud Storage dictates.
//
// INSTRUMENT
func (o *ObjectHandle) NewRangeReader(ctx context.Context, offset, length int64) (r *Reader, err error) {
	rdr, err := o.ObjectHandle.NewRangeReader(ctx, offset, length)
	return &Reader{rdr}, err
}

// Reader reads a Cloud Storage object.
// It implements io.Reader.
//
// Typically, a Reader computes the CRC of the downloaded content and compares it to
// the stored CRC, returning an error from Read if there is a mismatch. This integrity check
// is skipped if transcoding occurs. See https://cloud.google.com/storage/docs/transcoding.
type Reader struct {
	*storage.Reader
}

// INSTRUMENT
func (r *Reader) Read(p []byte) (int, error) {
	return r.Reader.Read(p)
}
