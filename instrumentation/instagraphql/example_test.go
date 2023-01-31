// (c) Copyright IBM Corp. 2023

package instagraphql_test

import (
	"context"
	"fmt"
	"log"

	"github.com/graphql-go/graphql"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instagraphql"
)

func Example() {
	// create an instance of the Instana sensor
	sensor := instana.NewSensor("go-graphql")

	// setup GraphQL normally
	fields := graphql.Fields{
		"myfield": &graphql.Field{
			Type: graphql.String,
		},
	}

	rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: fields}
	schemaConfig := graphql.SchemaConfig{Query: graphql.NewObject(rootQuery)}

	schema, err := graphql.NewSchema(schemaConfig)

	if err != nil {
		log.Fatalf("failed to create new schema, error: %v", err)
	}

	query := `query myQuery {
		myfield
	}`

	params := graphql.Params{Schema: schema, RequestString: query}

	// Call instagraphql.Do instead of the original graphql.Do.
	// Make sure to provide a valid context (usually an HTTP req.Context()) if any.
	r := instagraphql.Do(context.Background(), sensor, params)

	fmt.Println("do something with the result", r)
}
