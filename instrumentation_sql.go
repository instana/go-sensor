package instana

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"strings"

	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
)

// InstrumentSQLDriver instruments provided database driver for  use with `sql.Open()`.
// The instrumented version is registered with `_with_instana` suffix, e.g.
// if `postgres` provided as a name, the instrumented version is registered as
// `postgres_with_instana`.
func InstrumentSQLDriver(sensor *Sensor, name string, driver driver.Driver) {
	sql.Register(name+"_with_instana", &wrappedSQLDriver{
		Driver: driver,
		sensor: sensor,
	})
}

// OpenSQLDB is a convenience wrapper for `sql.Open()` to use the instrumented version
// of a driver previosly registered using `instana.InstrumentSQLDriver()`
func OpenSQLDB(driverName, dataSourceName string) (*sql.DB, error) {
	if !strings.HasSuffix(driverName, "_with_instana") {
		driverName += "_with_instana"
	}

	return sql.Open(driverName, dataSourceName)
}

type wrappedSQLDriver struct {
	driver.Driver

	sensor *Sensor
}

func (drv *wrappedSQLDriver) Open(name string) (driver.Conn, error) {
	conn, err := drv.Driver.Open(name)
	if err != nil {
		return conn, err
	}

	if conn, ok := conn.(*wrappedSQLConn); ok {
		return conn, nil
	}

	return &wrappedSQLConn{
		Conn:       conn,
		connString: name,
		sensor:     drv.sensor,
	}, nil
}

type wrappedSQLConn struct {
	driver.Conn

	connString string
	sensor     *Sensor
}

func (conn *wrappedSQLConn) Prepare(query string) (driver.Stmt, error) {
	stmt, err := conn.Conn.Prepare(query)
	if err != nil {
		return stmt, err
	}

	if stmt, ok := stmt.(*wrappedSQLStmt); ok {
		return stmt, nil
	}

	return &wrappedSQLStmt{
		Stmt:       stmt,
		connString: conn.connString,
		query:      query,
		sensor:     conn.sensor,
	}, nil
}

func (conn *wrappedSQLConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	var (
		stmt driver.Stmt
		err  error
	)
	if c, ok := conn.Conn.(driver.ConnPrepareContext); ok {
		stmt, err = c.PrepareContext(ctx, query)
	} else {
		stmt, err = conn.Prepare(query)
	}

	if err != nil {
		return stmt, err
	}

	if stmt, ok := stmt.(*wrappedSQLStmt); ok {
		return stmt, nil
	}

	return &wrappedSQLStmt{
		Stmt:       stmt,
		connString: conn.connString,
		query:      query,
		sensor:     conn.sensor,
	}, nil
}

func (conn *wrappedSQLConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	sp := startSQLSpan(ctx, conn.connString, query, conn.sensor)
	defer sp.Finish()

	if c, ok := conn.Conn.(driver.ExecerContext); ok {
		res, err := c.ExecContext(ctx, query, args)
		if err != nil && err != driver.ErrSkip {
			sp.LogFields(otlog.Error(err))
		}

		return res, err
	}

	if c, ok := conn.Conn.(driver.Execer); ok { //nolint:staticcheck
		values, err := sqlNamedValuesToValues(args)
		if err != nil {
			return nil, err
		}

		select {
		default:
		case <-ctx.Done():
			return nil, ctx.Err()
		}

		res, err := c.Exec(query, values)
		if err != nil && err != driver.ErrSkip {
			sp.LogFields(otlog.Error(err))
		}

		return res, err
	}

	return nil, driver.ErrSkip
}

type wrappedSQLStmt struct {
	driver.Stmt

	connString string
	query      string
	sensor     *Sensor
}

func (stmt *wrappedSQLStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	sp := startSQLSpan(ctx, stmt.connString, stmt.query, stmt.sensor)
	defer sp.Finish()

	if s, ok := stmt.Stmt.(driver.StmtExecContext); ok {
		res, err := s.ExecContext(ctx, args)
		if err != nil && err != driver.ErrSkip {
			sp.LogFields(otlog.Error(err))
		}

		return res, err
	}

	values, err := sqlNamedValuesToValues(args)
	if err != nil {
		return nil, err
	}

	select {
	default:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	res, err := stmt.Exec(values) //nolint:staticcheck
	if err != nil && err != driver.ErrSkip {
		sp.LogFields(otlog.Error(err))
	}

	return res, err
}

func (stmt *wrappedSQLStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	sp := startSQLSpan(ctx, stmt.connString, stmt.query, stmt.sensor)
	defer sp.Finish()

	if s, ok := stmt.Stmt.(driver.StmtQueryContext); ok {
		res, err := s.QueryContext(ctx, args)
		if err != nil && err != driver.ErrSkip {
			sp.LogFields(otlog.Error(err))
		}

		return res, err
	}

	values, err := sqlNamedValuesToValues(args)
	if err != nil {
		return nil, err
	}

	select {
	default:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	res, err := stmt.Stmt.Query(values) //nolint:staticcheck
	if err != nil && err != driver.ErrSkip {
		sp.LogFields(otlog.Error(err))
	}

	return res, err
}

func startSQLSpan(ctx context.Context, connString, query string, sensor *Sensor) ot.Span {
	opts := []ot.StartSpanOption{
		ext.SpanKindRPCClient,
		ot.Tags{
			string(ext.DBType):      "sql",
			string(ext.DBStatement): query,
			string(ext.DBInstance):  connString,
		},
	}
	if parentSpan, ok := SpanFromContext(ctx); ok {
		opts = append(opts, ot.ChildOf(parentSpan.Context()))
	}

	return sensor.Tracer().StartSpan("sdk.database", opts...)
}

// The following code is ported from $GOROOT/src/database/sql/ctxutil.go
//
// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
func sqlNamedValuesToValues(named []driver.NamedValue) ([]driver.Value, error) {
	dargs := make([]driver.Value, len(named))
	for n, param := range named {
		if len(param.Name) > 0 {
			return nil, errors.New("sql: driver does not support the use of Named Parameters")
		}
		dargs[n] = param.Value
	}
	return dargs, nil
}

type dsnConnector struct {
	dsn    string
	driver driver.Driver
}

func (t dsnConnector) Connect(_ context.Context) (driver.Conn, error) {
	return t.driver.Open(t.dsn)
}

func (t dsnConnector) Driver() driver.Driver {
	return t.driver
}
