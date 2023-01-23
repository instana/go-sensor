// (c) Copyright IBM Corp. 2023

package instagraphql_test

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/stretchr/testify/assert"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"
	"github.com/instana/go-sensor/instrumentation/instagraphql"
	"github.com/stretchr/testify/require"
)

// type user struct {
// 	ID   int    `json:"id"`
// 	Name string `json:"name"`
// }

// type city struct {
// 	ID   int    `json:"id"`
// 	Name string `json:"name"`
// }

type sampleData struct {
	query     string
	hasError  bool
	spanCount int
	spanKind  instana.SpanKind
	opName    string
	opType    string
	fields    map[string][]string
	args      map[string][]string
}

type row struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

var rowType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Row",
	Fields: graphql.Fields{
		"id":     &graphql.Field{Type: graphql.Int},
		"name":   &graphql.Field{Type: graphql.String},
		"active": &graphql.Field{Type: graphql.Boolean},
	},
})

func createField(name string, tp graphql.Output, resolveVal interface{}, args graphql.FieldConfigArgument) *graphql.Field {
	return &graphql.Field{
		Name: name,
		Type: tp,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			return resolveVal, nil
		},
		Args: args,
	}
}

func TestGraphQLServer(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder),
	)

	qFields := graphql.Fields{
		"aaa": createField("someString", graphql.String, "some string value", nil),
		"row": createField("The row", rowType, row{1, "Row Name", true}, nil),
	}

	mFields := graphql.Fields{
		"insertRow": createField("Add a new row", rowType, row{1, "Row Name", true}, graphql.FieldConfigArgument{
			"name": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"active": &graphql.ArgumentConfig{
				Type: graphql.Boolean,
			},
		}),
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

	samples := map[string]sampleData{
		"query success": {
			query: `query myQuery {
				aaa
			}`,
			hasError:  false,
			spanCount: 1,
			spanKind:  instana.EntrySpanKind,
			opName:    "myQuery",
			opType:    "query",
			fields:    map[string][]string{"aaa": nil},
			args:      map[string][]string{"aaa": nil},
		},
		"query error": {
			query: `query myQuery {
				aaa { invalidField }
			}`,
			hasError:  true,
			spanCount: 2,
			spanKind:  instana.EntrySpanKind,
			opName:    "myQuery",
			opType:    "query",
			fields:    map[string][]string{"aaa": {"invalidField"}},
			args:      map[string][]string{"aaa": nil},
		},
		"query object type": {
			query: `query getRow {
				row { id name active }
			}`,
			hasError:  false,
			spanCount: 1,
			spanKind:  instana.EntrySpanKind,
			opName:    "getRow",
			opType:    "query",
			fields:    map[string][]string{"row": {"id", "name", "active"}},
			args:      map[string][]string{"row": nil},
		},
		"mutation object type": {
			query: `mutation newRow {
				insertRow(name: "row two", active: true) {
					id
					name
					active
				}
			}`,
			hasError:  false,
			spanCount: 1,
			spanKind:  instana.EntrySpanKind,
			opName:    "newRow",
			opType:    "mutation",
			fields:    map[string][]string{"insertRow": {"id", "name", "active"}},
			args:      map[string][]string{"insertRow": {"name", "active"}},
		},
	}

	for title, sample := range samples {
		t.Run(title, func(t *testing.T) {
			params := graphql.Params{Schema: schema, RequestString: sample.query}

			instagraphql.Do(context.Background(), sensor, params)

			var spans []instana.Span

			assert.Eventually(t, func() bool {
				return recorder.QueuedSpansCount() == sample.spanCount
			}, time.Second*2, time.Millisecond*500)

			spans = recorder.GetQueuedSpans()
			assert.Len(t, spans, sample.spanCount)

			require.IsType(t, instana.GraphQLSpanData{}, spans[0].Data)

			data := spans[0].Data.(instana.GraphQLSpanData)

			assert.Equal(t, sample.spanKind, data.Kind())
			assert.Equal(t, sample.opName, data.Tags.OperationName)
			assert.Equal(t, sample.opType, data.Tags.OperationType)
			assert.Equal(t, sample.hasError, data.Tags.Error != "")
			assert.Equal(t, sample.fields, data.Tags.Fields)
			assert.Equal(t, sample.args, data.Tags.Args)
		})
	}
}

type alwaysReadyClient struct{}

func (alwaysReadyClient) Ready() bool                                       { return true }
func (alwaysReadyClient) SendMetrics(data acceptor.Metrics) error           { return nil }
func (alwaysReadyClient) SendEvent(event *instana.EventData) error          { return nil }
func (alwaysReadyClient) SendSpans(spans []instana.Span) error              { return nil }
func (alwaysReadyClient) SendProfiles(profiles []autoprofile.Profile) error { return nil }
func (alwaysReadyClient) Flush(context.Context) error                       { return nil }
