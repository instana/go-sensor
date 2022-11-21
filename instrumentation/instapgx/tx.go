// (c) Copyright IBM Corp. 2022

package instapgx

import (
	"context"
	"fmt"
	"sync"

	instana "github.com/instana/go-sensor"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/opentracing/opentracing-go"
)

type instaTx struct {
	pgx.Tx
	sensor      *instana.Sensor
	config      *pgx.ConnConfig
	ctxWithSpan context.Context
}

// Begin wraps (*pgx.dbTx).Begin and adds tracing to it.
func (iTx *instaTx) Begin(ctx context.Context) (pgx.Tx, error) {
	var span opentracing.Span
	var childCtx context.Context

	childCtx, span = iTx.contextWithChildSpan(ctx, childCtx, span)
	defer span.Finish()

	tt, err := iTx.Tx.Begin(ctx)
	if err != nil {
		recordAnError(span, err)
	}

	return &instaTx{
		Tx:          tt,
		sensor:      iTx.sensor,
		config:      iTx.config,
		ctxWithSpan: childCtx,
	}, err
}

// BeginFunc wraps (*pgx.dbTx).BeginFunc and adds tracing to it.
func (iTx *instaTx) BeginFunc(ctx context.Context, f func(pgx.Tx) error) error {
	var span opentracing.Span
	var childCtx context.Context

	childCtx, span = iTx.contextWithChildSpan(ctx, childCtx, span)

	wf := func(tt pgx.Tx) error {
		fErr := f(&instaTx{
			Tx:          tt,
			sensor:      iTx.sensor,
			config:      iTx.config,
			ctxWithSpan: childCtx,
		})

		if fErr != nil {
			_, span := contextWithChildSpan(childCtx, "ROLLBACK", iTx.config, iTx.sensor)
			span.Finish()
		} else {
			_, span := contextWithChildSpan(childCtx, "COMMIT", iTx.config, iTx.sensor)
			span.Finish()
		}

		return fErr
	}

	span.Finish()
	err := iTx.Tx.BeginFunc(ctx, wf)
	if err != nil {
		recordAnError(span, err)
	}

	return err
}

// Commit wraps (*pgx.dbTx).Commit and adds tracing to it.
func (iTx *instaTx) Commit(ctx context.Context) error {
	var span opentracing.Span

	if iTx.ctxWithSpan != nil {
		_, span = contextWithChildSpan(iTx.ctxWithSpan, "COMMIT", iTx.config, iTx.sensor)
	} else {
		_, span = contextWithChildSpan(ctx, "COMMIT", iTx.config, iTx.sensor)
	}
	defer span.Finish()

	err := iTx.Tx.Commit(ctx)
	if err != nil {
		if err == pgx.ErrTxCommitRollback {
			span.SetTag("pg.stmt", "ROLLBACK")
		}
		recordAnError(span, err)
	}

	return err
}

// Rollback wraps (*pgx.dbTx).Rollback and adds tracing to it.
func (iTx *instaTx) Rollback(ctx context.Context) error {
	var span opentracing.Span

	if iTx.ctxWithSpan != nil {
		_, span = contextWithChildSpan(iTx.ctxWithSpan, "ROLLBACK", iTx.config, iTx.sensor)
	} else {
		_, span = contextWithChildSpan(ctx, "ROLLBACK", iTx.config, iTx.sensor)
	}
	defer span.Finish()

	err := iTx.Tx.Rollback(ctx)
	if err != nil {
		recordAnError(span, err)
	}

	return err
}

// CopyFrom wraps (*pgx.dbTx).CopyFrom and adds tracing to it. It does not provide details about the operation.
func (iTx *instaTx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	var span opentracing.Span

	if iTx.ctxWithSpan != nil {
		_, span = contextWithChildSpan(iTx.ctxWithSpan, "PGCOPY", iTx.config, iTx.sensor)
	} else {
		_, span = contextWithChildSpan(ctx, "PGCOPY", iTx.config, iTx.sensor)
	}

	defer span.Finish()

	n, err := iTx.Tx.CopyFrom(ctx, tableName, columnNames, rowSrc)
	if err != nil {
		recordAnError(span, err)
	}

	return n, err
}

// SendBatch wraps (*pgx.dbTx).SendBatch and adds tracing to it. Call the EnableDetailedBatchMode(), to have
// sql statements in the span. Amount of the sql that might be reported are limited by detailedBatchModeMaxEntries var.
func (iTx *instaTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	sql := "-- SEND BATCH"
	if detailedBatchMode {
		sql = appendBatchDetails(b, sql)
	}

	var span opentracing.Span
	var childCtx context.Context

	if iTx.ctxWithSpan != nil {
		childCtx, span = contextWithChildSpan(iTx.ctxWithSpan, sql, iTx.config, iTx.sensor)
	} else {
		childCtx, span = contextWithChildSpan(ctx, sql, iTx.config, iTx.sensor)
	}
	defer span.Finish()

	br := iTx.Tx.SendBatch(ctx, b)

	return &batchResults{
		BatchResults: br,
		ctx:          childCtx,
		config:       iTx.config,
		sensor:       iTx.sensor,
	}
}

// Prepare wraps (*pgx.dbTx).Prepare and adds tracing to it.
func (iTx *instaTx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	_, span := contextWithChildSpan(ctx, fmt.Sprintf("PREPARE %s AS %s", name, sql), iTx.config, iTx.sensor)
	defer span.Finish()

	sd, err := iTx.Tx.Prepare(ctx, name, sql)
	if err != nil {
		recordAnError(span, err)
	}

	return sd, err
}

// Exec wraps (*pgx.dbTx).Exec and adds tracing to it.
func (iTx *instaTx) Exec(ctx context.Context, sql string, arguments ...interface{}) (commandTag pgconn.CommandTag, err error) {
	_, span := contextWithChildSpan(ctx, sql, iTx.config, iTx.sensor)
	defer span.Finish()

	tags, err := iTx.Tx.Exec(ctx, sql, arguments...)
	if err != nil {
		recordAnError(span, err)
	}
	return tags, err
}

// Query wraps (*pgx.dbTx).Query and adds tracing to it.
func (iTx *instaTx) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	_, span := contextWithChildSpan(ctx, sql, iTx.config, iTx.sensor)
	defer span.Finish()

	rows, err := iTx.Tx.Query(ctx, sql, args...)
	if err != nil {
		recordAnError(span, err)
	}
	return rows, err
}

// QueryRow wraps (*pgx.dbTx).QueryRow method and adds tracing to it.
// It requires to calling Scan method on the returned row to finish a span.
func (iTx *instaTx) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	_, span := contextWithChildSpan(ctx, sql, iTx.config, iTx.sensor)

	r := iTx.Tx.QueryRow(ctx, sql, args...)
	return &row{
		Row:  r,
		span: span,
		m:    &sync.Mutex{},
	}
}

// QueryFunc wraps (*pgx.dbTx).QueryFunc method and adds tracing to it.
func (iTx *instaTx) QueryFunc(ctx context.Context, sql string, args []interface{}, scans []interface{}, f func(pgx.QueryFuncRow) error) (pgconn.CommandTag, error) {
	var span opentracing.Span
	if iTx.ctxWithSpan != nil {
		_, span = contextWithChildSpan(iTx.ctxWithSpan, sql, iTx.config, iTx.sensor)
	} else {
		_, span = contextWithChildSpan(ctx, sql, iTx.config, iTx.sensor)
	}
	defer span.Finish()

	tags, err := iTx.Tx.QueryFunc(ctx, sql, args, scans, f)
	if err != nil {
		recordAnError(span, err)
	}
	return tags, err
}

func (iTx *instaTx) contextWithChildSpan(ctx context.Context, childCtx context.Context, span opentracing.Span) (context.Context, opentracing.Span) {
	if iTx.ctxWithSpan != nil {
		childCtx, span = contextWithChildSpan(iTx.ctxWithSpan, "BEGIN", iTx.config, iTx.sensor)
	} else {
		childCtx, span = contextWithChildSpan(ctx, "BEGIN", iTx.config, iTx.sensor)
	}
	return childCtx, span
}
