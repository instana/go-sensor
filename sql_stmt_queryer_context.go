// (c) Copyright IBM Corp. 2023
package instana

import (
	"context"
	"database/sql/driver"

	otlog "github.com/opentracing/opentracing-go/log"
)

type wStmtQueryContext struct {
	driver.StmtQueryContext
	connDetails dbConnDetails
	sensor      *Sensor
	query       string
}

func (stmt *wStmtQueryContext) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	sp := startSQLSpan(ctx, stmt.connDetails, stmt.query, stmt.sensor)
	defer sp.Finish()

	res, err := stmt.StmtQueryContext.QueryContext(ctx, args)
	if err != nil && err != driver.ErrSkip {
		sp.LogFields(otlog.Error(err))
	}

	return res, err

}
