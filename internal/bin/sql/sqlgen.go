// (c) Copyright IBM Corp. 2022
//go:build go1.18

//go:generate sh -c "go run . > ../../../sql_wrappers.go && go fmt ../../../sql_wrappers.go"

package main

import (
	"os"
	"sort"
	"strings"
	"text/template"

	combinations "github.com/mxschmitt/golang-combinations"
)

const driverStmt = "driver.Stmt"
const driverConn = "driver.Conn"

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
