// (c) Copyright IBM Corp. 2023

package main

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/graphql-go/graphql"
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
http://localhost:9191/graphql
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

func main() {
	sensor := instana.NewSensor("go-graphql-test")

	dt, err := loadData()

	if err != nil {
		log.Fatal(err)
	}

	// Schema
	qFields := queriesWithoutResolve()

	qFields["characters"].Resolve = func(p graphql.ResolveParams) (interface{}, error) {
		if id, ok := p.Args["id"].(int); ok {
			return []*character{dt.findChar(id)}, nil
		}

		return dt.Chars, nil
	}

	qFields["ships"].Resolve = func(p graphql.ResolveParams) (interface{}, error) {
		if id, ok := p.Args["id"].(int); ok {
			return []*ship{dt.findShip(id)}, nil
		}

		return dt.Ships, nil
	}

	mFields := mutationsWithoutResolve()

	mFields["insertCharacter"].Resolve = func(p graphql.ResolveParams) (interface{}, error) {
		var name, profession string
		var ok, crewMember bool

		if name, ok = p.Args["name"].(string); !ok {
			return nil, errors.New("name not found")
		}

		if profession, ok = p.Args["profession"].(string); !ok {
			return nil, errors.New("profession not found")
		}

		if crewMember, ok = p.Args["crewMember"].(bool); !ok {
			return nil, errors.New("crewMember not found")
		}

		c := character{
			Name:       name,
			Profession: profession,
			CrewMember: crewMember,
		}

		dt.addChar(c)

		return c, nil
	}

	mFields["insertShip"].Resolve = func(p graphql.ResolveParams) (interface{}, error) {
		var name, origin string
		var ok bool

		if name, ok = p.Args["name"].(string); !ok {
			return nil, errors.New("name not found")
		}

		if origin, ok = p.Args["origin"].(string); !ok {
			return nil, errors.New("origin not found")
		}

		s := ship{
			Name:   name,
			Origin: origin,
		}

		dt.addShip(s)

		return s, nil
	}

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

	http.HandleFunc("/graphql", handleGraphQLQuery(schema, sensor))

	if err := http.ListenAndServe("0.0.0.0:9191", nil); err != nil {
		log.Fatal(err)
	}
}
