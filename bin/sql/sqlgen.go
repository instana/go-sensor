// (c) Copyright IBM Corp. 2022

package main

import (
	"fmt"
	"github.com/mxschmitt/golang-combinations"
	"sort"
	"strings"
)

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

	for _, v := range fconn {
		s := strings.Join(v, " && ")
		s = strings.ReplaceAll(s, "driver.", "is")

		fmt.Printf("if %s {\n", s)

		name := strings.ReplaceAll(strings.Join(v, "_"), "driver.", "")

		fmt.Printf("return &%s {\n", prefix+name)

		if t == "driver.Stmt" {
			fmt.Println("Stmt: &wStmt{\nStmt:stmt,\nconnDetails: connDetails,\nquery: query,\nsensor: sensor,\n},")
		} else {
			fmt.Println("Conn: &wConn{\nConn:conn,\nconnDetails: connDetails,\nsensor: sensor,\n},")
		}

		for _, tt := range v {
			n := strings.ReplaceAll(tt, "driver.", "")

			if n != "ColumnConverter" && n != "NamedValueChecker" {
				fmt.Printf("%s: &w%s{\n", n, n)
				fmt.Printf("%s: %s,\n", n, n)
				fmt.Println("connDetails: connDetails,")
				fmt.Println("sensor: sensor,")
				if t == "driver.Stmt" {
					fmt.Println("query: query,")
				}
				fmt.Println("},")
			} else {
				fmt.Printf("%s: %s,\n", n, n)
			}
		}

		fmt.Println("}")

		fmt.Println("}")
	}

	if t == "driver.Stmt" {
		fmt.Println("return stmt")
	} else {
		fmt.Println("return conn")
	}

	fmt.Println("}")

}
