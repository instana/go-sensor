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
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder),
	)

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

	instagraphql.Do(context.Background(), sensor, params)

	var spans []instana.Span

	assert.Eventually(t, func() bool {
		return recorder.QueuedSpansCount() == 1
	}, time.Second*2, time.Millisecond*500)

	spans = recorder.GetQueuedSpans()
	assert.Len(t, spans, 1)

	require.IsType(t, instana.GraphQLSpanData{}, spans[0].Data)

	data := spans[0].Data.(instana.GraphQLSpanData)

	assert.Equal(t, instana.EntrySpanKind, data.Kind())
	assert.Equal(t, "myQuery", data.Tags.OperationName)
	assert.Equal(t, "query", data.Tags.OperationType)
	assert.Equal(t, "", data.Tags.Error)
	assert.Equal(t, map[string][]string{"aaa": nil}, data.Tags.Fields)
	assert.Equal(t, map[string][]string{"aaa": nil}, data.Tags.Args)
}

type alwaysReadyClient struct{}

func (alwaysReadyClient) Ready() bool                                       { return true }
func (alwaysReadyClient) SendMetrics(data acceptor.Metrics) error           { return nil }
func (alwaysReadyClient) SendEvent(event *instana.EventData) error          { return nil }
func (alwaysReadyClient) SendSpans(spans []instana.Span) error              { return nil }
func (alwaysReadyClient) SendProfiles(profiles []autoprofile.Profile) error { return nil }
func (alwaysReadyClient) Flush(context.Context) error                       { return nil }
