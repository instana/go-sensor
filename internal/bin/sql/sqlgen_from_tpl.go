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

var funcMap template.FuncMap
var connInterfacesNoBasicType []string
var stmtInterfacesNoBasicType []string

func init() {
	funcMap = template.FuncMap{
		"replace": strings.ReplaceAll,
		"join":    strings.Join,
		"connInterfaces": func() []string {
			return connInterfacesNoBasicType
		},
		"stmtInterfaces": func() []string {
			return stmtInterfacesNoBasicType
		},
		"stmtMap": func() map[string]string {
			return _stmt_m2
		},
		"connMap": func() map[string]string {
			return _conn_m2
		},
		"driverTypes": func(dc []DriverCombo, isConn bool) []string {
			var drivers []string

			for _, d := range dc {
				if d.IsConn == isConn {
					drivers = append(drivers, d.TypeName)
				}
			}

			return drivers
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

	for _, subset := range connSubsets {
		name := strings.ReplaceAll(strings.Join(subset, "_"), "driver.", "")
		genConnMap2("w_conn_", connInterfacesNoBasicType, subset, "get_conn_"+name)
	}

	for _, subset := range stmtSubsets {
		name := strings.ReplaceAll(strings.Join(subset, "_"), "driver.", "")
		genConnMap2("w_stmt_", stmtInterfacesNoBasicType, subset, "get_stmt_"+name)
	}

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
