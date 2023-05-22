// (c) Copyright IBM Corp. 2023
// Code generated by sqlgen. DO NOT EDIT.

package instana

import "database/sql/driver"

// Types
{{range .Drivers}}
// {{.Interfaces}}
type w_{{.TypeName}} struct {
	{{.BasicType}}
	{{- range .Interfaces}}
	{{if eq . "driver.ColumnConverter"}}cc {{end}}{{.}}
	{{- end}}
}

{{if .HasColumnConverter}}
func (w *w_{{.TypeName}}) ColumnConverter(idx int) driver.ValueConverter {
	return w.cc.ColumnConverter(idx)
}
{{- end -}}
{{- end -}}

// connAlreadyWrapped returns true if conn is already instrumented
func connAlreadyWrapped(conn driver.Conn) bool {
	{{$connTypes := driverTypes .Drivers true -}}
	switch conn.(type) {
	case *wConn{{range $connTypes -}}, *w_{{.}}{{- end}}:
		return true
	case w_{{index $connTypes 0}}{{range slice $connTypes 1 -}}, w_{{.}}{{- end}}:
		return true
	}
	return false
}

// wrapConn wraps the matching type around the driver.Conn based on which interfaces the driver implements
func wrapConn(connDetails DbConnDetails, conn driver.Conn, sensor TracerLogger) driver.Conn {
	{{range connInterfaces -}}
	{{replace . "driver." ""}}, is{{replace . "driver." ""}} := conn.({{.}})
	{{end -}}

	{{- $interfaceList := join connInterfaces ", "}}
	{{- $isList := replace $interfaceList "driver." "is"}}
	{{- $noPkgList := replace $interfaceList "driver." ""}}
	if f, ok := _conn_n[convertBooleansToInt({{$isList}})]; ok {
		return f(connDetails, conn, sensor, {{$noPkgList}})
  }

  return &wConn{
    Conn:conn,
    connDetails: connDetails,
    sensor: sensor,
  }
}

// driver.Conn Constructors
{{range .Drivers}}
{{if .IsConn}}
func get_{{.TypeName}}(connDetails DbConnDetails, conn driver.Conn, sensor TracerLogger{{range connInterfaces}}, {{replace . "driver." ""}} {{.}}{{end}}) driver.Conn {
	return &w_{{.TypeName}} {
		Conn: &wConn{
			Conn:	conn,
			connDetails:	connDetails,
			sensor:	sensor,
		},
		{{- range .Interfaces -}}
			{{$theType := replace . "driver." ""}}
			{{- if eq $theType "NamedValueChecker"}} NamedValueChecker: NamedValueChecker,
			{{- else if eq $theType "ColumnConverter"}} cc: ColumnConverter,
			{{- else}}
			{{$theType}}: &w{{$theType -}}{
			{{$theType}}:	{{$theType}},
			connDetails:	connDetails,
			sensor:	sensor,
		},
			{{- end -}}
		{{- end -}}
	}
}
{{- end -}} {{- /* if .IsConn */ -}}
{{- end -}} {{- /* range .Drivers*/ -}}

// driver.Stmt Constructors
{{range .Drivers}}
{{if eq .IsConn false}}
func get_{{.TypeName}}(stmt driver.Stmt, query string, connDetails DbConnDetails, sensor TracerLogger{{range stmtInterfaces}}, {{replace . "driver." ""}} {{.}}{{end}}) driver.Stmt {
	return &w_{{.TypeName}} {
		Stmt: &wStmt{
			Stmt:	stmt,
			connDetails:	connDetails,
			query:	query,
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
			query:	query,
		},
			{{- end -}}
		{{- end -}}
	}
}
{{- end -}} {{- /* if eq .IsConn false */ -}}
{{- end -}} {{- /* range .Drivers*/}}

// stmtAlreadyWrapped returns true if stmt is already instrumented
func stmtAlreadyWrapped(stmt driver.Stmt) bool {
	{{$stmtTypes := driverTypes .Drivers false -}}
	switch stmt.(type) {
	case *wStmt{{range $stmtTypes -}}, *w_{{.}}{{- end}}:
		return true
	case w_{{index $stmtTypes 0}}{{range slice $stmtTypes 1 -}}, w_{{.}}{{- end}}:
		return true
	}
	return false
}

// wrapStmt wraps the matching type around the driver.Stmt based on which interfaces the driver implements
func wrapStmt(stmt driver.Stmt, query string, connDetails DbConnDetails, sensor TracerLogger) driver.Stmt {
	{{range stmtInterfaces -}}
	{{replace . "driver." ""}}, is{{replace . "driver." ""}} := stmt.({{.}})
	{{end -}}

	{{- $interfaceList := join stmtInterfaces ", "}}
	{{- $isList := replace $interfaceList "driver." "is"}}
	{{- $noPkgList := replace $interfaceList "driver." ""}}
	if f, ok := _stmt_n[convertBooleansToInt({{$isList}})]; ok {
		return f(stmt, query, connDetails, sensor, {{$noPkgList}})
  }

	return &wStmt{
		Stmt:        stmt,
		connDetails: connDetails,
		query:       query,
		sensor:      sensor,
	}
}

// A map of all possible driver.Conn types. The key represents which interfaces are "turned on". eg: 0b1001.
//
// In the example above, the following constructor is returned: get_conn_Queryer_NamedValueChecker
//
// Each bit sequentially represents the interfaces: Execer, ExecerContext, Queryer, QueryerContext, ConnPrepareContext, NamedValueChecker
var _conn_n = map[int]func(DbConnDetails, driver.Conn, TracerLogger, {{join connInterfaces ", "}}) driver.Conn {
	{{range $k, $v := connMap -}}
	{{$k}}: {{$v}},
	{{end}}
}

// A map of all possible driver.Stmt types. The key represents which interfaces are "turned on". eg: 0b1001.
//
// In the example above, the following constructor is returned: get_stmt_StmtExecContext_ColumnConverter
//
// Each bit sequentially represents the interfaces: StmtExecContext, StmtQueryContext, NamedValueChecker, ColumnConverter
var _stmt_n = map[int]func(driver.Stmt, string, DbConnDetails, TracerLogger, {{join stmtInterfaces ", "}}) driver.Stmt {
	{{range $k, $v := stmtMap -}}
	{{$k}}: {{$v}},
	{{end}}
}

// convertBooleansToInt converts a slice of bools to a binary representation.
//
// Example:
//	convertBooleansToInt(true, false, true, true) = 0b1011
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
