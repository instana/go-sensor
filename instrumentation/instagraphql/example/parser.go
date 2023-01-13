package main

import (
	"fmt"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
)

func handleField(f *ast.Field) {
	fmt.Println("field:", f.Name.Value)

	for _, arg := range f.Arguments {
		fmt.Println("arg:", arg.Name.Value, "=", arg.Value.GetValue())
	}

	sset := f.GetSelectionSet()

	if sset != nil {
		for _, s := range sset.Selections {
			if field, ok := s.(*ast.Field); ok {
				handleField(field)
			}
		}
	}
}

func handleVarDefs(varDefs []*ast.VariableDefinition) {
	for _, vd := range varDefs {
		fmt.Printf("var def var value: %v, type %[1]T\n", vd.Variable.GetValue())
		fmt.Printf("var def var name: %v, type %[1]T\n", vd.Variable.GetName())
	}
}

func detailQuery(q string) {
	src := source.NewSource(&source.Source{
		Body: []byte(q),
		Name: "test_name",
	})

	astDoc, err := parser.Parse(parser.ParseParams{Source: src})

	if err != nil {
		panic(err)
	}

	for _, def := range astDoc.Definitions {
		switch df := def.(type) {
		case *ast.OperationDefinition:
			fmt.Println("operation: ", df.Operation)
			fmt.Println("name: ", df.Name)
			handleVarDefs(df.GetVariableDefinitions())

			for _, s := range df.GetSelectionSet().Selections {
				switch field := s.(type) {
				case *ast.Field:
					handleField(field)
				default:
					fmt.Printf("type is %T\n", field)
				}
			}

		case *ast.FragmentDefinition:
			fmt.Println("case FragmentDefinition:", df)
		default:
			fmt.Printf("GraphQL cannot execute a request containing a %v\n", def.GetKind())
		}
	}

	fmt.Println("---------------------------")
}
