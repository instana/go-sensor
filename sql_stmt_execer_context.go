// (c) Copyright IBM Corp. 2023

package instana

import (
	"context"
	"database/sql/driver"

	otlog "github.com/opentracing/opentracing-go/log"
)

type wStmtExecContext struct {
	driver.StmtExecContext
	// connDetails DbConnDetails
	sensor TracerLogger
	// query       string
	sqlSpan *sqlSpanData
}

func (stmt *wStmtExecContext) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	// sp, dbKey := startSQLSpan(ctx, stmt.connDetails, stmt.query, stmt.sensor)
	sp, dbKey := stmt.sqlSpan.start(ctx, stmt.sensor)
	defer sp.Finish()

	res, err := stmt.StmtExecContext.ExecContext(ctx, args)

	if err != nil && err != driver.ErrSkip {
		sp.LogFields(otlog.Error(err))
		sp.SetTag(dbKey+".error", err.Error())
	}

	return res, err
}
