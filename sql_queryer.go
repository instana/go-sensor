// (c) Copyright IBM Corp. 2023

package instana

import (
	"context"
	"database/sql/driver"

	otlog "github.com/opentracing/opentracing-go/log"
)

type wQueryer struct {
	driver.Queryer
	connDetails dbConnDetails
	sensor      *Sensor
}

func (conn *wQueryer) Query(query string, args []driver.Value) (driver.Rows, error) {
	ctx := context.Background()
	sp := startSQLSpan(ctx, conn.connDetails, query, conn.sensor)
	defer sp.Finish()

	res, err := conn.Queryer.Query(query, args)
	if err != nil && err != driver.ErrSkip {
		sp.LogFields(otlog.Error(err))
	}

	return res, err

}
