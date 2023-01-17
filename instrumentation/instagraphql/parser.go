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

type fieldsEntity []string
type argsEntity []string

func parseEntities(f *ast.Field) (fieldsEntity, argsEntity) {
	fieldMap := []string{}
	argMap := []string{}

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

	return fieldMap, argMap
}

func parseQuery(q string) gqlData {
	var data gqlData = gqlData{
		fieldMap: make(map[string][]string),
		argMap:   make(map[string][]string),
	}

	src := source.NewSource(&source.Source{
		Body: []byte(q),
	})

	astDoc, err := parser.Parse(parser.ParseParams{Source: src})

	if err != nil {
		panic(err)
	}

	for _, def := range astDoc.Definitions {
		switch df := def.(type) {
		case *ast.OperationDefinition:

			if df.GetName() != nil {
				data.opName = df.GetName().Value
			}

			data.opType = df.Operation

			for _, s := range df.GetSelectionSet().Selections {
				switch field := s.(type) {
				case *ast.Field:
					fm, am := parseEntities(field)

					data.fieldMap[field.Name.Value] = fm
					data.argMap[field.Name.Value] = am
				}
			}
		}
	}

	return data
}
