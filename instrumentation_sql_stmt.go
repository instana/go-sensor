// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instana

import (
	"context"
	"database/sql/driver"
	otlog "github.com/opentracing/opentracing-go/log"
)

type wrappedSQLStmt struct {
	driver.Stmt

	connDetails dbConnDetails
	query       string
	sensor      *Sensor
}

func (stmt *wrappedSQLStmt) Exec(args []driver.Value) (driver.Result, error) {
	ctx := context.Background()
	sp := startSQLSpan(ctx, stmt.connDetails, stmt.query, stmt.sensor)
	defer sp.Finish()

	res, err := stmt.Stmt.Exec(args) //nolint:staticcheck
	if err != nil && err != driver.ErrSkip {
		sp.LogFields(otlog.Error(err))
	}

	return res, err
}

func (stmt *wrappedSQLStmt) Query(args []driver.Value) (driver.Rows, error) {
	ctx := context.Background()
	sp := startSQLSpan(ctx, stmt.connDetails, stmt.query, stmt.sensor)
	defer sp.Finish()

	res, err := stmt.Stmt.Query(args) //nolint:staticcheck
	if err != nil && err != driver.ErrSkip {
		sp.LogFields(otlog.Error(err))
	}

	return res, err
}
