// (c) Copyright IBM Corp. 2022

package instapgx

import (
	"sync"

	"github.com/jackc/pgx/v4"
	"github.com/opentracing/opentracing-go"
)

type row struct {
	pgx.Row
	span         opentracing.Span
	m            *sync.Mutex
	spanFinished bool
}

// Scan reads the result returned by (*instapgx.Conn).QueryRow.
// It is required to call Scan to finish a span and detect errors from (*instapgx.Conn).QueryRow if any.
func (r *row) Scan(dest ...interface{}) error {
	r.m.Lock()
	err := r.Row.Scan(dest...)
	if !r.spanFinished {
		if err != nil {
			recordAnError(r.span, err)
		}

		r.spanFinished = true
		r.span.Finish()
	}
	r.m.Unlock()

	return err
}
