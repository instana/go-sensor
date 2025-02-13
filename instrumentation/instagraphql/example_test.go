// (c) Copyright IBM Corp. 2023

package instagraphql_test

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instagraphql"
)

func ExampleDo() {
	// create an instance of the Instana collector
	c := instana.InitCollector(&instana.Options{
		Service: "go-graphql",
	})

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
	r := instagraphql.Do(context.Background(), c, params)

	fmt.Println("do something with the result", r)
}

func ExampleResultCallbackFn() {
	// create an instance of the Instana collector
	c := instana.InitCollector(&instana.Options{
		Service: "go-graphql",
	})
	defer instana.ShutdownCollector()

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

	// Setup the handler

	// The original ResultCallbackFn function if you have one. Otherwise, just pass nil
	var fn handler.ResultCallbackFn = func(ctx context.Context, params *graphql.Params, result *graphql.Result, responseBody []byte) {
		fmt.Println("Original callback function")
	}

	h := handler.New(&handler.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: true,
		// instagraphql.ResultCallbackFn instruments the code and returns the original function, if any.
		ResultCallbackFn: instagraphql.ResultCallbackFn(c, fn),
	})

	http.Handle("/graphql", h)

	if err := http.ListenAndServe("0.0.0.0:9191", nil); err != nil {
		log.Fatal(err)
	}
}
