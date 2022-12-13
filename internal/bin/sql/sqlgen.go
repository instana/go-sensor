// (c) Copyright IBM Corp. 2022

package main

import (
	"fmt"
	"sort"
	"strings"

	combinations "github.com/mxschmitt/golang-combinations"
)

const driverStmt = "driver.Stmt"

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
	gen(driverStmt, "w_stmt_", arrayStmt)

	printFunctionMaps()
	printUtil()
}

func gen(basicTypeForWrapper, prefix string, listOfTheOriginalTypes []string) {
	// generate all subsets of the interfaces
	conn := combinations.All(listOfTheOriginalTypes)
	var filteredSubsets [][]string

	for _, v := range conn {
		if inArray(basicTypeForWrapper, v) {
			//remove basic type
			v = v[1:]
			if len(v) == 0 {
				continue
			}
			filteredSubsets = append(filteredSubsets, v)
		}

	}

	sort.Slice(filteredSubsets, func(i, j int) bool {
		return len(filteredSubsets[i]) > len(filteredSubsets[j])
	})

	var typeNames []string
	for _, filteredSubset := range filteredSubsets {
		fmt.Println("//", filteredSubset) // just a comment
		// generate types and return type name
		typeNames = append(typeNames, genType(prefix, filteredSubset, basicTypeForWrapper))
	}

	genIsAlreadyWrapped(basicTypeForWrapper, typeNames)

	listOfTheInterfacesWithoutBasicType := removeOnceFromArr(basicTypeForWrapper, listOfTheOriginalTypes)

	genWrapper(basicTypeForWrapper, listOfTheInterfacesWithoutBasicType)

	generateAndPrintConstructors(basicTypeForWrapper, prefix, filteredSubsets, listOfTheInterfacesWithoutBasicType)

	generateConstructorMap(basicTypeForWrapper, filteredSubsets, listOfTheInterfacesWithoutBasicType)

}

func generateConstructorMap(basicTypeForWrapper string, filteredSubsets [][]string, listOfTheInterfacesWithoutBasicType []string) {
	for _, subset := range filteredSubsets {
		name := strings.ReplaceAll(strings.Join(subset, "_"), "driver.", "")

		if basicTypeForWrapper == driverStmt {
			genConnMap("w_stmt_", listOfTheInterfacesWithoutBasicType, subset, "get_stmt_"+name)
		} else {
			genConnMap("w_conn_", listOfTheInterfacesWithoutBasicType, subset, "get_conn_"+name)
		}
	}
}

func generateAndPrintConstructors(basicTypeForWrapper string, prefix string, filteredSubsets [][]string, listOfTheInterfacesWithoutBasicType []string) {
	var funcs string
	for _, subset := range filteredSubsets {
		name := strings.ReplaceAll(strings.Join(subset, "_"), "driver.", "")

		funcs += generateConstructors(prefix, basicTypeForWrapper, name, subset, listOfTheInterfacesWithoutBasicType)
	}

	fmt.Println(funcs)
}

func genType(prefix string, arr []string, basicTypeForWrapper string) string {
	name := strings.ReplaceAll(strings.Join(arr, "_"), "driver.", "")
	fmt.Println("type " + prefix + name + " struct {")
	fmt.Println(basicTypeForWrapper)
	for _, v := range arr {
		fmt.Println(v)

	}
	fmt.Println("}")

	return prefix + name
}

// generate function that checks if basic type is one of the existing wrappers
func genIsAlreadyWrapped(t string, types []string) {
	noPrefix := strings.ReplaceAll(t, "driver.", "")
	fmt.Printf("func %sAlreadyWrapped(%s %s) bool {\n", strings.ToLower(noPrefix), strings.ToLower(noPrefix), t)
	fmt.Printf("switch %s.(type) {\n", strings.ToLower(noPrefix))

	wrapperTypes := ""
	for _, v := range types {
		wrapperTypes += "*" + v + ","
	}
	wrapperTypes = strings.TrimRight(wrapperTypes, ",") + ":"
	fmt.Printf("case %s\n", wrapperTypes)
	fmt.Println("return true")
	fmt.Println("}")
	fmt.Println("return false")
	fmt.Println("}")
}

func genWrapper(basicTypeForWrapper string, listOfTheInterfacesWithoutBasicType []string) {

	// generate function definition
	if basicTypeForWrapper == driverStmt {
		fmt.Println("func wrapStmt(stmt driver.Stmt, query string, connDetails dbConnDetails, sensor *Sensor) driver.Stmt {")
	} else {
		fmt.Println("func wrapConn(connDetails dbConnDetails, conn driver.Conn, sensor *Sensor) driver.Conn {")
	}

	// we do not need to check the basic interface, because it is a part of the signature
	for _, t := range listOfTheInterfacesWithoutBasicType {
		n := strings.ReplaceAll(t, "driver.", "")
		// generate if statements to find what interfaces type implements
		if basicTypeForWrapper == driverStmt {
			fmt.Printf("%s, is%s := stmt.(%s)\n", n, n, t)
		} else {
			fmt.Printf("%s, is%s := conn.(%s)\n", n, n, t)
		}

	}

	// turn list of the types names, to the list of the function arguments for the constructor
	args := ""
	for _, tt := range listOfTheInterfacesWithoutBasicType {
		args += "," + strings.ReplaceAll(tt, "driver.", "")
	}

	// listOfBooleans to do lookup in the map
	listOfBooleans := strings.Join(listOfTheInterfacesWithoutBasicType, ",")
	listOfBooleans = strings.ReplaceAll(listOfBooleans, "driver.", "is")

	// generate wrapper body
	if basicTypeForWrapper == driverStmt {
		fmt.Printf(`if f, ok := _stmt_n[_btu(%s)]; ok {
				return f(stmt , query , connDetails , sensor  %s )
	}
	return &wStmt{
Stmt:stmt,
connDetails: connDetails,
query: query,
sensor: sensor,
}
	`, listOfBooleans, args)

	} else {
		fmt.Printf(`if f, ok := _conn_n[_btu(%s)]; ok {
				return f(connDetails , conn , sensor  %s )
	}
	return &wConn{
Conn:conn,
connDetails: connDetails,
sensor: sensor,
}
	`, listOfBooleans, args)
	}

	fmt.Println("}")

}

// all the constructors have same signature
func generateConstructors(prefix string, basicTypeForWrapper string, name string, subset []string, listOfTheInterfacesWithoutBasicType []string) string {
	res := ""

	args := ""
	for _, tt := range listOfTheInterfacesWithoutBasicType {
		args += "," + strings.ReplaceAll(tt, "driver.", "") + " " + tt
	}

	// generate function definition
	if basicTypeForWrapper == driverStmt {
		res += "func get_stmt_" + name + "(stmt driver.Stmt, query string, connDetails dbConnDetails, sensor *Sensor " + args + ") driver.Stmt {\n"
		res += fmt.Sprintf("return &%s {\n", prefix+name)
		res += fmt.Sprintf("Stmt: &wStmt{\nStmt:stmt,\nconnDetails: connDetails,\nquery: query,\nsensor: sensor,\n},")
	} else {
		res += "func get_conn_" + name + "(connDetails dbConnDetails, conn driver.Conn, sensor *Sensor" + args + ") driver.Conn {\n"
		res += fmt.Sprintf("return &%s {\n", prefix+name)
		res += fmt.Sprintf("Conn: &wConn{\nConn:conn,\nconnDetails: connDetails,\nsensor: sensor,\n},")
	}

	for _, tt := range subset {
		n := strings.ReplaceAll(tt, "driver.", "")

		// we do not use wrappers for ColumnConverter and NamedValueChecker
		if n != "ColumnConverter" && n != "NamedValueChecker" {
			// add wrapper for a specific type
			res += fmt.Sprintf("%s: &w%s{\n", n, n)
			res += fmt.Sprintf("%s: %s,\n", n, n)
			res += fmt.Sprintf("connDetails: connDetails,\n")
			res += fmt.Sprintf("sensor: sensor,\n")
			if basicTypeForWrapper == driverStmt {
				res += fmt.Sprintf("query: query,\n")
			}
			res += fmt.Sprintf("},")
		} else {
			// add types that we do not wrap (ColumnConverter and NamedValueChecker) currently
			res += fmt.Sprintf("%s: %s,\n", n, n)
		}
	}

	res += fmt.Sprintf("}\n")
	res += fmt.Sprintf("}\n")
	return res
}

// generate map to do a function lookup
func genConnMap(prefix string, listOfTheInterfacesWithoutBasicType, subset []string, funcName string) {
	var mask []bool

	for _, v := range listOfTheInterfacesWithoutBasicType {
		if inArray(v, subset) {
			mask = append(mask, true)
		} else {
			mask = append(mask, false)
		}
	}

	if prefix == "w_stmt_" {
		_stmt_m[booleansToBinaryRepresentation(mask...)] = funcName
	} else {
		_conn_m[booleansToBinaryRepresentation(mask...)] = funcName
	}

}

func booleansToBinaryRepresentation(args ...bool) string {
	res := 0

	for k, v := range args {
		if v {
			res = res | 0x1
		}

		if len(args)-1 != k {
			res = res << 1
		}
	}

	return fmt.Sprintf("0b%b", res)
}

func printFunctionMaps() {
	fmt.Println("var _conn_n = map[int]func(connDetails dbConnDetails, conn driver.Conn, sensor *Sensor, Execer driver.Execer, ExecerContext driver.ExecerContext, Queryer driver.Queryer, QueryerContext driver.QueryerContext, ConnPrepareContext driver.ConnPrepareContext, NamedValueChecker driver.NamedValueChecker, ColumnConverter driver.ColumnConverter) driver.Conn {")
	for k, v := range _conn_m {
		fmt.Println(k, ":", v, ",")
	}
	fmt.Println("}")

	fmt.Println("var _stmt_n = map[int]func( driver.Stmt,  string,  dbConnDetails,  *Sensor,  driver.StmtExecContext,  driver.StmtQueryContext,  driver.NamedValueChecker,  driver.ColumnConverter) driver.Stmt {")
	for k, v := range _stmt_m {
		fmt.Println(k, ":", v, ",")
	}
	fmt.Println("}")
}

func printUtil() {
	fmt.Println(`func _btu(args ...bool) int {
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
}`)
}
