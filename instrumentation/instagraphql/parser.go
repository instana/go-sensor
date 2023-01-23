// (c) Copyright IBM Corp. 2023

package instagraphql

import (
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

func parseEntities(f *ast.Field) (fieldMap, argMap []string) {
	for _, arg := range f.Arguments {
		if arg.Name != nil {
			argMap = append(argMap, arg.Name.Value)
		}
	}

	sset := f.GetSelectionSet()

	if sset != nil {
		for _, s := range sset.Selections {
			if field, ok := s.(*ast.Field); ok {
				fieldMap = append(fieldMap, field.Name.Value)
			}
		}
	}

	return // fieldMap, argMap
}

func parseQuery(q string) (*gqlData, error) {
	var data gqlData = gqlData{
		fieldMap: make(map[string][]string),
		argMap:   make(map[string][]string),
	}

	src := source.NewSource(&source.Source{
		Body: []byte(q),
	})

	astDoc, err := parser.Parse(parser.ParseParams{Source: src})

	if err != nil {
		return nil, err
	}

defLoop:
	for _, def := range astDoc.Definitions {
		switch df := def.(type) {
		case *ast.OperationDefinition:

			if df.GetName() != nil {
				data.opName = df.GetName().Value
			}

			data.opType = df.Operation

			if sel := df.GetSelectionSet().Selections; sel != nil {
				for _, s := range sel {
					switch field := s.(type) {
					case *ast.Field:
						fm, am := parseEntities(field)

						data.fieldMap[field.Name.Value] = fm
						data.argMap[field.Name.Value] = am
					}
				}
			}

			break defLoop
		}
	}

	return &data, nil
}
