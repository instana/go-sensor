// (c) Copyright IBM Corp. 2022
// Code generated by sqlgen. DO NOT EDIT.

package instana

import "database/sql/driver"

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

func wrapConn(connDetails dbConnDetails, conn driver.Conn, sensor *Sensor) driver.Conn {
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

// Conn Constructors
{{range .Drivers}}
{{if .IsConn}}
func get_{{.TypeName}}(connDetails dbConnDetails, conn driver.Conn, sensor *Sensor{{range connInterfaces}}, {{replace . "driver." ""}} {{.}}{{end}}) driver.Conn {
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

// Stmt Constructors
{{range .Drivers}}
{{if eq .IsConn false}}
func get_{{.TypeName}}(stmt driver.Stmt, query string, connDetails dbConnDetails, sensor *Sensor{{range stmtInterfaces}}, {{replace . "driver." ""}} {{.}}{{end}}) driver.Stmt {
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

func wrapStmt(stmt driver.Stmt, query string, connDetails dbConnDetails, sensor *Sensor) driver.Stmt {
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

var _conn_n = map[int]func(dbConnDetails, driver.Conn, *Sensor, {{join connInterfaces ", "}}) driver.Conn {
	{{range $k, $v := connMap -}}
	{{$k}}: {{$v}},
	{{end}}
}

var _stmt_n = map[int]func(driver.Stmt, string, dbConnDetails, *Sensor, {{join stmtInterfaces ", "}}) driver.Stmt {
	{{range $k, $v := stmtMap -}}
	{{$k}}: {{$v}},
	{{end}}
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
