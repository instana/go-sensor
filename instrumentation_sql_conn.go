// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instana

import (
	"context"
	"database/sql/driver"
	otlog "github.com/opentracing/opentracing-go/log"
)

type wConn interface {
	driver.Conn
	Details() dbConnDetails
	Sensor() *Sensor
	GetConn() driver.Conn
}

type wrappedSQLConn struct {
	driver.Conn

	details dbConnDetails
	sensor  *Sensor
}

func (conn *wrappedSQLConn) Details() dbConnDetails {
	return conn.details
}

func (conn *wrappedSQLConn) Sensor() *Sensor {
	return conn.sensor
}

func (conn *wrappedSQLConn) GetConn() driver.Conn {
	if v, ok := conn.Conn.(wConn); ok {
		return v.GetConn()
	}

	return conn.Conn
}

func (conn *wrappedSQLConn) Prepare(query string) (driver.Stmt, error) {
	stmt, err := conn.GetConn().Prepare(query)
	if err != nil {
		return stmt, err
	}

	switch stmt.(type) {
	case wConn:
		return stmt, nil
	}

	w := conn.wrap(query, stmt)

	return w, nil
}

func (conn *wrappedSQLConn) wrap(query string, stmt driver.Stmt) wSQLStmt {
	var w wSQLStmt
	w = &wrappedSQLStmt{
		Stmt:        stmt,
		connDetails: conn.details,
		query:       query,
		sensor:      conn.sensor,
	}

	if _, ok := stmt.(driver.NamedValueChecker); ok {
		w = &wStmtNamedValueChecker{
			w,
		}
	}

	if _, ok := stmt.(driver.StmtQueryContext); ok {
		w = &wStmtQueryContext{
			w,
		}
	}

	if _, ok := stmt.(driver.StmtExecContext); ok {
		w = &wStmtExecContext{
			w,
		}
	}

	return w
}

type wNamedValueChecker struct {
	wConn
}

func (conn *wNamedValueChecker) CheckNamedValue(d *driver.NamedValue) error {
	if c, ok := conn.wConn.(driver.NamedValueChecker); ok {
		return c.CheckNamedValue(d)
	}

	return nil
}

type wQueryerContext struct {
	wConn
}

func (conn *wQueryerContext) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	sp := startSQLSpan(ctx, conn.Details(), query, conn.Sensor())
	defer sp.Finish()

	if c, ok := conn.GetConn().(driver.QueryerContext); ok {
		res, err := c.QueryContext(ctx, query, args)
		if err != nil && err != driver.ErrSkip {
			sp.LogFields(otlog.Error(err))
		}

		return res, err
	}

	if c, ok := conn.GetConn().(driver.Queryer); ok { //nolint:staticcheck
		values, err := sqlNamedValuesToValues(args)
		if err != nil {
			return nil, err
		}

		select {
		default:
		case <-ctx.Done():
			return nil, ctx.Err()
		}

		res, err := c.Query(query, values)
		if err != nil && err != driver.ErrSkip {
			sp.LogFields(otlog.Error(err))
		}

		return res, err
	}

	return nil, driver.ErrSkip
}

type wConnPrepareContext struct {
	wConn
}

func (conn *wConnPrepareContext) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	var (
		stmt driver.Stmt
		err  error
	)
	if c, ok := conn.GetConn().(driver.ConnPrepareContext); ok {
		stmt, err = c.PrepareContext(ctx, query)
	} else {
		stmt, err = conn.Prepare(query)
	}

	if err != nil {
		return stmt, err
	}

	if _, ok := stmt.(wSQLStmt); ok {
		return stmt, nil
	}

	return &wrappedSQLStmt{
		Stmt:        stmt,
		connDetails: conn.Details(),
		query:       query,
		sensor:      conn.Sensor(),
	}, nil
}

type wExecerContext struct {
	wConn
}

func (conn *wExecerContext) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	sp := startSQLSpan(ctx, conn.Details(), query, conn.Sensor())
	defer sp.Finish()

	if c, ok := conn.GetConn().(driver.ExecerContext); ok {
		res, err := c.ExecContext(ctx, query, args)
		if err != nil && err != driver.ErrSkip {
			sp.LogFields(otlog.Error(err))
		}

		return res, err
	}

	if c, ok := conn.GetConn().(driver.Execer); ok { //nolint:staticcheck
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
