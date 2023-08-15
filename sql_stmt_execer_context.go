// (c) Copyright IBM Corp. 2023

package instana

import (
	"context"
	"database/sql/driver"

	otlog "github.com/opentracing/opentracing-go/log"
)

type wStmtExecContext struct {
	driver.StmtExecContext
	connDetails DbConnDetails
	sensor      TracerLogger
	query       string
}

func (stmt *wStmtExecContext) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	res, err := stmt.StmtExecContext.ExecContext(ctx, args)

	stmt.connDetails.Error = err

	sp := startSQLSpan(ctx, stmt.connDetails, stmt.query, stmt.sensor)
	defer sp.Finish()

	if err != nil && err != driver.ErrSkip {
		sp.LogFields(otlog.Error(err))
	}

	return res, err
}
