// (c) Copyright IBM Corp. 2023

package instagraphql_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"github.com/stretchr/testify/assert"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"
	"github.com/instana/go-sensor/instrumentation/instagraphql"
	"github.com/stretchr/testify/require"
)

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

var samples = map[string]sampleData{
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

func (s sampleData) queryAsJSON() string {
	q := strings.ReplaceAll(s.query, "\n", " ")
	q = strings.ReplaceAll(q, `"`, `\"`)
	q = strings.ReplaceAll(q, "\t", "")

	return fmt.Sprintf(`{"query": "%s"}`, q)
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

func createField(name string, tp graphql.Output, resolveVal interface{}, args graphql.FieldConfigArgument, pool *pubsub) *graphql.Field {
	return &graphql.Field{
		Name: name,
		Type: tp,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {

			// if args is provided, we assume that it is a mutation
			if args != nil {
				pool.pub(tp.Name(), resolveVal)
			}

			return resolveVal, nil
		},
		Args: args,
	}
}

func getSchema(hasSubscription bool, pool *pubsub) (graphql.Schema, error) {
	qFields := graphql.Fields{
		"aaa": createField("someString", graphql.String, "some string value", nil, pool),
		"row": createField("The row", rowType, row{1, "Row Name", true}, nil, pool),
	}

	mFields := graphql.Fields{
		"insertRow": createField("Add a new row", rowType, row{1, "Row Name", true}, graphql.FieldConfigArgument{
			"name": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"active": &graphql.ArgumentConfig{
				Type: graphql.Boolean,
			},
		}, pool),
	}

	rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: qFields}
	rootMutation := graphql.ObjectConfig{Name: "RootMutation", Fields: mFields}

	schemaConfig := graphql.SchemaConfig{
		Query:    graphql.NewObject(rootQuery),
		Mutation: graphql.NewObject(rootMutation),
	}

	if hasSubscription {
		sFields := graphql.Fields{
			"newRowSubscription": &graphql.Field{
				Name: "row",
				Type: rowType,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return p.Source, nil
				},
				Subscribe: func(p graphql.ResolveParams) (interface{}, error) {
					ch := make(chan interface{})
					rName := rowType.Name()

					pool.sub(rName, ch)

					return ch, nil
				},
			},
		}

		rootSubscription := graphql.ObjectConfig{Name: "RootSubscription", Fields: sFields}
		schemaConfig.Subscription = graphql.NewObject(rootSubscription)
	}

	return graphql.NewSchema(schemaConfig)
}

func assertSample(t *testing.T, sample sampleData, data instana.GraphQLSpanData) {
	assert.Equal(t, sample.spanKind, data.Kind())
	assert.Equal(t, sample.opName, data.Tags.OperationName)
	assert.Equal(t, sample.opType, data.Tags.OperationType)
	assert.Equal(t, sample.hasError, data.Tags.Error != "")
	assert.Equal(t, sample.fields, data.Tags.Fields)
	assert.Equal(t, sample.args, data.Tags.Args)
}

func TestGraphQLWithoutHTTP(t *testing.T) {
	pool := &pubsub{}
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder),
	)

	defer instana.ShutdownSensor()

	schema, err := getSchema(false, pool)

	if err != nil {
		log.Fatalf("failed to create new schema, error: %v", err)
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

			assertSample(t, sample, data)
		})
	}
}

func TestGraphQLWithCustomHTTP(t *testing.T) {
	pool := &pubsub{}
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder),
	)
	defer instana.ShutdownSensor()

	schema, err := getSchema(false, pool)

	if err != nil {
		log.Fatalf("failed to create new schema, error: %v", err)
	}

	for title, sample := range samples {
		t.Run(title, func(t *testing.T) {
			srv := httptest.NewServer(instana.TracingHandlerFunc(sensor, "/graphql", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				b, err := ioutil.ReadAll(req.Body)

				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					io.WriteString(w, err.Error())
					return
				}

				defer req.Body.Close()
				io.CopyN(ioutil.Discard, req.Body, 1<<62)

				var p struct {
					Query         string `json:"query"`
					OperationName string `json:"operationName"`
					Variables     string `json:"variables"`
				}

				err = json.Unmarshal(b, &p)

				if err != nil {
					io.WriteString(w, err.Error())
					return
				}

				params := graphql.Params{Schema: schema, RequestString: p.Query}

				instagraphql.Do(req.Context(), sensor, params)
			})))

			defer srv.Close()

			c := http.DefaultClient
			r := bytes.NewReader([]byte(sample.queryAsJSON()))
			req, _ := http.NewRequest(http.MethodPost, srv.URL, r)
			c.Do(req)

			var spans []instana.Span

			assert.Eventually(t, func() bool {
				return recorder.QueuedSpansCount() >= sample.spanCount
			}, time.Second*2, time.Millisecond*500)

			spans = recorder.GetQueuedSpans()
			assert.Len(t, spans, sample.spanCount)

			require.IsType(t, instana.GraphQLSpanData{}, spans[0].Data)

			data := spans[0].Data.(instana.GraphQLSpanData)

			assertSample(t, sample, data)
		})
	}
}
func TestGraphQLWithBuiltinHTTP(t *testing.T) {
	pool := &pubsub{}
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder),
	)
	defer instana.ShutdownSensor()

	schema, err := getSchema(false, pool)

	if err != nil {
		log.Fatalf("failed to create new schema, error: %v", err)
	}

	for title, sample := range samples {
		t.Run(title, func(t *testing.T) {

			var fn handler.ResultCallbackFn = func(ctx context.Context, params *graphql.Params, result *graphql.Result, responseBody []byte) {
				fmt.Println("I am the original callback function")
			}

			h := handler.New(&handler.Config{
				Schema:           &schema,
				Pretty:           true,
				GraphiQL:         true,
				ResultCallbackFn: instagraphql.ResultCallbackFn(sensor, fn),
			})

			srv := httptest.NewServer(h)

			defer srv.Close()

			c := http.DefaultClient
			r := bytes.NewReader([]byte(sample.queryAsJSON()))
			req, _ := http.NewRequest(http.MethodPost, srv.URL, r)
			c.Do(req)

			var spans []instana.Span

			assert.Eventually(t, func() bool {
				return recorder.QueuedSpansCount() >= sample.spanCount
			}, time.Second*2, time.Millisecond*500)

			spans = recorder.GetQueuedSpans()
			assert.Len(t, spans, sample.spanCount)

			require.IsType(t, instana.GraphQLSpanData{}, spans[0].Data)

			data := spans[0].Data.(instana.GraphQLSpanData)

			assertSample(t, sample, data)
		})
	}
}

func TestGraphQLWithSubscription(t *testing.T) {
	pool := &pubsub{}
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder),
	)
	defer instana.ShutdownSensor()

	schema, err := getSchema(true, pool)

	if err != nil {
		log.Fatalf("failed to create new schema, error: %v", err)
	}

	h := handler.New(&handler.Config{
		Schema:           &schema,
		Pretty:           true,
		GraphiQL:         true,
		ResultCallbackFn: instagraphql.ResultCallbackFn(sensor, nil),
	})

	srv := httptest.NewServer(h)

	defer srv.Close()

	mutSample := sampleData{
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
	}

	subSample := sampleData{
		query: `subscription myRowSubs {
			newRowSubscription {
				name
				active
			}
		}`,
		hasError:  false,
		spanCount: 1,
		spanKind:  instana.ExitSpanKind,
		opName:    "myRowSubs",
		opType:    "subscription-update",
		fields:    map[string][]string{"newRowSubscription": {"name", "active"}},
		args:      map[string][]string{"newRowSubscription": []string(nil)},
	}

	go func() {
		ctx := context.Background()
		q := subSample.query

		subscribeParams := graphql.Params{
			Context:       ctx,
			RequestString: q,
			Schema:        schema,
		}

		instagraphql.Subscribe(ctx, sensor, subscribeParams)
	}()

	c := http.DefaultClient
	r := bytes.NewReader([]byte(mutSample.queryAsJSON()))
	req, _ := http.NewRequest(http.MethodPost, srv.URL, r)
	c.Do(req)

	var spans []instana.Span

	assert.Eventually(t, func() bool {
		return recorder.QueuedSpansCount() == 2
	}, time.Second*2, time.Millisecond*50)

	spans = recorder.GetQueuedSpans()
	assert.Len(t, spans, 2)

	var mutSpan, subSpan instana.Span

	if spans[0].SpanID == spans[0].TraceID {
		mutSpan, subSpan = spans[0], spans[1]
	} else {
		mutSpan, subSpan = spans[1], spans[0]
	}

	require.IsType(t, instana.GraphQLSpanData{}, mutSpan.Data)
	data := mutSpan.Data.(instana.GraphQLSpanData)
	assertSample(t, mutSample, data)

	require.IsType(t, instana.GraphQLSpanData{}, subSpan.Data)
	data = subSpan.Data.(instana.GraphQLSpanData)
	assertSample(t, subSample, data)
}

func TestGraphQLQueryParseError(t *testing.T) {
	pool := &pubsub{}
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder),
	)
	defer instana.ShutdownSensor()

	schema, err := getSchema(false, pool)

	if err != nil {
		log.Fatalf("failed to create new schema, error: %v", err)
	}

	sample := sampleData{
		query: `query myBrokenQuery {
			someField
		`,
		hasError:  true,
		spanCount: 2,
		spanKind:  instana.EntrySpanKind,
		opName:    "",
		opType:    "",
		fields:    nil,
		args:      nil,
	}

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

	assertSample(t, sample, data)
}

type alwaysReadyClient struct{}

func (alwaysReadyClient) Ready() bool                                       { return true }
func (alwaysReadyClient) SendMetrics(data acceptor.Metrics) error           { return nil }
func (alwaysReadyClient) SendEvent(event *instana.EventData) error          { return nil }
func (alwaysReadyClient) SendSpans(spans []instana.Span) error              { return nil }
func (alwaysReadyClient) SendProfiles(profiles []autoprofile.Profile) error { return nil }
func (alwaysReadyClient) Flush(context.Context) error                       { return nil }
