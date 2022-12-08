// (c) Copyright IBM Corp. 2022

package instana

import (
	"context"
	"database/sql/driver"
	otlog "github.com/opentracing/opentracing-go/log"
)

type wQueryerContext struct {
	originalConn driver.QueryerContext
	driver.Conn
	connDetails dbConnDetails
	sensor      *Sensor
}

func (conn *wQueryerContext) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	sp := startSQLSpan(ctx, conn.connDetails, query, conn.sensor)
	defer sp.Finish()

	res, err := conn.originalConn.QueryContext(ctx, query, args)
	if err != nil && err != driver.ErrSkip {
		sp.LogFields(otlog.Error(err))
	}

	return res, err
}
