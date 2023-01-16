package main

import (
	"fmt"
	"strings"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
)

type gqlData struct {
	opName   string
	opType   string
	fieldMap map[string][]string
	argMap   map[string][]string
}

func (d gqlData) String() string {
	s := "Operation name: " + d.opName + "\n"
	s += "Operation type:" + d.opType + "\n"
	s += "Fields:\n"

	for k, v := range d.fieldMap {
		s += "\t" + k + ": " + strings.Join(v, ",") + "\n"
	}

	s += "Args:"

	for k, v := range d.argMap {
		s += "\t" + k + ": " + strings.Join(v, ",") + "\n"
	}

	return s
}

func handleField(f *ast.Field) ([]string, []string) {
	fieldMap := []string{}
	argMap := []string{}

	// if f.Name != nil {
	// 	fieldMap[f.Name.Value] = []string{}
	// 	argMap[f.Name.Value] = []string{}
	// }

	// fmt.Println("field:", f.Name.Value)

	for _, arg := range f.Arguments {
		// TODO: check if arg.Name is not nil
		argMap = append(argMap, arg.Name.Value)
	}

	sset := f.GetSelectionSet()

	if sset != nil {
		for _, s := range sset.Selections {
			if field, ok := s.(*ast.Field); ok {
				fieldMap = append(fieldMap, field.Name.Value)
			}
		}
	}

	return fieldMap, argMap
}

func detailQuery(q string) gqlData {
	var data gqlData = gqlData{
		fieldMap: make(map[string][]string),
		argMap:   make(map[string][]string),
	}

	var opName, opType string

	src := source.NewSource(&source.Source{
		Body: []byte(q),
	})

	astDoc, err := parser.Parse(parser.ParseParams{Source: src})

	if err != nil {
		panic(err)
	}

	for _, def := range astDoc.Definitions {
		def := def
		switch df := def.(type) {
		case *ast.OperationDefinition:

			if df.GetName() != nil {
				opName = df.GetName().Value
			}

			opType = df.Operation

			data.opName = opName
			data.opType = opType

			for _, s := range df.GetSelectionSet().Selections {
				s := s
				switch field := s.(type) {
				case *ast.Field:
					fmt.Println(">", field.Name.Value)
					fm, am := handleField(field)

					data.fieldMap[field.Name.Value] = fm
					data.argMap[field.Name.Value] = am
				default:
					fmt.Printf("type is %T\n", field)
				}
			}

		default:
			fmt.Printf("GraphQL cannot execute a request containing a %v\n", def.GetKind())
		}
	}

	fmt.Println("---------------------------")
	return data
}
