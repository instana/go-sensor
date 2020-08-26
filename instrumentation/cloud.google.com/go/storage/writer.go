package storage

import (
	"cloud.google.com/go/storage"
)

// A Writer writes a Cloud Storage object.
type Writer struct {
	*storage.Writer
}

// Write appends to w. It implements the io.Writer interface.
//
// Since writes happen asynchronously, Write may return a nil
// error even though the write failed (or will fail). Always
// use the error returned from Writer.Close to determine if
// the upload was successful.
//
// Writes will be retried on transient errors from the server, unless
// Writer.ChunkSize has been set to zero.
//
// INSTRUMENT
func (w *Writer) Write(p []byte) (n int, err error) {
	return w.Writer.Write(p)
}
