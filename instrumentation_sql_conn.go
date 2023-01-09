// (c) Copyright IBM Corp. 2023

package instana

import (
	"database/sql/driver"
)

type wConn struct {
	driver.Conn

	connDetails dbConnDetails
	sensor      *Sensor
}

func (conn *wConn) Prepare(query string) (driver.Stmt, error) {
	stmt, err := conn.Conn.Prepare(query)
	if err != nil {
		return stmt, err
	}

	if stmtAlreadyWrapped(stmt) {
		return stmt, nil
	}

	w := wrapStmt(stmt, query, conn.connDetails, conn.sensor)

	return w, nil
}
