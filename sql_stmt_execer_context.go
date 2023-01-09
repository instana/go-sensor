// (c) Copyright IBM Corp. 2023
package instana

import (
	"context"
	"database/sql/driver"

	otlog "github.com/opentracing/opentracing-go/log"
)

type wStmtExecContext struct {
	driver.StmtExecContext
	connDetails dbConnDetails
	sensor      *Sensor
	query       string
}

func (stmt *wStmtExecContext) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	sp := startSQLSpan(ctx, stmt.connDetails, stmt.query, stmt.sensor)
	defer sp.Finish()

	res, err := stmt.StmtExecContext.ExecContext(ctx, args)
	if err != nil && err != driver.ErrSkip {
		sp.LogFields(otlog.Error(err))
	}

	return res, err
}
