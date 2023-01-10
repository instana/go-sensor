// (c) Copyright IBM Corp. 2023
//go:build go1.18

//go:generate sh -c "go run . > ../../../sql_wrappers.go && go fmt ../../../sql_wrappers.go"

package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/template"

	combinations "github.com/mxschmitt/golang-combinations"
)

const driverStmt = "driver.Stmt"
const driverConn = "driver.Conn"

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
}

var arrayStmt = []string{
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
	return inSlice("driver.ColumnConverter", d.Interfaces)
}

type WrapperData struct {
	Drivers []DriverCombo
}

// Returns a boolean indicating that el was found in the slice.
func inSlice(el string, arr []string) bool {
	for _, v := range arr {
		if v == el {
			return true
		}
	}

	return false
}

// Returns a new slice without the first occurrence of the matched element el
func removeElementOnce(el string, arr []string) []string {
	for k, v := range arr {
		if v == el {
			var res []string
			_ = copy(res, arr[:k])
			return append(res, arr[k+1:]...)
		}
	}
	return arr
}

var funcMap template.FuncMap
var connInterfacesNoBasicType []string
var stmtInterfacesNoBasicType []string

func init() {
	connInterfacesNoBasicType = removeElementOnce(driverConn, arrayConn)
	stmtInterfacesNoBasicType = removeElementOnce(driverStmt, arrayStmt)

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
			return _stmt_m
		},
		"connMap": func() map[string]string {
			return _conn_m
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

func main() {
	tpl, err := template.New("sql_wrappers.tpl").Funcs(funcMap).ParseFiles("sql_wrappers.tpl")

	if err != nil {
		panic(err)
	}

	// Builds type names for all combinations
	connSubsets := getFilteredSubsets(driverConn, arrayConn)
	stmtSubsets := getFilteredSubsets(driverStmt, arrayStmt)

	// Feeds maps for conn and stmt where the key is a string (eg: 0b1001) and the value is the constructor name
	for _, subset := range connSubsets {
		name := strings.ReplaceAll(strings.Join(subset, "_"), "driver.", "")
		_conn_m[mapKey(connInterfacesNoBasicType, subset)] = "get_conn_" + name
	}

	for _, subset := range stmtSubsets {
		name := strings.ReplaceAll(strings.Join(subset, "_"), "driver.", "")
		_stmt_m[mapKey(stmtInterfacesNoBasicType, subset)] = "get_stmt_" + name
	}

	var drivers []DriverCombo

	// List of all possible types for driver.Conn and driver.Stmt
	connTypes := getTypeNames(driverConn, "conn_", connSubsets)
	stmtTypes := getTypeNames(driverStmt, "stmt_", stmtSubsets)

	for i := 0; i < len(connTypes); i++ {
		d := DriverCombo{
			TypeName:   connTypes[i],
			Interfaces: connSubsets[i],
			BasicType:  driverConn,
			IsConn:     true,
		}
		drivers = append(drivers, d)
	}

	for i := 0; i < len(stmtTypes); i++ {
		d := DriverCombo{
			TypeName:   stmtTypes[i],
			Interfaces: stmtSubsets[i],
			BasicType:  driverStmt,
			IsConn:     false,
		}
		drivers = append(drivers, d)
	}

	if err = tpl.Execute(os.Stdout, WrapperData{drivers}); err != nil {
		panic(err)
	}
}

// Builds names of each generated type based on the filtered subset and returns a slice of strings with the names.
func getTypeNames(basicTypeForWrapper, prefix string, filteredSubsets [][]string) []string {
	var typeNames []string
	for _, filteredSubset := range filteredSubsets {
		name := strings.ReplaceAll(strings.Join(filteredSubset, "_"), "driver.", "")
		typeNames = append(typeNames, prefix+name)
	}

	return typeNames
}

// Returns a slice of all combinations from listOfTheOriginalTypes without repeating basicTypeForWrapper
//
// Example:
//
//	listOfTheOriginalTypes = ['a', 'b']
//	basicTypeForWrapper = 'a'
//
//	All combinations: ['a'], ['b'], ['a', 'b']
//	Filtered subsets: ['b']
func getFilteredSubsets(basicTypeForWrapper string, listOfTheOriginalTypes []string) [][]string {
	// generate all subsets of the interfaces
	typesCombinations := combinations.All(listOfTheOriginalTypes)
	var filteredSubsets [][]string

	for _, v := range typesCombinations {
		if inSlice(basicTypeForWrapper, v) && len(v) > 1 {
			filteredSubsets = append(filteredSubsets, removeElementOnce(basicTypeForWrapper, v))
		}
	}

	sort.Slice(filteredSubsets, func(i, j int) bool {
		return len(filteredSubsets[i]) > len(filteredSubsets[j])
	})

	return filteredSubsets
}

// Identifies which interfaces the type implements and returns a binary key used in the maps
func mapKey(listOfTheInterfacesWithoutBasicType, subset []string) string {
	var mask []bool

	for _, v := range listOfTheInterfacesWithoutBasicType {
		if inSlice(v, subset) {
			mask = append(mask, true)
		} else {
			mask = append(mask, false)
		}
	}

	return booleansToBinaryRepresentation(mask...)
}

// Receives an array of booleans and returns a binary representation as string. eg: [true, false, false, true] = "1001"
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
