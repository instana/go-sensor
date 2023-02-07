// (c) Copyright IBM Corp. 2023

package main

import (
	"context"
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

type payload struct {
	Query         string `json:"query"`
	OperationName string `json:"operationName"`
	Variables     string `json:"variables"`
}

func handleGraphQLQuery(schema graphql.Schema, sensor *instana.Sensor) http.HandlerFunc {
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

var withHandler bool

func main() {
	flag.BoolVar(&withHandler, "handler", false, "enables the built-in handler from the graphql API")
	flag.Parse()

	sensor := instana.NewSensor("go-graphql-test")

	dt, err := loadData()

	if err != nil {
		log.Fatal(err)
	}

	// Schema
	qFields := queries(dt)
	mFields := mutations(dt)

	rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: qFields}
	rootMutation := graphql.ObjectConfig{Name: "RootMutation", Fields: mFields}
	schemaConfig := graphql.SchemaConfig{
		Query:    graphql.NewObject(rootQuery),
		Mutation: graphql.NewObject(rootMutation),
	}
	schema, err := graphql.NewSchema(schemaConfig)

	if err != nil {
		log.Fatalf("failed to create new schema, error: %v", err)
	}

	if withHandler {
		var fn handler.ResultCallbackFn = func(ctx context.Context, params *graphql.Params, result *graphql.Result, responseBody []byte) {
			fmt.Println("I am the original callback function")
		}

		h := handler.New(&handler.Config{
			Schema:           &schema,
			Pretty:           true,
			GraphiQL:         true,
			ResultCallbackFn: instagraphql.ResultCallbackFn(sensor, fn),
		})

		http.Handle("/graphql", h)
	} else {
		http.HandleFunc("/graphql", handleGraphQLQuery(schema, sensor))
	}

	fmt.Println("Starting app with graphql handler:", withHandler)

	if err := http.ListenAndServe("0.0.0.0:9191", nil); err != nil {
		log.Fatal(err)
	}
}
