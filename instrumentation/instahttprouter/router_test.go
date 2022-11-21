// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

package instahttprouter_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instahttprouter"
	"github.com/instana/go-sensor/secrets"
	"github.com/instana/go-sensor/w3ctrace"
	"github.com/julienschmidt/httprouter"
	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRouter_Handle_StartTrace(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{
		Tracer: instana.TracerOptions{
			Secrets:                secrets.NewEqualsMatcher("secret"),
			CollectableHTTPHeaders: []string{"X-Custom-1"},
		},
		AgentClient: alwaysReadyClient{},
	}, recorder)
	sensor := instana.NewSensorWithTracer(tracer)

	r := instahttprouter.Wrap(httprouter.New(), sensor)
	r.Handle(http.MethodGet, "/user/:id", func(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
		assert.Equal(t, httprouter.Params{
			httprouter.Param{Key: "id", Value: "id1"},
		}, params)

		w.WriteHeader(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/user/id1?q=test&secret=password", nil)
	req.Header.Set("X-Custom-1", "custom1")
	req.Header.Set("X-Custom-2", "custom2")

	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusNoContent, resp.Code)

	assert.NotEmpty(t, resp.Header().Get(instana.FieldT))
	assert.NotEmpty(t, resp.Header().Get(instana.FieldS))
	assert.NotEmpty(t, resp.Header().Get(instana.FieldL))
	assert.NotEmpty(t, resp.Header().Get(w3ctrace.TraceParentHeader))
	assert.NotEmpty(t, resp.Header().Get(w3ctrace.TraceStateHeader))

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span := spans[0]

	assert.NotEmpty(t, span.TraceID)
	assert.Empty(t, span.ParentID)
	assert.NotEmpty(t, span.SpanID)

	assert.Equal(t, "g.http", span.Name)
	assert.Equal(t, 0, span.Ec)

	require.IsType(t, instana.HTTPSpanData{}, span.Data)
	data := span.Data.(instana.HTTPSpanData)

	assert.Equal(t, instana.HTTPSpanTags{
		Status:       http.StatusNoContent,
		Method:       http.MethodGet,
		Path:         "/user/id1",
		PathTemplate: "/user/:id",
		Host:         "example.com",
		Params:       "q=test&secret=%3Credacted%3E",
		Headers:      map[string]string{"X-Custom-1": "custom1"},
	}, data.Tags)
}

func TestRouter_TracePropagation(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder))

	r := instahttprouter.Wrap(httprouter.New(), sensor)
	r.Handle(http.MethodGet, "/user/:id", func(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
		if sp, ok := instana.SpanFromContext(req.Context()); ok {
			defer sp.Tracer().StartSpan("handler", opentracing.ChildOf(sp.Context())).Finish()
		}

		w.WriteHeader(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/user/id1?q=test&secret=password", nil)

	parentSpan := instana.NewRootSpanContext()
	sensor.Tracer().Inject(parentSpan, opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))

	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusNoContent, resp.Code)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 2)

	handlerSp, routerSp := spans[0], spans[1]

	assert.Equal(t, "g.http", routerSp.Name)
	assert.Equal(t, parentSpan.TraceID, routerSp.TraceID)
	assert.Equal(t, parentSpan.TraceIDHi, routerSp.TraceIDHi)
	assert.Equal(t, parentSpan.SpanID, routerSp.ParentID)

	assert.Equal(t, "sdk", handlerSp.Name)
	assert.Equal(t, routerSp.TraceID, handlerSp.TraceID)
	assert.Equal(t, routerSp.TraceIDHi, handlerSp.TraceIDHi)
	assert.Equal(t, routerSp.SpanID, handlerSp.ParentID)
}

func TestRouter_Handle_ErrorHandling(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder))

	r := instahttprouter.Wrap(httprouter.New(), sensor)
	r.Handle(http.MethodGet, "/user/:id", func(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
		http.Error(w, "something went wrong", http.StatusInternalServerError)
	})

	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, httptest.NewRequest(http.MethodGet, "/user/id1", nil))

	assert.Equal(t, http.StatusInternalServerError, resp.Code)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 2)

	span, logSpan := spans[0], spans[1]

	assert.NotEmpty(t, span.TraceID)
	assert.Empty(t, span.ParentID)
	assert.NotEmpty(t, span.SpanID)

	assert.Equal(t, "g.http", span.Name)
	assert.Equal(t, 1, span.Ec)

	require.IsType(t, instana.HTTPSpanData{}, span.Data)
	data := span.Data.(instana.HTTPSpanData)

	assert.Equal(t, instana.HTTPSpanTags{
		Status:       http.StatusInternalServerError,
		Method:       http.MethodGet,
		Path:         "/user/id1",
		PathTemplate: "/user/:id",
		Host:         "example.com",
		Error:        "Internal Server Error",
	}, data.Tags)

	assert.Equal(t, span.TraceID, logSpan.TraceID)
	assert.Equal(t, span.SpanID, logSpan.ParentID)
	assert.Equal(t, "log.go", logSpan.Name)

	// assert that log message has been recorded within the span interval
	assert.GreaterOrEqual(t, logSpan.Timestamp, span.Timestamp)
	assert.LessOrEqual(t, logSpan.Duration, span.Duration)

	require.IsType(t, instana.LogSpanData{}, logSpan.Data)
	logData := logSpan.Data.(instana.LogSpanData)

	assert.Equal(t, instana.LogSpanTags{
		Level:   "ERROR",
		Message: `error: "Internal Server Error"`,
	}, logData.Tags)
}

func TestRouter_Handle_PanicHandling(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder))

	r := instahttprouter.Wrap(httprouter.New(), sensor)
	r.Handle(http.MethodGet, "/user/:id", func(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
		panic("something went wrong")
	})

	resp := httptest.NewRecorder()
	assert.Panics(t, func() {
		r.ServeHTTP(resp, httptest.NewRequest(http.MethodGet, "/user/id1", nil))
	})

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 2)

	span, logSpan := spans[0], spans[1]

	assert.NotEmpty(t, span.TraceID)
	assert.Empty(t, span.ParentID)
	assert.NotEmpty(t, span.SpanID)

	assert.Equal(t, "g.http", span.Name)
	assert.Equal(t, 1, span.Ec)

	require.IsType(t, instana.HTTPSpanData{}, span.Data)
	data := span.Data.(instana.HTTPSpanData)

	assert.Equal(t, instana.HTTPSpanTags{
		Status:       http.StatusInternalServerError,
		Method:       http.MethodGet,
		Path:         "/user/id1",
		PathTemplate: "/user/:id",
		Host:         "example.com",
		Error:        "something went wrong",
	}, data.Tags)

	assert.Equal(t, span.TraceID, logSpan.TraceID)
	assert.Equal(t, span.SpanID, logSpan.ParentID)
	assert.Equal(t, "log.go", logSpan.Name)

	// assert that log message has been recorded within the span interval
	assert.GreaterOrEqual(t, logSpan.Timestamp, span.Timestamp)
	assert.LessOrEqual(t, logSpan.Duration, span.Duration)

	require.IsType(t, instana.LogSpanData{}, logSpan.Data)
	logData := logSpan.Data.(instana.LogSpanData)

	assert.Equal(t, instana.LogSpanTags{
		Level:   "ERROR",
		Message: `error: "something went wrong"`,
	}, logData.Tags)
}

func TestRouter_Helpers(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder))

	r := instahttprouter.Wrap(httprouter.New(), sensor)

	examples := map[string]func(httprouter.Handle) string{
		"GET": func(h httprouter.Handle) string {
			r.GET("/get", h)

			return "/get"
		},
		"POST": func(h httprouter.Handle) string {
			r.POST("/post", h)

			return "/post"
		},
		"HEAD": func(h httprouter.Handle) string {
			r.HEAD("/head", h)

			return "/head"
		},
		"OPTIONS": func(h httprouter.Handle) string {
			r.OPTIONS("/options", h)

			return "/options"
		},
		"DELETE": func(h httprouter.Handle) string {
			r.DELETE("/delete", h)

			return "/delete"
		},
		"PUT": func(h httprouter.Handle) string {
			r.PUT("/put", h)

			return "/put"
		},
		"PATCH": func(h httprouter.Handle) string {
			r.PATCH("/patch", h)

			return "/patch"
		},
	}

	for method, register := range examples {
		t.Run(method, func(t *testing.T) {
			path := register(func(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
				w.WriteHeader(http.StatusNoContent)
			})

			resp := httptest.NewRecorder()
			r.ServeHTTP(resp, httptest.NewRequest(method, path, nil))

			assert.Equal(t, http.StatusNoContent, resp.Code)

			spans := recorder.GetQueuedSpans()
			require.Len(t, spans, 1)

			span := spans[0]

			assert.NotEmpty(t, span.TraceID)
			assert.Empty(t, span.ParentID)
			assert.NotEmpty(t, span.SpanID)

			assert.Equal(t, "g.http", span.Name)
			assert.Equal(t, 0, span.Ec)

			require.IsType(t, instana.HTTPSpanData{}, span.Data)
			data := span.Data.(instana.HTTPSpanData)

			assert.Equal(t, instana.HTTPSpanTags{
				Status: http.StatusNoContent,
				Method: method,
				Path:   path,
				Host:   "example.com",
			}, data.Tags)
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
