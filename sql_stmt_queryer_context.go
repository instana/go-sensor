// (c) Copyright IBM Corp. 2023

package instana

import (
	"context"
	"database/sql/driver"

	otlog "github.com/opentracing/opentracing-go/log"
)

type wStmtQueryContext struct {
	driver.StmtQueryContext
	sensor TracerLogger

	sqlSpan *sqlSpanData
}

func (stmt *wStmtQueryContext) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {

	sp, dbKey := stmt.sqlSpan.start(ctx, stmt.sensor)
	defer sp.Finish()

	res, err := stmt.StmtQueryContext.QueryContext(ctx, args)

	if err != nil && err != driver.ErrSkip {
		sp.LogFields(otlog.Error(err))
		sp.SetTag(dbKey+".error", err.Error())
	}

	return res, err

}
