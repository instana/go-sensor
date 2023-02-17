// (c) Copyright IBM Corp. 2023

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instagraphql"
)

/*
Query with multiple entities
curl -X POST \
-H "Content-Type: application/json" \
-d '{"query": "query myQuery { characters {id name profession crewMember} ships {name}}"}' \
http://localhost:9191/graphql | jq
*/

/*
curl -X POST \
-H "Content-Type: application/json" \
-d '{"query": "{ characters {id name profession crewMember} }"}' \
http://localhost:9191/graphql | jq
*/

/*
curl -X POST \
-H "Content-Type: application/json" \
-d '{"query": "{ ships {id name origin} }"}' \
http://localhost:9191/graphql
*/

/*
curl -v -X POST \
-H "Content-Type: application/json" \
-d '{"query": "mutation {insertCharacter(name: \"lala char\", profession: \"engineer\", crewMember: true){name}}"}' \
http://localhost:9191/graphql | jq
*/

/*
curl -X POST \
-H "Content-Type: application/json" \
-d '{"query": "mutation {insertShip(name: \"Sheep One\", origin: \"Brazil\") {name origin}}"}' \
http://localhost:9191/graphql | jq
*/

/*
query with error:

curl -X POST \
-H "Content-Type: application/json" \
-d '{"query": "query myQuery { characters {id name profession crewMember naotem } ships {name origin}}"}' \
http://localhost:9191/graphql | jq
*/

var (
	sensor      *instana.Sensor
	withHandler bool
)

func init() {
	sensor = instana.NewSensor("go-graphql-test")
}

type payload struct {
	Query         string `json:"query"`
	OperationName string `json:"operationName"`
	Variables     string `json:"variables"`
}

func handleGraphQLQuery(schema graphql.Schema) http.HandlerFunc {
	fn := func(w http.ResponseWriter, req *http.Request) {
		var query string

		if req.Method == http.MethodGet {
			query = req.URL.Query().Get("query")
		}

		if req.Method == http.MethodPost {
			b, err := ioutil.ReadAll(req.Body)

			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				io.WriteString(w, err.Error())
				return
			}

			defer req.Body.Close()
			io.CopyN(ioutil.Discard, req.Body, 1<<62)

			var p payload

			err = json.Unmarshal(b, &p)

			if err != nil {
				io.WriteString(w, err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			query = p.Query
		}

		if query == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		params := graphql.Params{Schema: schema, RequestString: query}

		r := instagraphql.Do(req.Context(), sensor, params)

		w.Header().Add("Content-Type", "application/json")

		rJSON, _ := json.Marshal(r)
		w.Write(rJSON)
	}

	return instana.TracingHandlerFunc(sensor, "/graphql", fn)
}

// This test application can be started in two different ways:
// * Without a handler: This will start a custom HTTP handler. It can be tested by making HTTP calls to the server.
// Some cRUL examples are provided at the beginning of this file.
// * With a handler from https://github.com/graphql-go/handler: This provides a GraphiQL UI and/or a Playground where
// it makes it easier to test GraphQL queries, including subscriptions.
//
// To start with the handler, run `go run . -handler`
// To start with the custom HTTP handler, run `go run .`
func main() {
	flag.BoolVar(&withHandler, "handler", false, "enables the built-in handler from the graphql API")
	flag.Parse()

	dt, err := loadData()

	if err != nil {
		log.Fatal(err)
	}

	// Schema
	qFields := queries(dt)
	mFields := mutations(dt)
	sFields := subscriptions(dt)

	rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: qFields}
	rootMutation := graphql.ObjectConfig{Name: "RootMutation", Fields: mFields}
	rootSubscription := graphql.ObjectConfig{Name: "RootSubscription", Fields: sFields}

	schemaConfig := graphql.SchemaConfig{
		Query:        graphql.NewObject(rootQuery),
		Mutation:     graphql.NewObject(rootMutation),
		Subscription: graphql.NewObject(rootSubscription),
	}
	schema, err := graphql.NewSchema(schemaConfig)

	if err != nil {
		log.Fatalf("failed to create new schema, error: %v", err)
	}

	if withHandler {
		h := handler.New(&handler.Config{
			Schema:           &schema,
			Pretty:           true,
			GraphiQL:         false,
			Playground:       true,
			ResultCallbackFn: instagraphql.ResultCallbackFn(sensor, nil),
		})

		http.Handle("/graphql", h)
		http.HandleFunc("/subscriptions", SubsHandlerWithSchema(schema))
	} else {
		http.HandleFunc("/graphql", handleGraphQLQuery(schema))
	}

	fmt.Println("Starting app with graphql handler:", withHandler)

	if err := http.ListenAndServe("0.0.0.0:9191", nil); err != nil {
		log.Fatal(err)
	}
}
