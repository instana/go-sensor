// (c) Copyright IBM Corp. 2022

package main

import (
	"fmt"
	"github.com/mxschmitt/golang-combinations"
	"sort"
	"strings"
)

var funcs string

var _conn_m = map[string]string{}
var _stmt_m = map[string]string{}

var arrayConn = []string{
	"driver.Conn",
	"driver.Execer",
	"driver.ExecerContext",
	"driver.Queryer",
	"driver.QueryerContext",
	"driver.ConnPrepareContext",
	"driver.NamedValueChecker",
	"driver.ColumnConverter",
}

var arrayStmt = []string{
	"driver.Stmt",
	"driver.StmtExecContext",
	"driver.StmtQueryContext",
	"driver.NamedValueChecker",
	"driver.ColumnConverter",
}

func inArray(el string, arr []string) bool {
	for _, v := range arr {
		if v == el {
			return true
		}
	}

	return false
}

func removeOnceFromArr(el string, arr []string) []string {
	for k, v := range arr {
		if v == el {
			return append(arr[:k], arr[k+1:]...)
		}
	}
	return arr
}

func main() {

	fmt.Println("// (c) Copyright IBM Corp. 2022")
	fmt.Println("package instana")
	fmt.Println("import \"database/sql/driver\"")

	gen("driver.Conn", "w_conn_", arrayConn)
	gen("driver.Stmt", "w_stmt_", arrayStmt)

	fmt.Println("var _conn_n = map[int]func(connDetails dbConnDetails, conn driver.Conn, sensor *Sensor, Execer driver.Execer, ExecerContext driver.ExecerContext, Queryer driver.Queryer, QueryerContext driver.QueryerContext, ConnPrepareContext driver.ConnPrepareContext, NamedValueChecker driver.NamedValueChecker, ColumnConverter driver.ColumnConverter) driver.Conn {")
	//fmt.Println(_conn_m)
	for k, v := range _conn_m {
		fmt.Println(k, ":", v, ",")
	}
	fmt.Println("}")
	//fmt.Println(_stmt_m)

	fmt.Println("var _stmt_n = map[int]func( driver.Stmt,  string,  dbConnDetails,  *Sensor,  driver.StmtExecContext,  driver.StmtQueryContext,  driver.NamedValueChecker,  driver.ColumnConverter) driver.Stmt {")
	//fmt.Println(_conn_m)
	for k, v := range _stmt_m {
		fmt.Println(k, ":", v, ",")
	}
	fmt.Println("}")

	/////

	fmt.Println(`
func _btu(args ...bool) int {
	res := 0x1
	res = res << 1

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
`)
}

func gen(droverType, prefix string, arr []string) {
	funcs = ""
	conn := combinations.All(arr)
	fconn := [][]string{}
	for _, v := range conn {
		if inArray(droverType, v) {
			//remove basic type
			v = v[1:]
			if len(v) == 0 {
				continue
			}
			fconn = append(fconn, v)
		}

	}

	sort.Slice(fconn, func(i, j int) bool {
		return len(fconn[i]) > len(fconn[j])
	})

	t := []string{}
	for _, v := range fconn {
		fmt.Println("//", v)
		t = append(t, genType(prefix, v, droverType))
	}

	genIsAlreadyWrapped(droverType, t)

	genWrapper(prefix, droverType, arr, fconn)

	fmt.Println(funcs)

}

func genType(prefix string, arr []string, droverType string) string {
	name := strings.ReplaceAll(strings.Join(arr, "_"), "driver.", "")
	fmt.Println("type " + prefix + name + " struct {")
	fmt.Println(droverType)
	for _, v := range arr {
		fmt.Println(v)

	}
	fmt.Println("}")

	return prefix + name
}

func genIsAlreadyWrapped(t string, types []string) {
	noPrefix := strings.ReplaceAll(t, "driver.", "")
	fmt.Printf("func %sAlreadyWrapped(%s %s) bool {\n", strings.ToLower(noPrefix), strings.ToLower(noPrefix), t)
	fmt.Printf("switch %s.(type) {\n", strings.ToLower(noPrefix))

	cTypes := ""
	for _, v := range types {
		cTypes += "*" + v + ","
	}
	cTypes = strings.TrimRight(cTypes, ",") + ":"
	fmt.Printf("case %s\n", cTypes)
	fmt.Println("return true")
	fmt.Println("}")
	fmt.Println("return false")
	fmt.Println("}")
}

func genWrapper(prefix, t string, originalTypes []string, fconn [][]string) {

	if t == "driver.Stmt" {
		fmt.Println("func wrapStmt(stmt driver.Stmt, query string, connDetails dbConnDetails, sensor *Sensor) driver.Stmt {")
	} else {
		fmt.Println("func wrapConn(connDetails dbConnDetails, conn driver.Conn, sensor *Sensor) driver.Conn {")
	}

	mTypes := removeOnceFromArr(t, originalTypes)

	for _, t := range mTypes {
		n := strings.ReplaceAll(t, "driver.", "")

		//todo: change
		if prefix == "w_stmt_" {
			fmt.Printf("%s, is%s := stmt.(%s)\n", n, n, t)
		} else {
			fmt.Printf("%s, is%s := conn.(%s)\n", n, n, t)
		}

	}

	args := ""
	for _, tt := range mTypes {
		args += "," + strings.ReplaceAll(tt, "driver.", "")
	}

	for _, subFconn := range fconn {
		//s := strings.Join(subFconn, " && ")
		//s = strings.ReplaceAll(s, "driver.", "is")

		//fmt.Printf("if %s {\n", s)

		name := strings.ReplaceAll(strings.Join(subFconn, "_"), "driver.", "")

		funcs += returnSt(prefix, t, name, subFconn, mTypes)

		//todo: change
		if prefix == "w_stmt_" {
			//n := "get_stmt_" + name + "(stmt , query , connDetails , sensor  " + args + ")"
			genConnMap("w_stmt_", mTypes, subFconn, "get_stmt_"+name)

			//fmt.Println("return " + n)
		} else {
			//n := "get_conn_" + name + "(connDetails, conn  , sensor  " + args + ")"
			genConnMap("w_conn_", mTypes, subFconn, "get_conn_"+name)
			//fmt.Println("return " + n)
		}

		//fmt.Println("}")
	}

	s := strings.Join(mTypes, " && ")
	s = strings.ReplaceAll(s, "driver.", "is")

	if t == "driver.Stmt" {
		fmt.Printf(`if f, ok := _stmt_n[_btu(%s)]; ok {
				return f(stmt , query , connDetails , sensor  %s )
	}
	return &wStmt{
Stmt:stmt,
connDetails: connDetails,
query: query,
sensor: sensor,
}
	`, strings.ReplaceAll(s, "&&", ","), args)

	} else {
		fmt.Printf(`if f, ok := _conn_n[_btu(%s)]; ok {
				return f(connDetails , conn , sensor  %s )
	}
	return &wConn{
Conn:conn,
connDetails: connDetails,
sensor: sensor,
}
	`, strings.ReplaceAll(s, "&&", ","), args)
	}

	fmt.Println("}")

}

func returnSt(prefix string, t string, name string, v []string, arr []string) string {
	res := ""

	args := ""
	for _, tt := range arr {
		args += "," + strings.ReplaceAll(tt, "driver.", "") + " " + tt
	}

	if t == "driver.Stmt" {
		res += "func get_stmt_" + name + "(stmt driver.Stmt, query string, connDetails dbConnDetails, sensor *Sensor " + args + ") driver.Stmt {\n"
		res += fmt.Sprintf("return &%s {\n", prefix+name)
		res += fmt.Sprintf("Stmt: &wStmt{\nStmt:stmt,\nconnDetails: connDetails,\nquery: query,\nsensor: sensor,\n},")
	} else {
		res += "func get_conn_" + name + "(connDetails dbConnDetails, conn driver.Conn, sensor *Sensor" + args + ") driver.Conn {\n"
		res += fmt.Sprintf("return &%s {\n", prefix+name)
		res += fmt.Sprintf("Conn: &wConn{\nConn:conn,\nconnDetails: connDetails,\nsensor: sensor,\n},")
	}

	for _, tt := range v {
		n := strings.ReplaceAll(tt, "driver.", "")

		if n != "ColumnConverter" && n != "NamedValueChecker" {
			res += fmt.Sprintf("%s: &w%s{\n", n, n)
			res += fmt.Sprintf("%s: %s,\n", n, n)
			res += fmt.Sprintf("connDetails: connDetails,\n")
			res += fmt.Sprintf("sensor: sensor,\n")
			if t == "driver.Stmt" {
				res += fmt.Sprintf("query: query,\n")
			}
			res += fmt.Sprintf("},")
		} else {
			res += fmt.Sprintf("%s: %s,\n", n, n)
		}
	}

	res += fmt.Sprintf("}\n")
	res += fmt.Sprintf("}\n")
	return res
}

func genConnMap(prefix string, arr, fconn []string, funcName string) {

	mask := []bool{}

	for _, v := range arr {
		if inArray(v, fconn) {
			mask = append(mask, true)
		} else {
			mask = append(mask, false)
		}
	}

	if prefix == "w_stmt_" {
		_stmt_m[_btu(mask...)] = funcName
	} else {
		_conn_m[_btu(mask...)] = funcName
	}

}

func _btu(args ...bool) string {
	res := 0x1
	res = res << 1

	for k, v := range args {
		if v {
			res = res | 0x1
		}

		if len(args)-1 != k {
			res = res << 1
		}
	}

	return fmt.Sprintf("0x%b", res)
}

//var _conn_m map[int]func(connDetails dbConnDetails, conn driver.Conn, sensor *Sensor, Execer driver.Execer, ExecerContext driver.ExecerContext, Queryer driver.Queryer, QueryerContext driver.QueryerContext, ConnPrepareContext driver.ConnPrepareContext, NamedValueChecker driver.NamedValueChecker, ColumnConverter driver.ColumnConverter) driver.Conn {
//1 : get_conn_Execer_ExecerContext_Queryer_QueryerContext_ConnPrepareContext_NamedValueChecker,
//}
