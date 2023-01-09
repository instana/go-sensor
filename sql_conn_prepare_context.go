// (c) Copyright IBM Corp. 2023

package instana

import (
	"context"
	"database/sql/driver"
)

type wConnPrepareContext struct {
	driver.ConnPrepareContext
	connDetails dbConnDetails
	sensor      *Sensor
}

func (conn *wConnPrepareContext) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	stmt, err := conn.ConnPrepareContext.PrepareContext(ctx, query)

	if err != nil {
		return stmt, err
	}

	if stmtAlreadyWrapped(stmt) {
		return stmt, nil
	}

	return wrapStmt(stmt, query, conn.connDetails, conn.sensor), nil
}
