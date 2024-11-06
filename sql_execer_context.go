// (c) Copyright IBM Corp. 2023

package instana

import (
	"context"
	"database/sql/driver"

	otlog "github.com/opentracing/opentracing-go/log"
)

type wExecerContext struct {
	driver.ExecerContext
	sensor TracerLogger

	sqlSpan *sqlSpanData
}

func (conn *wExecerContext) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {

	// updating db query in sqlSpanData instance
	conn.sqlSpan.updateDBQuery(query)

	sp, dbKey := conn.sqlSpan.start(ctx, conn.sensor)
	defer sp.Finish()

	res, err := conn.ExecerContext.ExecContext(ctx, query, args)

	if err != nil && err != driver.ErrSkip {
		sp.LogFields(otlog.Error(err))
		sp.SetTag(dbKey+".error", err.Error())
	}

	return res, err
}
