// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2016

package instagin_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"

	"github.com/stretchr/testify/require"

	"github.com/opentracing/opentracing-go"

	"github.com/stretchr/testify/assert"

	"github.com/gin-gonic/gin"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instagin"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	gin.DefaultWriter = ioutil.Discard

	instana.InitSensor(&instana.Options{
		Service: "gin-test",
		Tracer: instana.TracerOptions{
			CollectableHTTPHeaders: []string{"x-custom-header-1", "x-custom-header-2"},
		},
		AgentClient: alwaysReadyClient{},
	})

	os.Exit(m.Run())
}

func TestAddMiddleware(t *testing.T) {
	const expectedHandlersAmount = 3

	// create a gin engine with default handlers
	engine := gin.Default()

	handlerN1Pointer := reflect.ValueOf(engine.Handlers[0]).Pointer()
	handlerN2Pointer := reflect.ValueOf(engine.Handlers[1]).Pointer()
	assert.Len(t, engine.Handlers, expectedHandlersAmount-1)

	// create a gin engine with default handlers and add middleware
	engine = getInstrumentedEngine()

	assert.Len(t, engine.Handlers, expectedHandlersAmount)

	// check that middleware was added as a first handler
	assert.NotEqual(t, handlerN1Pointer, reflect.ValueOf(engine.Handlers[0]).Pointer())
	assert.Equal(t, handlerN1Pointer, reflect.ValueOf(engine.Handlers[1]).Pointer())
	assert.Equal(t, handlerN2Pointer, reflect.ValueOf(engine.Handlers[2]).Pointer())
}

func TestPropagation(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(nil, recorder)
	defer instana.ShutdownSensor()
	sensor := instana.NewSensorWithTracer(tracer)

	engines := map[string]func() *gin.Engine{
		"AddMiddleware": func() *gin.Engine {
			engine := gin.Default()
			instagin.AddMiddleware(sensor, engine)
			return engine
		},
		"New": func() *gin.Engine {
			return instagin.New(sensor)
		},
		"Default": func() *gin.Engine {
			return instagin.Default(sensor)
		},
	}

	traceIDHeader := "0000000000001234"
	spanIDHeader := "0000000000004567"

	for _, getEngine := range engines {
		engine := getEngine()
		engine.GET("/foo", func(c *gin.Context) {

			parent, ok := instana.SpanFromContext(c.Request.Context())
			assert.True(t, ok)

			sp := parent.Tracer().StartSpan("sub-call", opentracing.ChildOf(parent.Context()))
			sp.Finish()

			c.Header("x-custom-header-2", "response")
			c.JSON(200, gin.H{})
		})

		req := httptest.NewRequest("GET", "https://example.com/foo?SECRET_VALUE=%3Credacted%3E&myPassword=%3Credacted%3E&q=term&sensitive_key=%3Credacted%3E", nil)

		req.Header.Add(instana.FieldT, traceIDHeader)
		req.Header.Add(instana.FieldS, spanIDHeader)
		req.Header.Add(instana.FieldL, "1")
		req.Header.Set("X-Custom-Header-1", "request")

		w := httptest.NewRecorder()

		engine.ServeHTTP(w, req)

		// Response headers assertions
		assert.NotEmpty(t, w.Header().Get("X-Instana-T"))
		assert.NotEmpty(t, w.Header().Get("X-Instana-S"))
		assert.NotEmpty(t, w.Header().Get("X-Instana-L"))
		assert.NotEmpty(t, w.Header().Get("Traceparent"))
		assert.NotEmpty(t, w.Header().Get("Tracestate"))

		spans := recorder.GetQueuedSpans()
		require.Len(t, spans, 2)

		entrySpan, interSpan := spans[1], spans[0]

		assert.EqualValues(t, instana.EntrySpanKind, entrySpan.Kind)
		assert.EqualValues(t, instana.IntermediateSpanKind, interSpan.Kind)

		assert.Equal(t, entrySpan.TraceID, interSpan.TraceID)
		assert.Equal(t, entrySpan.SpanID, interSpan.ParentID)

		assert.Equal(t, traceIDHeader, instana.FormatID(entrySpan.TraceID))
		assert.Equal(t, spanIDHeader, instana.FormatID(entrySpan.ParentID))

		// ensure that entry span contains all necessary data
		require.IsType(t, instana.HTTPSpanData{}, entrySpan.Data)
		entrySpanData := entrySpan.Data.(instana.HTTPSpanData)

		assert.Equal(t, instana.HTTPSpanTags{
			Method:   "GET",
			Status:   http.StatusOK,
			Path:     "/foo",
			URL:      "",
			Host:     "example.com",
			Protocol: "https",
			Params:   "SECRET_VALUE=%3Credacted%3E&myPassword=%3Credacted%3E&q=term&sensitive_key=%3Credacted%3E",
			Headers: map[string]string{
				"x-custom-header-1": "request",
				"x-custom-header-2": "response",
			},
		}, entrySpanData.Tags)
	}
}

func getInstrumentedEngine() *gin.Engine {
	sensor := instana.NewSensor("gin-test")
	engine := gin.Default()
	instagin.AddMiddleware(sensor, engine)
	return engine
}

type alwaysReadyClient struct{}

func (alwaysReadyClient) Ready() bool                                       { return true }
func (alwaysReadyClient) SendMetrics(data acceptor.Metrics) error           { return nil }
func (alwaysReadyClient) SendEvent(event *instana.EventData) error          { return nil }
func (alwaysReadyClient) SendSpans(spans []instana.Span) error              { return nil }
func (alwaysReadyClient) SendProfiles(profiles []autoprofile.Profile) error { return nil }
func (alwaysReadyClient) Flush(context.Context) error                       { return nil }
