// (c) Copyright IBM Corp. 2023

package instana

import (
	"context"
	"database/sql/driver"

	otlog "github.com/opentracing/opentracing-go/log"
)

type wQueryer struct {
	driver.Queryer
	connDetails DbConnDetails
	sensor      TracerLogger
}

func (conn *wQueryer) Query(query string, args []driver.Value) (driver.Rows, error) {
	ctx := context.Background()

	sp, errKey := startSQLSpan(ctx, conn.connDetails, query, conn.sensor)
	defer sp.Finish()

	res, err := conn.Queryer.Query(query, args)

	if err != nil && err != driver.ErrSkip {
		sp.LogFields(otlog.Error(err))
		sp.SetTag(errKey, err.Error())
	}

	return res, err

}
