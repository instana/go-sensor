// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instana

import (
	"context"
	"database/sql/driver"
	otlog "github.com/opentracing/opentracing-go/log"
)

type wSQLStmt interface {
	driver.Stmt
	Details() dbConnDetails
	Sensor() *Sensor
	GetQuery() string
	GetStmt() driver.Stmt
}

type wrappedSQLStmt struct {
	driver.Stmt

	connDetails dbConnDetails
	query       string
	sensor      *Sensor
}

func (stmt *wrappedSQLStmt) Details() dbConnDetails {
	return stmt.connDetails
}

func (stmt *wrappedSQLStmt) Sensor() *Sensor {
	return stmt.sensor
}

func (stmt *wrappedSQLStmt) GetQuery() string {
	return stmt.query
}

func (stmt *wrappedSQLStmt) GetStmt() driver.Stmt {
	if v, ok := stmt.Stmt.(wSQLStmt); ok {
		return v.GetStmt()
	}

	return stmt.Stmt
}

type wStmtQueryContext struct {
	wSQLStmt
}

func (stmt *wStmtQueryContext) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	sp := startSQLSpan(ctx, stmt.Details(), stmt.GetQuery(), stmt.Sensor())
	defer sp.Finish()

	if s, ok := stmt.GetStmt().(driver.StmtQueryContext); ok {
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

	res, err := stmt.GetStmt().Query(values) //nolint:staticcheck
	if err != nil && err != driver.ErrSkip {
		sp.LogFields(otlog.Error(err))
	}

	return res, err
}

type wStmtNamedValueChecker struct {
	wSQLStmt
}

func (stmt *wStmtNamedValueChecker) CheckNamedValue(d *driver.NamedValue) error {
	if s, ok := stmt.wSQLStmt.(driver.NamedValueChecker); ok {
		return s.CheckNamedValue(d)
	}

	return nil
}

type wStmtExecContext struct {
	wSQLStmt
}

func (stmt *wStmtExecContext) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	sp := startSQLSpan(ctx, stmt.Details(), stmt.GetQuery(), stmt.Sensor())
	defer sp.Finish()

	if s, ok := stmt.GetStmt().(driver.StmtExecContext); ok {
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

	res, err := stmt.GetStmt().Exec(values) //nolint:staticcheck
	if err != nil && err != driver.ErrSkip {
		sp.LogFields(otlog.Error(err))
	}

	return res, err
}
