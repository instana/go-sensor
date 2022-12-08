// (c) Copyright IBM Corp. 2022

package instana

import (
	"context"
	"database/sql/driver"
	otlog "github.com/opentracing/opentracing-go/log"
)

type wExecer struct {
	originalConn driver.Execer
	driver.Conn
	connDetails dbConnDetails
	sensor      *Sensor
}

func (conn *wExecer) Exec(query string, args []driver.Value) (driver.Result, error) {
	ctx := context.Background()

	sp := startSQLSpan(ctx, conn.connDetails, query, conn.sensor)
	defer sp.Finish()

	res, err := conn.originalConn.Exec(query, args)
	if err != nil && err != driver.ErrSkip {
		sp.LogFields(otlog.Error(err))
	}

	return res, err
}
