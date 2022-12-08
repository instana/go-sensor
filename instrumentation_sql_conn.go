// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instana

import (
	"database/sql/driver"
)

type wrappedSQLConn struct {
	driver.Conn

	details dbConnDetails
	sensor  *Sensor
}

func (conn *wrappedSQLConn) Prepare(query string) (driver.Stmt, error) {
	stmt, err := conn.Conn.Prepare(query)
	if err != nil {
		return stmt, err
	}

	if stmtAlreadyWrapped(stmt) {
		return stmt, nil
	}

	w := wrapStmt(stmt, query, conn.details, conn.sensor)

	return w, nil
}

func stmtAlreadyWrapped(stmt driver.Stmt) bool {
	switch stmt.(type) {
	case wStmtQueryContext, wStmtExecContext:
		return true
	case *wrappedSQLStmt, *wStmtQueryContext, *wStmtExecContext:
		return true
	}

	return false
}

func wrapStmt(stmt driver.Stmt, query string, connDetails dbConnDetails, sensor *Sensor) driver.Stmt {
	var w driver.Stmt

	w = &wrappedSQLStmt{
		Stmt:        stmt,
		connDetails: connDetails,
		query:       query,
		sensor:      sensor,
	}

	//if s, ok := stmt.(driver.NamedValueChecker); ok {
	//	w = &wStmtNamedValueChecker{
	//		originalStmt: s,
	//		Stmt:         w,
	//	}
	//}

	//if s, ok := stmt.(driver.NamedValueChecker); ok {
	//	w
	//}

	if s, ok := stmt.(driver.StmtQueryContext); ok {
		if ss, ok := w.(driver.Stmt); ok {
			w = &wStmtQueryContext{
				StmtQueryContext: s,
				Stmt:             ss,
				connDetails:      connDetails,
				sensor:           sensor,
				query:            query,
			}
		}
	}

	if s, ok := stmt.(driver.StmtExecContext); ok {
		if ss, ok := w.(driver.Stmt); ok {
			w = &wStmtExecContext{
				StmtExecContext: s,
				Stmt:            ss,
				connDetails:     connDetails,
				sensor:          sensor,
				query:           query,
			}
		}

	}

	//if _, ok := stmt.(driver.NamedValueChecker); ok {
	//	panic(1)
	//}

	//if _, ok := (w).(driver.StmtExecContext); ok {
	//	panic(1)
	//}

	if _, ok := (w).(driver.StmtQueryContext); ok {
		panic(1)
	}

	if _, ok := (w).(driver.StmtQueryContext); ok {
		panic(1)
	}

	if _, ok := (w).(driver.NamedValueChecker); ok {
		panic(1)
	}
	return w
}

// TODO: REMOVE
type wStmtNamedValueChecker struct {
	originalStmt driver.NamedValueChecker
	driver.Stmt
}

func (stmt *wStmtNamedValueChecker) CheckNamedValue(d *driver.NamedValue) error {
	return stmt.CheckNamedValue(d)
}
