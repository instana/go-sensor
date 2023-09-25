// (c) Copyright IBM Corp. 2023

package instana

import (
	"context"
	"database/sql/driver"

	otlog "github.com/opentracing/opentracing-go/log"
)

type wExecer struct {
	driver.Execer
	connDetails DbConnDetails
	sensor      TracerLogger
}

func (conn *wExecer) Exec(query string, args []driver.Value) (driver.Result, error) {
	ctx := context.Background()
	sp, errKey := startSQLSpan(ctx, conn.connDetails, query, conn.sensor)
	defer sp.Finish()

	res, err := conn.Execer.Exec(query, args)

	if err != nil && err != driver.ErrSkip {
		sp.LogFields(otlog.Error(err))
		sp.SetTag(errKey, err.Error())
	}

	return res, err
}
