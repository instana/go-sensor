// (c) Copyright IBM Corp. 2023

package instana

import (
	"context"
	"database/sql/driver"

	otlog "github.com/opentracing/opentracing-go/log"
)

type wStmt struct {
	driver.Stmt

	sqlSpan *sqlSpanData
	sensor  TracerLogger
}

func (stmt *wStmt) Exec(args []driver.Value) (driver.Result, error) {
	ctx := context.Background()

	sp, dbKey := stmt.sqlSpan.start(ctx, stmt.sensor)
	defer sp.Finish()

	res, err := stmt.Stmt.Exec(args) //nolint:staticcheck
	if err != nil && err != driver.ErrSkip {
		sp.LogFields(otlog.Error(err))
		sp.SetTag(dbKey+".error", err.Error())
	}

	return res, err
}

func (stmt *wStmt) Query(args []driver.Value) (driver.Rows, error) {
	ctx := context.Background()

	sp, dbKey := stmt.sqlSpan.start(ctx, stmt.sensor)
	defer sp.Finish()

	res, err := stmt.Stmt.Query(args) //nolint:staticcheck
	if err != nil && err != driver.ErrSkip {
		sp.LogFields(otlog.Error(err))
		sp.SetTag(dbKey+".error", err.Error())
	}

	return res, err
}
