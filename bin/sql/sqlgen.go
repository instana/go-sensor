// (c) Copyright IBM Corp. 2022

package main

import (
	"fmt"
	"github.com/mxschmitt/golang-combinations"
	"sort"
	"strings"
)

var funcs string

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

	for _, subFconn := range fconn {
		s := strings.Join(subFconn, " && ")
		s = strings.ReplaceAll(s, "driver.", "is")

		fmt.Printf("if %s {\n", s)

		name := strings.ReplaceAll(strings.Join(subFconn, "_"), "driver.", "")

		funcs += returnSt(prefix, t, name, subFconn)

		args := ""
		for _, tt := range subFconn {
			args += "," + strings.ReplaceAll(tt, "driver.", "")
		}

		//todo: change
		if prefix == "w_stmt_" {
			fmt.Println("return " + "get_stmt_" + name + "(stmt , query , connDetails , sensor  " + args + ")")
		} else {
			fmt.Println("return " + "get_conn_" + name + "(connDetails, conn  , sensor  " + args + ")")
		}

		fmt.Println("}")
	}

	if t == "driver.Stmt" {
		fmt.Println("return stmt")
	} else {
		fmt.Println("return conn")
	}

	fmt.Println("}")

}

func returnSt(prefix string, t string, name string, v []string) string {
	res := ""

	args := ""
	for _, tt := range v {
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
