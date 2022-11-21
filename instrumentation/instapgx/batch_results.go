// (c) Copyright IBM Corp. 2022

package instapgx

import (
	"context"
	"sync"

	instana "github.com/instana/go-sensor"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

// batchResults is a wrapper for pgx.batchResults
type batchResults struct {
	pgx.BatchResults
	ctx    context.Context
	sensor *instana.Sensor
	config *pgx.ConnConfig
}

// Exec wraps (*pgx.batchResults).Exec and adds tracing to it.
func (b *batchResults) Exec() (pgconn.CommandTag, error) {
	_, span := contextWithChildSpan(b.ctx, "-- BatchResults EXEC", b.config, b.sensor)
	defer span.Finish()

	ct, err := b.BatchResults.Exec()
	if err != nil {
		recordAnError(span, err)
	}

	return ct, err
}

// Query wraps (*pgx.batchResults).Query and adds tracing to it.
func (b *batchResults) Query() (pgx.Rows, error) {
	_, span := contextWithChildSpan(b.ctx, "-- BatchResults QUERY", b.config, b.sensor)
	defer span.Finish()

	rows, err := b.BatchResults.Query()
	if err != nil {
		recordAnError(span, err)
	}

	return rows, err
}

// QueryRow wraps (*pgx.batchResults).QueryRow and adds tracing to it.
func (b *batchResults) QueryRow() pgx.Row {
	_, span := contextWithChildSpan(b.ctx, "-- BatchResults QUERY ROW", b.config, b.sensor)

	return &row{
		Row:  b.BatchResults.QueryRow(),
		span: span,
		m:    &sync.Mutex{},
	}
}

// QueryFunc wraps (*pgx.batchResults).QueryFunc and adds tracing to it.
func (b *batchResults) QueryFunc(scans []interface{}, f func(pgx.QueryFuncRow) error) (pgconn.CommandTag, error) {
	_, span := contextWithChildSpan(b.ctx, "-- BatchResults QUERY FUNC", b.config, b.sensor)
	defer span.Finish()

	ct, err := b.BatchResults.QueryFunc(scans, f)
	if err != nil {
		recordAnError(span, err)
	}

	return ct, err
}
