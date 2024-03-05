// (c) Copyright IBM Corp. 2023

//go:build go1.18
// +build go1.18

package instabeego_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	beego "github.com/beego/beego/v2/server/web"
	beecontext "github.com/beego/beego/v2/server/web/context"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"
	"github.com/instana/go-sensor/instrumentation/instabeego"
	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ListJson defines the response for the test application
type ListJson struct {
	Value string `json:"Value"`
}

// shutdownBeeApp close the beego server for testing
func shutdownBeeApp() {
	beego.BeeApp.Server.Shutdown(context.TODO())
}

func sleep(t time.Duration) {
	time.Sleep(t)
}

// initBeeApp deploy a beego server for testing
func initBeeApp(t *testing.T) {

	var serverDepTime time.Duration = 2 * time.Second

	beego.Get("/foo", func(ctx *beecontext.Context) {
		listJson := ListJson{
			Value: "abcd",
		}

		parent, ok := instana.SpanFromContext(ctx.Request.Context())
		assert.True(t, ok)

		sp := parent.Tracer().StartSpan("sub-call", opentracing.ChildOf(parent.Context()))
		sp.Finish()

		ctx.Output.SetStatus(http.StatusOK)
		ctx.Output.Header("x-custom-header-2", "response")
		ctx.JSONResp(listJson)
	})

	beego.BConfig.RunMode = "test"
	go beego.Run()
	sleep(serverDepTime)
}

func initSensor() {
	instana.InitCollector(&instana.Options{
		Service: "beego-test",
		Tracer: instana.TracerOptions{
			CollectableHTTPHeaders: []string{"x-custom-header-1", "x-custom-header-2"},
		},
		AgentClient: alwaysReadyClient{},
	})
}

func TestPropagation(t *testing.T) {
	initSensor()
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{
		Service:     "beego-test",
		AgentClient: alwaysReadyClient{},
	}, recorder)
	defer instana.ShutdownSensor()
	sensor := instana.NewSensorWithTracer(tracer)

	instabeego.InstrumentWebServer(sensor)

	defer shutdownBeeApp()

	initBeeApp(t)

	traceIDHeader := "0000000000001234"
	spanIDHeader := "0000000000004567"

	req := httptest.NewRequest("GET", "https://example.com/foo?SECRET_VALUE=%3Credacted%3E&myPassword=%3Credacted%3E&q=term&sensitive_key=%3Credacted%3E", nil)
	req.Header.Add(instana.FieldT, traceIDHeader)
	req.Header.Add(instana.FieldS, spanIDHeader)
	req.Header.Add(instana.FieldL, "1")
	req.Header.Set("X-Custom-Header-1", "request")

	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, req)

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

type alwaysReadyClient struct{}

func (alwaysReadyClient) Ready() bool                                       { return true }
func (alwaysReadyClient) SendMetrics(data acceptor.Metrics) error           { return nil }
func (alwaysReadyClient) SendEvent(event *instana.EventData) error          { return nil }
func (alwaysReadyClient) SendSpans(spans []instana.Span) error              { return nil }
func (alwaysReadyClient) SendProfiles(profiles []autoprofile.Profile) error { return nil }
func (alwaysReadyClient) Flush(context.Context) error                       { return nil }
