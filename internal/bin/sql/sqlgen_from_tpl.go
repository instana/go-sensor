// (c) Copyright IBM Corp. 2022
//go:build go1.18

package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/template"

	combinations "github.com/mxschmitt/golang-combinations"
)

const driverStmt2 = "driver.Stmt"
const driverConn2 = "driver.Conn"

var _conn_m2 = map[string]string{}
var _stmt_m2 = map[string]string{}

var arrayConn2 = []string{
	"driver.Conn",
	"driver.Execer",
	"driver.ExecerContext",
	"driver.Queryer",
	"driver.QueryerContext",
	"driver.ConnPrepareContext",
	"driver.NamedValueChecker",
}

var arrayStmt2 = []string{
	"driver.Stmt",
	"driver.StmtExecContext",
	"driver.StmtQueryContext",
	"driver.NamedValueChecker",
	"driver.ColumnConverter",
}

type DriverCombo struct {
	TypeName   string
	BasicType  string
	Interfaces []string
	IsConn     bool
}

func (d DriverCombo) HasColumnConverter() bool {
	return inArray2("driver.ColumnConverter", d.Interfaces)
}

type WrapperData struct {
	Drivers []DriverCombo
}

func inArray2(el string, arr []string) bool {
	for _, v := range arr {
		if v == el {
			return true
		}
	}

	return false
}

func removeOnceFromArr2(el string, arr []string) []string {
	for k, v := range arr {
		if v == el {
			return append(arr[:k], arr[k+1:]...)
		}
	}
	return arr
}

func replace(s, old, new string) string {
	return strings.Replace(s, old, new, -1)
}

var funcMap template.FuncMap

var connInterfacesNoBasicType []string
var stmtInterfacesNoBasicType []string

func init() {
	funcMap = template.FuncMap{
		"replace": replace,
		"connInterfaces": func() []string {
			return connInterfacesNoBasicType
		},
		"stmtInterfaces": func() []string {
			return stmtInterfacesNoBasicType
		},
	}
}

func main1() {
	tpl, err := template.New("sql_wrappers.tpl").Funcs(funcMap).ParseFiles("sql_wrappers.tpl")

	if err != nil {
		panic(err)
	}

	wrapperData := WrapperData{}

	arrayConnCopy := make([]string, len(arrayConn2))
	arrayStmtCopy := make([]string, len(arrayStmt2))

	copy(arrayStmtCopy, arrayStmt2)
	copy(arrayConnCopy, arrayConn2)

	var connSubsets, stmtSubsets [][]string
	var connTypes, stmtTypes []string

	connSubsets, connTypes, connInterfacesNoBasicType = getTypeCombinations(driverConn2, "conn_", arrayConn2)
	stmtSubsets, stmtTypes, stmtInterfacesNoBasicType = getTypeCombinations(driverStmt2, "stmt_", arrayStmt2)

	var drivers []DriverCombo

	for i := 0; i < len(connTypes); i++ {
		d := DriverCombo{
			TypeName:   connTypes[i],
			Interfaces: connSubsets[i],
			BasicType:  driverConn2,
			IsConn:     true,
		}
		drivers = append(drivers, d)
	}

	for i := 0; i < len(stmtTypes); i++ {
		d := DriverCombo{
			TypeName:   stmtTypes[i],
			Interfaces: stmtSubsets[i],
			BasicType:  driverStmt2,
			IsConn:     false,
		}
		drivers = append(drivers, d)
	}

	wrapperData.Drivers = drivers

	err = tpl.Execute(os.Stdout, wrapperData)

	if err != nil {
		panic(err)
	}

	// printFunctionMaps(arrayConnCopy, arrayStmtCopy)
	// printUtil()
}

func getTypeCombinations(basicTypeForWrapper, prefix string, listOfTheOriginalTypes []string) ([][]string, []string, []string) {
	// generate all subsets of the interfaces
	typesCombinations := combinations.All(listOfTheOriginalTypes)
	var filteredSubsets [][]string

	for _, v := range typesCombinations {
		if inArray2(basicTypeForWrapper, v) && len(v) > 1 {
			filteredSubsets = append(filteredSubsets, removeOnceFromArr2(basicTypeForWrapper, v))
		}
	}

	sort.Slice(filteredSubsets, func(i, j int) bool {
		return len(filteredSubsets[i]) > len(filteredSubsets[j])
	})

	var typeNames []string
	for _, filteredSubset := range filteredSubsets {
		name := strings.ReplaceAll(strings.Join(filteredSubset, "_"), "driver.", "")
		typeNames = append(typeNames, prefix+name)
	}
	listOfTheInterfacesWithoutBasicType := removeOnceFromArr2(basicTypeForWrapper, listOfTheOriginalTypes)

	return filteredSubsets, typeNames, listOfTheInterfacesWithoutBasicType
}

func gen2(basicTypeForWrapper, prefix string, listOfTheOriginalTypes []string) {
	// generate all subsets of the interfaces
	typesCombinations := combinations.All(listOfTheOriginalTypes)
	var filteredSubsets [][]string

	for _, v := range typesCombinations {
		if inArray2(basicTypeForWrapper, v) && len(v) > 1 {
			filteredSubsets = append(filteredSubsets, removeOnceFromArr2(basicTypeForWrapper, v))
		}
	}

	sort.Slice(filteredSubsets, func(i, j int) bool {
		return len(filteredSubsets[i]) > len(filteredSubsets[j])
	})

	var typeNames []string
	// for _, filteredSubset := range filteredSubsets {
	// 	fmt.Println("//", filteredSubset) // just a comment
	// 	// generate types and return type name
	// 	typeNames = append(typeNames, genType2(prefix, filteredSubset, basicTypeForWrapper))
	// }

	genIsAlreadyWrapped2(basicTypeForWrapper, typeNames)

	listOfTheInterfacesWithoutBasicType := removeOnceFromArr2(basicTypeForWrapper, listOfTheOriginalTypes)

	genWrapper2(basicTypeForWrapper, listOfTheInterfacesWithoutBasicType)

	generateAndPrintConstructors2(basicTypeForWrapper, prefix, filteredSubsets, listOfTheInterfacesWithoutBasicType)

	generateConstructorMap2(basicTypeForWrapper, filteredSubsets, listOfTheInterfacesWithoutBasicType)

}

func generateConstructorMap2(basicTypeForWrapper string, filteredSubsets [][]string, listOfTheInterfacesWithoutBasicType []string) {
	for _, subset := range filteredSubsets {
		name := strings.ReplaceAll(strings.Join(subset, "_"), "driver.", "")

		if basicTypeForWrapper == driverStmt2 {
			genConnMap2("w_stmt_", listOfTheInterfacesWithoutBasicType, subset, "get_stmt_"+name)
		} else {
			genConnMap2("w_conn_", listOfTheInterfacesWithoutBasicType, subset, "get_conn_"+name)
		}
	}
}

func generateAndPrintConstructors2(basicTypeForWrapper string, prefix string, filteredSubsets [][]string, listOfTheInterfacesWithoutBasicType []string) {
	var funcs string
	for _, subset := range filteredSubsets {
		name := strings.ReplaceAll(strings.Join(subset, "_"), "driver.", "")

		funcs += generateConstructors2(prefix, basicTypeForWrapper, name, subset, listOfTheInterfacesWithoutBasicType)
	}

	fmt.Println(funcs)
}

// func genType2(prefix string, filteredSubset []string, basicTypeForWrapper string) string {
// 	isColumnConverter := false

// 	name := strings.ReplaceAll(strings.Join(filteredSubset, "_"), "driver.", "")
// 	fmt.Println("type " + prefix + name + " struct {")
// 	fmt.Println(basicTypeForWrapper)
// 	for _, v := range filteredSubset {
// 		if v == "driver.ColumnConverter" {
// 			isColumnConverter = true
// 			fmt.Println("cc ", v)
// 		} else {
// 			fmt.Println(v)
// 		}
// 	}
// 	fmt.Println("}")

// 	if isColumnConverter {
// 		fmt.Printf(`
// func (w *%s) ColumnConverter(idx int) driver.ValueConverter {
// 	return w.cc.ColumnConverter(idx)
// }
// `, prefix+name)
// 	}
// 	return prefix + name
// }

// generate function that checks if basic type is one of the existing wrappers
func genIsAlreadyWrapped2(basicTypeForWrapper string, types []string) {
	noPrefix := strings.ReplaceAll(basicTypeForWrapper, "driver.", "")
	fmt.Printf("func %sAlreadyWrapped(%s %s) bool {\n", strings.ToLower(noPrefix), strings.ToLower(noPrefix), basicTypeForWrapper)
	fmt.Printf("switch %s.(type) {\n", strings.ToLower(noPrefix))

	pointerWrapperTypes := ""
	for _, v := range types {
		pointerWrapperTypes += "*" + v + ","
	}
	pointerWrapperTypes = "*w" + noPrefix + "," + strings.TrimRight(pointerWrapperTypes, ",") + ":"
	fmt.Printf("case %s\n", pointerWrapperTypes)
	fmt.Println("return true")

	wrapperTypes := ""
	for _, v := range types {
		wrapperTypes += v + ","
	}
	wrapperTypes = strings.TrimRight(wrapperTypes, ",") + ":"
	fmt.Printf("case %s\n", wrapperTypes)
	fmt.Println("return true")
	fmt.Println("}")

	fmt.Println("return false")
	fmt.Println("}")
}

func genWrapper2(basicTypeForWrapper string, listOfTheInterfacesWithoutBasicType []string) {

	// generate function definition
	if basicTypeForWrapper == driverStmt2 {
		fmt.Println("func wrapStmt(stmt driver.Stmt, query string, connDetails dbConnDetails, sensor *Sensor) driver.Stmt {")
	} else {
		fmt.Println("func wrapConn(connDetails dbConnDetails, conn driver.Conn, sensor *Sensor) driver.Conn {")
	}

	// we do not need to check the basic interface, because it is a part of the signature
	for _, t := range listOfTheInterfacesWithoutBasicType {
		n := strings.ReplaceAll(t, "driver.", "")
		// generate if statements to find what interfaces type implements
		if basicTypeForWrapper == driverStmt2 {
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
	if basicTypeForWrapper == driverStmt2 {
		fmt.Printf(`if f, ok := _stmt_n[convertBooleansToInt(%s)]; ok {
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
		fmt.Printf(`if f, ok := _conn_n[convertBooleansToInt(%s)]; ok {
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
func generateConstructors2(prefix string, basicTypeForWrapper string, name string, subset []string, listOfTheInterfacesWithoutBasicType []string) string {
	res := ""

	args := ""
	for _, tt := range listOfTheInterfacesWithoutBasicType {
		args += "," + strings.ReplaceAll(tt, "driver.", "") + " " + tt
	}

	// generate function definition
	if basicTypeForWrapper == driverStmt2 {
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
		switch n {
		case "ColumnConverter":
			res += fmt.Sprintf("cc: %s,\n", n)
		case "NamedValueChecker":
			res += fmt.Sprintf("%s: %s,\n", n, n)
		default:
			// add wrapper for a specific type
			res += fmt.Sprintf("%s: &w%s{\n", n, n)
			res += fmt.Sprintf("%s: %s,\n", n, n)
			res += fmt.Sprintf("connDetails: connDetails,\n")
			res += fmt.Sprintf("sensor: sensor,\n")
			if basicTypeForWrapper == driverStmt2 {
				res += fmt.Sprintf("query: query,\n")
			}
			res += fmt.Sprintf("},")
		}
	}

	res += fmt.Sprintf("}\n")
	res += fmt.Sprintf("}\n")
	return res
}

// generate map to do a function lookup
func genConnMap2(prefix string, listOfTheInterfacesWithoutBasicType, subset []string, funcName string) {
	var mask []bool

	for _, v := range listOfTheInterfacesWithoutBasicType {
		if inArray2(v, subset) {
			mask = append(mask, true)
		} else {
			mask = append(mask, false)
		}
	}

	if prefix == "w_stmt_" {
		_stmt_m2[booleansToBinaryRepresentation2(mask...)] = funcName
	} else {
		_conn_m2[booleansToBinaryRepresentation2(mask...)] = funcName
	}

}

func booleansToBinaryRepresentation2(args ...bool) string {
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

func printFunctionMaps2(connTypes []string, stmtTypes []string) {
	ct := removeOnceFromArr2(driverConn2, connTypes)
	st := removeOnceFromArr2(driverStmt2, stmtTypes)

	fmt.Println("var _conn_n = map[int]func( dbConnDetails, driver.Conn, *Sensor," + strings.Join(ct, ",") + ") driver.Conn {")
	for k, v := range _conn_m2 {
		fmt.Println(k, ":", v, ",")
	}
	fmt.Println("}")

	fmt.Println("var _stmt_n = map[int]func( driver.Stmt,  string,  dbConnDetails,  *Sensor,  " + strings.Join(st, ",") + ") driver.Stmt {")
	for k, v := range _stmt_m2 {
		fmt.Println(k, ":", v, ",")
	}
	fmt.Println("}")
}

func printUtil2() {
	fmt.Println(`func convertBooleansToInt(args ...bool) int {
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
