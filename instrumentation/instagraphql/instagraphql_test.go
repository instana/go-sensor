// (c) Copyright IBM Corp. 2023

package instagraphql_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/graphql-go/graphql"
)

// type user struct {
// 	ID   int    `json:"id"`
// 	Name string `json:"name"`
// }

// type city struct {
// 	ID   int    `json:"id"`
// 	Name string `json:"name"`
// }

func createField(name string, tp *graphql.Scalar, resolveVal interface{}, err error) *graphql.Field {
	return &graphql.Field{
		Name: name,
		Type: tp,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			if err != nil {
				return nil, err
			}
			return resolveVal, nil
		},
	}
}

func TestGraphQLBasic(t *testing.T) {
	qFields := graphql.Fields{
		"aaa": createField("someString", graphql.String, "some string value", nil),
	}

	rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: qFields}

	schemaConfig := graphql.SchemaConfig{
		Query: graphql.NewObject(rootQuery),
	}

	schema, err := graphql.NewSchema(schemaConfig)

	query := `query myQuery {
		aaa
	}`

	params := graphql.Params{Schema: schema, RequestString: query}

	if err != nil {
		log.Fatalf("failed to create new schema, error: %v", err)
	}

	res := graphql.Do(params)

	fmt.Println(res)
}
