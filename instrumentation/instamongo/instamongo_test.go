// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

package instamongo_test

import (
	"context"
	"log"
	"sync"
	"testing"

	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instamongo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestWrapCommandMonitor_Succeeded(t *testing.T) {
	examples := map[string]struct {
		Database     string
		Command      string
		Data         interface{}
		ConnectionID string
		Expected     instana.MongoDBSpanTags
	}{
		"find": {
			Database: "testing-db",
			Command:  "find",
			Data: bson.M{
				"find":   "records",
				"filter": bson.D{bson.E{"value", "42"}, bson.E{"valid", true}},
			},
			ConnectionID: "localhost:27017",
			Expected: instana.MongoDBSpanTags{
				Service:   "localhost:27017",
				Namespace: "testing-db.records",
				Command:   "find",
				Filter:    `{"value": "42", "valid": true}`,
			},
		},
		"insert": {
			Database: "testing-db",
			Command:  "insert",
			Data: bson.M{
				"insert": "records",
				"documents": bson.A{
					bson.D{bson.E{"value", 1}},
					bson.D{bson.E{"value", 2}},
				},
			},
			ConnectionID: "localhost:27017",
			Expected: instana.MongoDBSpanTags{
				Service:   "localhost:27017",
				Namespace: "testing-db.records",
				Command:   "insert",
				JSON:      `[{"value": 1}, {"value": 2}]`,
			},
		},
		"update": {
			Database: "testing-db",
			Command:  "update",
			Data: bson.M{
				"update": "records",
				"query":  bson.D{bson.E{"valid", false}},
				"updates": bson.A{
					bson.D{
						bson.E{"q", bson.D{bson.E{"valid", false}}},
						bson.E{"u", bson.D{
							bson.E{"valid", true},
							bson.E{"value", 0},
						}},
					},
				},
			},
			ConnectionID: "localhost:27017",
			Expected: instana.MongoDBSpanTags{
				Service:   "localhost:27017",
				Namespace: "testing-db.records",
				Command:   "update",
				JSON:      `[{"q": {"valid": false}, "u": {"value": 0, "valid": true}}]`,
			},
		},
		"delete": {
			Database: "testing-db",
			Command:  "delete",
			Data: bson.M{
				"delete": "records",
				"deletes": bson.A{
					bson.D{bson.E{"q", bson.D{bson.E{"valid", false}}}},
					bson.D{bson.E{"q", bson.D{bson.E{"value", 0}}}, bson.E{"limit", 1}},
				},
			},
			ConnectionID: "localhost:27017",
			Expected: instana.MongoDBSpanTags{
				Service:   "localhost:27017",
				Namespace: "testing-db.records",
				Command:   "delete",
				JSON:      `[{"q": {"valid": false}}, {"q": {"value": 0}, "limit": 1}]`,
			},
		},
		"aggregate": {
			Database: "testing-db",
			Command:  "aggregate",
			Data: bson.M{
				"aggregate": "records",
				"pipeline": bson.A{
					bson.D{
						bson.E{"$search", bson.D{bson.E{"valid", true}}},
						bson.E{"$sum", "value"},
					},
				},
			},
			ConnectionID: "localhost:27017",
			Expected: instana.MongoDBSpanTags{
				Service:   "localhost:27017",
				Namespace: "testing-db.records",
				Command:   "aggregate",
				JSON:      `[{"$search": {"valid": true}, "$sum": "value"}]`,
			},
		},
		"mapReduce": {
			Database: "testing-db",
			Command:  "mapReduce",
			Data: bson.M{
				"mapReduce": "records",
				"map":       "function () { this.tags.forEach(function(z) { emit(z, 1); }); }",
				"reduce":    "function (key, values) { return len(values); }",
			},
			ConnectionID: "localhost:27017",
			Expected: instana.MongoDBSpanTags{
				Service:   "localhost:27017",
				Namespace: "testing-db.records",
				Command:   "mapReduce",
				JSON:      `{"map": "function () { this.tags.forEach(function(z) { emit(z, 1); }); }", "reduce": "function (key, values) { return len(values); }"}`,
			},
		},
		"network connection with id": {
			Database: "testing-db",
			Command:  "listDatabases",
			Data: bson.M{
				"listDatabases": 1,
			},
			ConnectionID: "mongo-host:12345-123",
			Expected: instana.MongoDBSpanTags{
				Service:   "mongo-host:12345",
				Namespace: "testing-db",
				Command:   "listDatabases",
			},
		},
		"unix socket": {
			Database: "testing-db",
			Command:  "listDatabases",
			Data: bson.M{
				"listDatabases": 1,
			},
			ConnectionID: "/var/run/mongo.sock",
			Expected: instana.MongoDBSpanTags{
				Service:   "/var/run/mongo.sock",
				Namespace: "testing-db",
				Command:   "listDatabases",
			},
		},
		"unix socket with id": {
			Database: "testing-db",
			Command:  "listDatabases",
			Data: bson.M{
				"listDatabases": 1,
			},
			ConnectionID: "/var/run/mongo.sock-123",
			Expected: instana.MongoDBSpanTags{
				Service:   "/var/run/mongo.sock",
				Namespace: "testing-db",
				Command:   "listDatabases",
			},
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			recorder := instana.NewTestRecorder()
			sensor := instana.NewSensorWithTracer(
				instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder),
			)
			defer instana.ShutdownSensor()
			mon := &monitorMock{}

			m := instamongo.WrapCommandMonitor(mon.Monitor(), sensor)

			sp := sensor.Tracer().StartSpan("testing")
			ctx := instana.ContextWithSpan(context.Background(), sp)

			started := &event.CommandStartedEvent{
				DatabaseName: example.Database,
				Command:      marshalBSON(t, example.Data),
				CommandName:  example.Command,
				RequestID:    1,
				ConnectionID: example.ConnectionID,
			}
			m.Started(ctx, started)

			success := &event.CommandSucceededEvent{
				CommandFinishedEvent: event.CommandFinishedEvent{
					DurationNanos: 123,
					CommandName:   example.Command,
					RequestID:     1,
					ConnectionID:  example.ConnectionID,
				},
				Reply: marshalBSON(t, bson.M{
					"databases": bson.A{},
				}),
			}
			m.Succeeded(context.Background(), success)

			sp.Finish()

			spans := recorder.GetQueuedSpans()
			require.Len(t, spans, 2)

			dbSpan, parentSpan := spans[0], spans[1]

			assert.Equal(t, parentSpan.TraceID, dbSpan.TraceID)
			assert.Equal(t, parentSpan.TraceIDHi, dbSpan.TraceIDHi)
			assert.Equal(t, parentSpan.SpanID, dbSpan.ParentID)

			assert.Equal(t, "mongo", dbSpan.Name)
			assert.EqualValues(t, instana.ExitSpanKind, dbSpan.Kind)
			assert.Empty(t, dbSpan.Ec)

			require.IsType(t, instana.MongoDBSpanData{}, dbSpan.Data)

			data := dbSpan.Data.(instana.MongoDBSpanData)

			// Check JSON fields separately to ignore formatting
			if example.Expected.Query != "" {
				assert.JSONEq(t, example.Expected.Query, data.Tags.Query)
			}

			if example.Expected.Filter != "" {
				assert.JSONEq(t, example.Expected.Filter, data.Tags.Filter)
			}

			if example.Expected.JSON != "" {
				assert.JSONEq(t, example.Expected.JSON, data.Tags.JSON)
			}

			data.Tags.Query = example.Expected.Query
			data.Tags.Filter = example.Expected.Filter
			data.Tags.JSON = example.Expected.JSON

			assert.Equal(t, example.Expected, data.Tags)

			// Check that events were propagated to the original CommandMonitor
			assert.Equal(t, []interface{}{started, success}, mon.Events())
		})
	}
}

func TestWrapCommandMonitor_Succeeded_NotStarted(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder),
	)
	defer instana.ShutdownSensor()
	mon := &monitorMock{}

	m := instamongo.WrapCommandMonitor(mon.Monitor(), sensor)

	success := &event.CommandSucceededEvent{
		CommandFinishedEvent: event.CommandFinishedEvent{
			DurationNanos: 123,
			CommandName:   "listDatabases",
			RequestID:     1,
			ConnectionID:  "localhost:27017-1",
		},
		Reply: marshalBSON(t, bson.M{
			"databases": bson.A{},
		}),
	}
	m.Succeeded(context.Background(), success)

	assert.Empty(t, recorder.GetQueuedSpans())

	// Check that events were propagated to the original CommandMonitor
	assert.Equal(t, []interface{}{success}, mon.Events())
}

func TestWrapCommandMonitor_Succeeded_NotTraced(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder),
	)
	defer instana.ShutdownSensor()
	mon := &monitorMock{}

	m := instamongo.WrapCommandMonitor(mon.Monitor(), sensor)

	started := &event.CommandStartedEvent{
		Command: marshalBSON(t, bson.M{
			"listDatabases": 1,
		}),
		CommandName:  "listDatabases",
		RequestID:    1,
		ConnectionID: "localhost:27017-1",
	}
	m.Started(context.Background(), started)

	success := &event.CommandSucceededEvent{
		CommandFinishedEvent: event.CommandFinishedEvent{
			DurationNanos: 123,
			CommandName:   "listDatabases",
			RequestID:     1,
			ConnectionID:  "localhost:27017-1",
		},
		Reply: marshalBSON(t, bson.M{
			"databases": bson.A{},
		}),
	}
	m.Succeeded(context.Background(), success)

	assert.Empty(t, recorder.GetQueuedSpans())

	// Check that events were propagated to the original CommandMonitor
	assert.Equal(t, []interface{}{started, success}, mon.Events())
}

func TestWrapCommandMonitor_Failed(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder),
	)
	defer instana.ShutdownSensor()
	mon := &monitorMock{}

	m := instamongo.WrapCommandMonitor(mon.Monitor(), sensor)

	sp := sensor.Tracer().StartSpan("testing")
	ctx := instana.ContextWithSpan(context.Background(), sp)

	started := &event.CommandStartedEvent{
		DatabaseName: "testing-db",
		Command:      marshalBSON(t, bson.M{"listDatabases": 1}),
		CommandName:  "listDatabases",
		RequestID:    1,
		ConnectionID: "mongo-host:12345-123",
	}
	m.Started(ctx, started)

	failed := &event.CommandFailedEvent{
		CommandFinishedEvent: event.CommandFinishedEvent{
			DurationNanos: 123,
			CommandName:   "listDatabases",
			RequestID:     1,
			ConnectionID:  "mongo-host:12345-123",
		},
		Failure: "something went wrong",
	}
	m.Failed(context.Background(), failed)

	sp.Finish()

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 3)

	dbSpan, logSpan, parentSpan := spans[0], spans[1], spans[2]

	assert.Equal(t, parentSpan.TraceID, dbSpan.TraceID)
	assert.Equal(t, parentSpan.TraceIDHi, dbSpan.TraceIDHi)
	assert.Equal(t, parentSpan.SpanID, dbSpan.ParentID)

	assert.Equal(t, "mongo", dbSpan.Name)
	assert.EqualValues(t, instana.ExitSpanKind, dbSpan.Kind)
	assert.Equal(t, 1, dbSpan.Ec)

	require.IsType(t, instana.MongoDBSpanData{}, dbSpan.Data)

	data := dbSpan.Data.(instana.MongoDBSpanData)
	assert.Equal(t, instana.MongoDBSpanTags{
		Service:   "mongo-host:12345",
		Namespace: "testing-db",
		Command:   "listDatabases",
		Error:     "something went wrong",
	}, data.Tags)

	assert.Equal(t, dbSpan.TraceID, logSpan.TraceID)
	assert.Equal(t, dbSpan.SpanID, logSpan.ParentID)
	assert.Equal(t, "log.go", logSpan.Name)

	// Assert that log message has been recorded within the span interval
	assert.GreaterOrEqual(t, logSpan.Timestamp, dbSpan.Timestamp)
	assert.LessOrEqual(t, logSpan.Duration, dbSpan.Duration)

	require.IsType(t, instana.LogSpanData{}, logSpan.Data)
	logData := logSpan.Data.(instana.LogSpanData)

	assert.Equal(t, instana.LogSpanTags{
		Level:   "ERROR",
		Message: `error: "something went wrong"`,
	}, logData.Tags)

	// Check that events were propagated to the original CommandMonitor
	assert.Equal(t, []interface{}{started, failed}, mon.Events())
}

func TestWrapCommandMonitor_Failed_NotStarted(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder),
	)
	defer instana.ShutdownSensor()
	mon := &monitorMock{}

	m := instamongo.WrapCommandMonitor(mon.Monitor(), sensor)

	failed := &event.CommandFailedEvent{
		CommandFinishedEvent: event.CommandFinishedEvent{
			DurationNanos: 123,
			CommandName:   "listDatabases",
			RequestID:     1,
			ConnectionID:  "localhost:27017-1",
		},
		Failure: "something went wrong",
	}
	m.Failed(context.Background(), failed)

	assert.Empty(t, recorder.GetQueuedSpans())

	// Check that events were propagated to the original CommandMonitor
	assert.Equal(t, []interface{}{failed}, mon.Events())
}

func TestWrapCommandMonitor_Failed_NotTraced(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder),
	)
	defer instana.ShutdownSensor()
	mon := &monitorMock{}

	m := instamongo.WrapCommandMonitor(mon.Monitor(), sensor)

	started := &event.CommandStartedEvent{
		Command: marshalBSON(t, bson.M{
			"listDatabases": 1,
		}),
		CommandName:  "listDatabases",
		RequestID:    1,
		ConnectionID: "localhost:27017-1",
	}
	m.Started(context.Background(), started)

	failed := &event.CommandFailedEvent{
		CommandFinishedEvent: event.CommandFinishedEvent{
			DurationNanos: 123,
			CommandName:   "listDatabases",
			RequestID:     1,
			ConnectionID:  "localhost:27017-1",
		},
		Failure: "something went wrong",
	}
	m.Failed(context.Background(), failed)

	assert.Empty(t, recorder.GetQueuedSpans())

	// Check that events were propagated to the original CommandMonitor
	assert.Equal(t, []interface{}{started, failed}, mon.Events())
}

// To instrument a mongo.Client created with mongo.Connect() replace mongo.Connect() with instamongo.Connect()
// and pass an instana.Sensor instance to use
func ExampleConnect() {
	// Initialize Instana sensor
	sensor := instana.NewSensor("mongo-client")

	// Replace mongo.Connect() with instamongo.Connect and pass the sensor instance
	client, err := instamongo.Connect(context.Background(), sensor, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	// Query MongoDB as usual using the instrumented client instance
	// ...
}

// To instrument a mongo.Client created with mongo.NewClient() replace mongo.NewClient() with instamongo.NewClient()
// and pass an instana.Sensor instance to use
func ExampleNewClient() {
	// Initialize Instana sensor
	sensor := instana.NewSensor("mongo-client")

	// Replace mongo.Connect() with instamongo.Connect and pass the sensor instance
	client, err := instamongo.NewClient(sensor, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}

	// Use instrumented client as usual
	client.Connect(context.Background())
}

type monitorMock struct {
	mu     sync.RWMutex
	events []interface{}
}

func (m *monitorMock) Started(ctx context.Context, evt *event.CommandStartedEvent) {
	m.recordEvent(evt)
}

func (m *monitorMock) Succeeded(ctx context.Context, evt *event.CommandSucceededEvent) {
	m.recordEvent(evt)
}

func (m *monitorMock) Failed(ctx context.Context, evt *event.CommandFailedEvent) {
	m.recordEvent(evt)
}

func (m *monitorMock) Monitor() *event.CommandMonitor {
	return &event.CommandMonitor{
		Started:   m.Started,
		Succeeded: m.Succeeded,
		Failed:    m.Failed,
	}
}

func (m *monitorMock) Events() []interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	events := make([]interface{}, len(m.events))
	copy(events, m.events)

	return events
}

func (m *monitorMock) recordEvent(evt interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.events = append(m.events, evt)
}

func marshalBSON(t *testing.T, data interface{}) bson.Raw {
	t.Helper()

	doc, err := bson.Marshal(data)
	require.NoError(t, err)

	return bson.Raw(doc)
}

type alwaysReadyClient struct{}

func (alwaysReadyClient) Ready() bool                                       { return true }
func (alwaysReadyClient) SendMetrics(data acceptor.Metrics) error           { return nil }
func (alwaysReadyClient) SendEvent(event *instana.EventData) error          { return nil }
func (alwaysReadyClient) SendSpans(spans []instana.Span) error              { return nil }
func (alwaysReadyClient) SendProfiles(profiles []autoprofile.Profile) error { return nil }
func (alwaysReadyClient) Flush(context.Context) error                       { return nil }
