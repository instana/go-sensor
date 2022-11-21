// (c) Copyright IBM Corp. 2022

package instapgx

import (
	"context"
	"fmt"
	"sync"

	instana "github.com/instana/go-sensor"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

var detailedBatchMode = false
var detailedBatchModeMaxEntries = 10

// EnableDetailedBatchMode allows reporting detailed information about sql statements when executing sql in batches.
func EnableDetailedBatchMode() {
	detailedBatchMode = true
}

// Conn wraps *pgx.Conn and adds tracing to it.
type Conn struct {
	*pgx.Conn
	sensor *instana.Sensor
	config *pgx.ConnConfig
}

// Prepare wraps (*pgx.Conn).Prepare method and adds tracing to it.
func (c *Conn) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	childCtx, span := contextWithChildSpan(ctx, fmt.Sprintf("PREPARE %s AS %s", name, sql), c.config, c.sensor)
	defer span.Finish()

	sd, err := c.Conn.Prepare(childCtx, name, sql)
	if err != nil {
		recordAnError(span, err)
	}

	return sd, err
}

// Begin wraps (*pgx.Conn).Begin method and adds tracing to it.
func (c *Conn) Begin(ctx context.Context) (pgx.Tx, error) {
	childCtx, span := contextWithChildSpan(ctx, "BEGIN", c.config, c.sensor)
	defer span.Finish()

	t, err := c.Conn.Begin(childCtx)
	if err != nil {
		recordAnError(span, err)
	}

	return &instaTx{
		Tx:     t,
		sensor: c.sensor,
		config: c.config,
	}, err
}

// BeginTx wraps (*pgx.Conn).BeginTx method and adds tracing to it.
func (c *Conn) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	childCtx, span := contextWithChildSpan(ctx, fmt.Sprintf("BEGIN %s %s %s", txOptions.IsoLevel, txOptions.AccessMode, txOptions.DeferrableMode), c.config, c.sensor)
	defer span.Finish()

	t, err := c.Conn.BeginTx(childCtx, txOptions)
	if err != nil {
		recordAnError(span, err)
	}

	return &instaTx{
		Tx:     t,
		sensor: c.sensor,
		config: c.config,
	}, err
}

// BeginFunc wraps (*pgx.Conn).BeginFunc method and adds tracing to it.
func (c *Conn) BeginFunc(ctx context.Context, f func(pgx.Tx) error) error {
	return c.BeginTxFunc(ctx, pgx.TxOptions{}, f)
}

// BeginTxFunc wraps (*pgx.Conn).BeginTxFunc method and adds tracing to it.
func (c *Conn) BeginTxFunc(ctx context.Context, txOptions pgx.TxOptions, f func(pgx.Tx) error) error {
	childCtx, span := contextWithChildSpan(ctx, fmt.Sprintf("BEGIN %s %s %s", txOptions.IsoLevel, txOptions.DeferrableMode, txOptions.AccessMode), c.config, c.sensor)

	wf := func(t pgx.Tx) error {
		fErr := f(&instaTx{
			Tx:          t,
			sensor:      c.sensor,
			config:      c.config,
			ctxWithSpan: childCtx,
		})

		if fErr != nil {
			_, span := contextWithChildSpan(childCtx, "ROLLBACK", c.config, c.sensor)
			span.Finish()
		} else {
			_, span := contextWithChildSpan(childCtx, "COMMIT", c.config, c.sensor)
			span.Finish()
		}

		return fErr
	}

	span.Finish()
	err := c.Conn.BeginTxFunc(childCtx, txOptions, wf)
	if err != nil {
		recordAnError(span, err)
	}

	return err
}

// CopyFrom wraps (*pgx.Conn).CopyFrom method and adds tracing to it. It doesn't provide details of the call.
func (c *Conn) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	childCtx, span := contextWithChildSpan(ctx, "PGCOPY", c.config, c.sensor)
	defer span.Finish()

	n, err := c.Conn.CopyFrom(childCtx, tableName, columnNames, rowSrc)
	if err != nil {
		recordAnError(span, err)
	}

	return n, err
}

// Ping wraps (*pgx.Conn).Ping method and adds tracing to it.
func (c *Conn) Ping(ctx context.Context) error {
	childCtx, span := contextWithChildSpan(ctx, ";", c.config, c.sensor)
	defer span.Finish()

	err := c.Conn.Ping(childCtx)
	if err != nil {
		recordAnError(span, err)
	}

	return err
}

// Exec wraps (*pgx.Conn).Exec method and adds tracing to it.
func (c *Conn) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	childCtx, span := contextWithChildSpan(ctx, sql, c.config, c.sensor)
	defer span.Finish()

	tags, err := c.Conn.Exec(childCtx, sql, arguments...)
	if err != nil {
		recordAnError(span, err)
	}

	return tags, err
}

// Query wraps (*pgx.Conn).Query method and adds tracing to it.
func (c *Conn) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	childCtx, span := contextWithChildSpan(ctx, sql, c.config, c.sensor)

	defer span.Finish()

	rows, err := c.Conn.Query(childCtx, sql, args...)
	if err != nil {
		recordAnError(span, err)
	}

	return rows, err
}

// QueryRow wraps (*pgx.Conn).QueryRow method and adds tracing to it.
// It requires to calling Scan method on the returned row to finish a span.
func (c *Conn) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	childCtx, span := contextWithChildSpan(ctx, sql, c.config, c.sensor)

	r := c.Conn.QueryRow(childCtx, sql, args...)
	return &row{
		Row:  r,
		span: span,
		m:    &sync.Mutex{},
	}
}

// QueryFunc wraps (*pgx.Conn).QueryFunc method and adds tracing to it.
func (c *Conn) QueryFunc(ctx context.Context, sql string, args []interface{}, scans []interface{}, f func(pgx.QueryFuncRow) error) (pgconn.CommandTag, error) {
	childCtx, span := contextWithChildSpan(ctx, sql, c.config, c.sensor)
	defer span.Finish()

	tags, err := c.Conn.QueryFunc(childCtx, sql, args, scans, f)
	if err != nil {
		recordAnError(span, err)
	}

	return tags, err
}

// SendBatch wraps (*pgx.Conn).SendBatch method and adds tracing to it. Call the EnableDetailedBatchMode(), to have
// sql statements in the span. Amount of the sql statements that might be reported are limited.
func (c *Conn) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	sql := "-- SEND BATCH"
	if detailedBatchMode {
		sql = appendBatchDetails(b, sql)
	}

	childCtx, span := contextWithChildSpan(ctx, sql, c.config, c.sensor)
	defer span.Finish()

	br := c.Conn.SendBatch(childCtx, b)

	return &batchResults{
		br,
		childCtx,
		c.sensor,
		c.config,
	}
}
