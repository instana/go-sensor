// (c) Copyright IBM Corp. 2022
// Code generated by sqlgen. DO NOT EDIT.

package instana

import "database/sql/driver"

{{- /*List of types
Example:

// [driver.Execer driver.ExecerContext driver.Queryer driver.QueryerContext driver.ConnPrepareContext driver.NamedValueChecker]
type w_conn_Execer_ExecerContext_Queryer_QueryerContext_ConnPrepareContext_NamedValueChecker struct {
        driver.Conn
        driver.Execer
        driver.ExecerContext
        driver.Queryer
        driver.QueryerContext
        driver.ConnPrepareContext
        driver.NamedValueChecker
}
*/}}
{{range .Drivers}}
// {{.Interfaces}}
type w_{{.TypeName}} struct {
	{{.BasicType}}
	{{- range .Interfaces}}
	{{if eq . "driver.ColumnConverter"}}cc {{end}}{{.}}
	{{- end}}
}

{{/* Cases with ColumnConverter are special and must be added.
	Go gets confused with an embedded type whose name and method name are the same.
Example:

func (w *w_stmt_StmtExecContext_StmtQueryContext_NamedValueChecker_ColumnConverter) ColumnConverter(idx int) driver.ValueConverter {
	return w.cc.ColumnConverter(idx)
}
*/}}
{{- if .HasColumnConverter}}
func (w *w_{{.TypeName}}) ColumnConverter(idx int) driver.ValueConverter {
	return w.cc.ColumnConverter(idx)
}
{{end}}
{{end}}

{{/* function connAlreadyWrapped
Example:

func connAlreadyWrapped(conn driver.Conn) bool {
	switch conn.(type) {
	case *wConn, *w_conn_Execer_ExecerContext_Queryer_QueryerContext_ConnPrepareContext_NamedValueChecker, ...:
		return true
	case w_conn_Execer_ExecerContext_Queryer_QueryerContext_ConnPrepareContext_NamedValueChecker, w_conn_Execer_ExecerContext_Queryer_QueryerContext_ConnPrepareContext, ...:
		return true
	}
	return false
}
*/}}
func connAlreadyWrapped(conn driver.Conn) bool {
	switch conn.(type) {
	case *wConn{{range .Drivers -}}
	  {{- if .IsConn -}}, *w_{{.TypeName}}{{end}}
	{{- end}}:
		return true
	{{- $firstDriver := index .Drivers 0}}
	case w_{{$firstDriver.TypeName}}{{range slice .Drivers 1 -}}
	  {{- if .IsConn -}}, w_{{.TypeName}}{{end}}
	{{- end}}:
		return true
	}
	return false
}

func wrapConn(connDetails dbConnDetails, conn driver.Conn, sensor *Sensor) driver.Conn {
	Execer, isExecer := conn.(driver.Execer)
	ExecerContext, isExecerContext := conn.(driver.ExecerContext)
	Queryer, isQueryer := conn.(driver.Queryer)
	QueryerContext, isQueryerContext := conn.(driver.QueryerContext)
	ConnPrepareContext, isConnPrepareContext := conn.(driver.ConnPrepareContext)
	NamedValueChecker, isNamedValueChecker := conn.(driver.NamedValueChecker)
	if f, ok := _conn_n[convertBooleansToInt(isExecer, isExecerContext, isQueryer, isQueryerContext, isConnPrepareContext, isNamedValueChecker)]; ok {
		return f(connDetails, conn, sensor, Execer, ExecerContext, Queryer, QueryerContext, ConnPrepareContext, NamedValueChecker)
	}
	return &wConn{
		Conn:        conn,
		connDetails: connDetails,
		sensor:      sensor,
	}
}

{{/* Getter functions for each driver combination.
Example:

func get_conn_Execer_ExecerContext_Queryer_QueryerContext_ConnPrepareContext_NamedValueChecker(connDetails dbConnDetails, conn driver.Conn, sensor *Sensor, Execer driver.Execer, ExecerContext driver.ExecerContext, Queryer driver.Queryer, QueryerContext driver.QueryerContext, ConnPrepareContext driver.ConnPrepareContext, NamedValueChecker driver.NamedValueChecker) driver.Conn {
	return &w_conn_Execer_ExecerContext_Queryer_QueryerContext_ConnPrepareContext_NamedValueChecker{
		Conn: &wConn{
			Conn:        conn,
			connDetails: connDetails,
			sensor:      sensor,
		}, Execer: &wExecer{
			Execer:      Execer,
			connDetails: connDetails,
			sensor:      sensor,
		}, ExecerContext: &wExecerContext{
			ExecerContext: ExecerContext,
			connDetails:   connDetails,
			sensor:        sensor,
		}, Queryer: &wQueryer{
			Queryer:     Queryer,
			connDetails: connDetails,
			sensor:      sensor,
		}, QueryerContext: &wQueryerContext{
			QueryerContext: QueryerContext,
			connDetails:    connDetails,
			sensor:         sensor,
		}, ConnPrepareContext: &wConnPrepareContext{
			ConnPrepareContext: ConnPrepareContext,
			connDetails:        connDetails,
			sensor:             sensor,
		}, NamedValueChecker: NamedValueChecker,
	}
}
*/}}

{{range .Drivers}}
func get_{{.TypeName}}(
	{{- if .IsConn -}}
		connDetails dbConnDetails, conn driver.Conn, sensor *Sensor{{range connInterfaces}}, {{replace . "driver." ""}} {{.}}{{end}}
	{{- else -}}
		stmt driver.Stmt, query string, connDetails dbConnDetails, sensor *Sensor{{range stmtInterfaces}}, {{replace . "driver." ""}} {{.}}{{end}}
	{{- end}}) {{if .IsConn}}driver.Conn{{else}}driver.Stmt{{end}} {
	return &w_{{.TypeName}}{
		{{- /*TODO: refactor this piece*/ -}}
		{{if .IsConn}}
		Conn: &wConn{
			Conn:	conn,
			connDetails:	connDetails,
			sensor:	sensor,
		},
		{{- range .Interfaces -}}
			{{$theType := replace . "driver." ""}}
			{{- if eq $theType "NamedValueChecker"}} NamedValueChecker: NamedValueChecker,
			{{- else if eq $theType "ColumnConverter"}}
			cc: ColumnConverter,
			{{- else}} {{$theType}}:	&w{{$theType -}}{
			{{$theType}}:	{{$theType}},
			connDetails:	connDetails,
			sensor:	sensor,
		},
			{{- end -}}
		{{- end -}}
		{{else}}
		Stmt: &wStmt{
			Stmt:	stmt,
			connDetails:	connDetails,
			query:	query,
			sensor:	sensor,
		},
		{{- range .Interfaces -}}
		  {{$theType := replace . "driver." ""}}
		  {{- if eq $theType "NamedValueChecker"}} NamedValueChecker: NamedValueChecker,
		  {{- else if eq $theType "ColumnConverter"}} cc: ColumnConverter,
		  {{- else}}
		  {{$theType}}:	&w{{$theType -}}{
		  {{$theType}}:	{{$theType}},
		  connDetails:	connDetails,
		  sensor:	sensor,
		  query:	query,
		},
			{{- end -}}
		{{- end -}}
		{{end}}
	}
}
{{end}}

func stmtAlreadyWrapped(stmt driver.Stmt) bool {
	switch stmt.(type) {
	case *wStmt, *w_stmt_StmtExecContext_StmtQueryContext_NamedValueChecker_ColumnConverter, *w_stmt_StmtExecContext_StmtQueryContext_NamedValueChecker, *w_stmt_StmtQueryContext_NamedValueChecker_ColumnConverter, *w_stmt_StmtExecContext_NamedValueChecker_ColumnConverter, *w_stmt_StmtExecContext_StmtQueryContext_ColumnConverter, *w_stmt_StmtQueryContext_ColumnConverter, *w_stmt_StmtQueryContext_NamedValueChecker, *w_stmt_StmtExecContext_ColumnConverter, *w_stmt_StmtExecContext_NamedValueChecker, *w_stmt_NamedValueChecker_ColumnConverter, *w_stmt_StmtExecContext_StmtQueryContext, *w_stmt_ColumnConverter, *w_stmt_StmtExecContext, *w_stmt_NamedValueChecker, *w_stmt_StmtQueryContext:
		return true
	case w_stmt_StmtExecContext_StmtQueryContext_NamedValueChecker_ColumnConverter, w_stmt_StmtExecContext_StmtQueryContext_NamedValueChecker, w_stmt_StmtQueryContext_NamedValueChecker_ColumnConverter, w_stmt_StmtExecContext_NamedValueChecker_ColumnConverter, w_stmt_StmtExecContext_StmtQueryContext_ColumnConverter, w_stmt_StmtQueryContext_ColumnConverter, w_stmt_StmtQueryContext_NamedValueChecker, w_stmt_StmtExecContext_ColumnConverter, w_stmt_StmtExecContext_NamedValueChecker, w_stmt_NamedValueChecker_ColumnConverter, w_stmt_StmtExecContext_StmtQueryContext, w_stmt_ColumnConverter, w_stmt_StmtExecContext, w_stmt_NamedValueChecker, w_stmt_StmtQueryContext:
		return true
	}
	return false
}

func wrapStmt(stmt driver.Stmt, query string, connDetails dbConnDetails, sensor *Sensor) driver.Stmt {
	StmtExecContext, isStmtExecContext := stmt.(driver.StmtExecContext)
	StmtQueryContext, isStmtQueryContext := stmt.(driver.StmtQueryContext)
	NamedValueChecker, isNamedValueChecker := stmt.(driver.NamedValueChecker)
	ColumnConverter, isColumnConverter := stmt.(driver.ColumnConverter)
	if f, ok := _stmt_n[convertBooleansToInt(isStmtExecContext, isStmtQueryContext, isNamedValueChecker, isColumnConverter)]; ok {
		return f(stmt, query, connDetails, sensor, StmtExecContext, StmtQueryContext, NamedValueChecker, ColumnConverter)
	}
	return &wStmt{
		Stmt:        stmt,
		connDetails: connDetails,
		query:       query,
		sensor:      sensor,
	}
}

var _conn_n = map[int]func(dbConnDetails, driver.Conn, *Sensor, driver.Execer, driver.ExecerContext, driver.Queryer, driver.QueryerContext, driver.ConnPrepareContext, driver.NamedValueChecker) driver.Conn{
	0b111101: get_conn_Execer_ExecerContext_Queryer_QueryerContext_NamedValueChecker,
	0b101011: get_conn_Execer_Queryer_ConnPrepareContext_NamedValueChecker,
	0b11100:  get_conn_ExecerContext_Queryer_QueryerContext,
	0b1000:   get_conn_Queryer,
	0b111010: get_conn_Execer_ExecerContext_Queryer_ConnPrepareContext,
	0b110101: get_conn_Execer_ExecerContext_QueryerContext_NamedValueChecker,
	0b110110: get_conn_Execer_ExecerContext_QueryerContext_ConnPrepareContext,
	0b1001:   get_conn_Queryer_NamedValueChecker,
	0b110:    get_conn_QueryerContext_ConnPrepareContext,
	0b110010: get_conn_Execer_ExecerContext_ConnPrepareContext,
	0b100100: get_conn_Execer_QueryerContext,
	0b111111: get_conn_Execer_ExecerContext_Queryer_QueryerContext_ConnPrepareContext_NamedValueChecker,
	0b110111: get_conn_Execer_ExecerContext_QueryerContext_ConnPrepareContext_NamedValueChecker,
	0b10100:  get_conn_ExecerContext_QueryerContext,
	0b100010: get_conn_Execer_ConnPrepareContext,
	0b100:    get_conn_QueryerContext,
	0b11000:  get_conn_ExecerContext_Queryer,
	0b10000:  get_conn_ExecerContext,
	0b111001: get_conn_Execer_ExecerContext_Queryer_NamedValueChecker,
	0b10110:  get_conn_ExecerContext_QueryerContext_ConnPrepareContext,
	0b111000: get_conn_Execer_ExecerContext_Queryer,
	0b1010:   get_conn_Queryer_ConnPrepareContext,
	0b101000: get_conn_Execer_Queryer,
	0b1:      get_conn_NamedValueChecker,
	0b11111:  get_conn_ExecerContext_Queryer_QueryerContext_ConnPrepareContext_NamedValueChecker,
	0b10111:  get_conn_ExecerContext_QueryerContext_ConnPrepareContext_NamedValueChecker,
	0b100011: get_conn_Execer_ConnPrepareContext_NamedValueChecker,
	0b10011:  get_conn_ExecerContext_ConnPrepareContext_NamedValueChecker,
	0b101001: get_conn_Execer_Queryer_NamedValueChecker,
	0b1101:   get_conn_Queryer_QueryerContext_NamedValueChecker,
	0b1111:   get_conn_Queryer_QueryerContext_ConnPrepareContext_NamedValueChecker,
	0b1110:   get_conn_Queryer_QueryerContext_ConnPrepareContext,
	0b110100: get_conn_Execer_ExecerContext_QueryerContext,
	0b10010:  get_conn_ExecerContext_ConnPrepareContext,
	0b10001:  get_conn_ExecerContext_NamedValueChecker,
	0b100001: get_conn_Execer_NamedValueChecker,
	0b101111: get_conn_Execer_Queryer_QueryerContext_ConnPrepareContext_NamedValueChecker,
	0b100111: get_conn_Execer_QueryerContext_ConnPrepareContext_NamedValueChecker,
	0b111100: get_conn_Execer_ExecerContext_Queryer_QueryerContext,
	0b110011: get_conn_Execer_ExecerContext_ConnPrepareContext_NamedValueChecker,
	0b111:    get_conn_QueryerContext_ConnPrepareContext_NamedValueChecker,
	0b110000: get_conn_Execer_ExecerContext,
	0b111011: get_conn_Execer_ExecerContext_Queryer_ConnPrepareContext_NamedValueChecker,
	0b101100: get_conn_Execer_Queryer_QueryerContext,
	0b11101:  get_conn_ExecerContext_Queryer_QueryerContext_NamedValueChecker,
	0b100110: get_conn_Execer_QueryerContext_ConnPrepareContext,
	0b11:     get_conn_ConnPrepareContext_NamedValueChecker,
	0b101110: get_conn_Execer_Queryer_QueryerContext_ConnPrepareContext,
	0b110001: get_conn_Execer_ExecerContext_NamedValueChecker,
	0b101:    get_conn_QueryerContext_NamedValueChecker,
	0b10:     get_conn_ConnPrepareContext,
	0b11010:  get_conn_ExecerContext_Queryer_ConnPrepareContext,
	0b10101:  get_conn_ExecerContext_QueryerContext_NamedValueChecker,
	0b100000: get_conn_Execer,
	0b11011:  get_conn_ExecerContext_Queryer_ConnPrepareContext_NamedValueChecker,
	0b11110:  get_conn_ExecerContext_Queryer_QueryerContext_ConnPrepareContext,
	0b101010: get_conn_Execer_Queryer_ConnPrepareContext,
	0b11001:  get_conn_ExecerContext_Queryer_NamedValueChecker,
	0b100101: get_conn_Execer_QueryerContext_NamedValueChecker,
	0b1100:   get_conn_Queryer_QueryerContext,
	0b111110: get_conn_Execer_ExecerContext_Queryer_QueryerContext_ConnPrepareContext,
	0b101101: get_conn_Execer_Queryer_QueryerContext_NamedValueChecker,
	0b1011:   get_conn_Queryer_ConnPrepareContext_NamedValueChecker,
}
var _stmt_n = map[int]func(driver.Stmt, string, dbConnDetails, *Sensor, driver.StmtExecContext, driver.StmtQueryContext, driver.NamedValueChecker, driver.ColumnConverter) driver.Stmt{
	0b1110: get_stmt_StmtExecContext_StmtQueryContext_NamedValueChecker,
	0b1000: get_stmt_StmtExecContext,
	0b1101: get_stmt_StmtExecContext_StmtQueryContext_ColumnConverter,
	0b11:   get_stmt_NamedValueChecker_ColumnConverter,
	0b10:   get_stmt_NamedValueChecker,
	0b100:  get_stmt_StmtQueryContext,
	0b111:  get_stmt_StmtQueryContext_NamedValueChecker_ColumnConverter,
	0b1011: get_stmt_StmtExecContext_NamedValueChecker_ColumnConverter,
	0b101:  get_stmt_StmtQueryContext_ColumnConverter,
	0b1:    get_stmt_ColumnConverter,
	0b1100: get_stmt_StmtExecContext_StmtQueryContext,
	0b1111: get_stmt_StmtExecContext_StmtQueryContext_NamedValueChecker_ColumnConverter,
	0b110:  get_stmt_StmtQueryContext_NamedValueChecker,
	0b1001: get_stmt_StmtExecContext_ColumnConverter,
	0b1010: get_stmt_StmtExecContext_NamedValueChecker,
}

func convertBooleansToInt(args ...bool) int {
	res := 0

	for k, v := range args {
		if v {
			res = res | 0x1
		}

		if len(args)-1 != k {
			res = res << 1
		}
	}

	return res
}
