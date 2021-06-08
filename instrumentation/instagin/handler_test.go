// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2016

// +build go1.11

package instagin_test

import (
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/instana/testify/require"

	"github.com/opentracing/opentracing-go"

	"github.com/instana/testify/assert"

	"github.com/gin-gonic/gin"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instagin"
)

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

func TestResponseHeadersCollecting(t *testing.T) {
	engine := getInstrumentedEngine()

	engine.GET("/foo", func(c *gin.Context) {
		c.JSON(200, gin.H{})
	})

	req := httptest.NewRequest("GET", "/foo", nil)
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	assert.NotEmpty(t, w.Header().Get("X-Instana-T"))
	assert.NotEmpty(t, w.Header().Get("X-Instana-S"))
	assert.NotEmpty(t, w.Header().Get("X-Instana-L"))
	assert.NotEmpty(t, w.Header().Get("Traceparent"))
	assert.NotEmpty(t, w.Header().Get("Tracestate"))
}

func TestPropagation(t *testing.T) {
	traceIDHeader := "0000000000001234"
	spanIDHeader := "0000000000004567"

	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(nil, recorder)

	sensor := instana.NewSensorWithTracer(tracer)

	engine := gin.Default()
	instagin.AddMiddleware(sensor, engine)
	engine.GET("/foo", func(c *gin.Context) {

		parent, ok := instana.SpanFromContext(c.Request.Context())
		assert.True(t, ok)

		sp := parent.Tracer().StartSpan("sub-call", opentracing.ChildOf(parent.Context()))
		sp.Finish()

		c.JSON(200, gin.H{})
	})

	req := httptest.NewRequest("GET", "https://example.com/foo", nil)

	req.Header.Add(instana.FieldT, traceIDHeader)
	req.Header.Add(instana.FieldS, spanIDHeader)
	req.Header.Add(instana.FieldL, "1")

	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 2)

	entrySpan, interSpan := spans[1], spans[0]

	assert.EqualValues(t, instana.EntrySpanKind, entrySpan.Kind)
	assert.EqualValues(t, instana.IntermediateSpanKind, interSpan.Kind)

	assert.Equal(t, entrySpan.TraceID, interSpan.TraceID)
	assert.Equal(t, entrySpan.SpanID, interSpan.ParentID)

	assert.Equal(t, traceIDHeader, instana.FormatID(entrySpan.TraceID))
	assert.Equal(t, spanIDHeader, instana.FormatID(entrySpan.ParentID))
}

func getInstrumentedEngine() *gin.Engine {
	sensor := instana.NewSensor("gin-test")
	engine := gin.Default()
	instagin.AddMiddleware(sensor, engine)
	return engine
}
