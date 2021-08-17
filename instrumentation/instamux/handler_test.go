// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2016

// +build go1.12

package instamux_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/instana/testify/require"
	"github.com/opentracing/opentracing-go"

	"github.com/instana/testify/assert"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instamux"

	"github.com/gorilla/mux"
)

func TestMain(m *testing.M) {
	instana.InitSensor(&instana.Options{
		Service: "gorillamux-test",
		Tracer: instana.TracerOptions{
			CollectableHTTPHeaders: []string{"x-custom-header-1", "x-custom-header-2"},
		},
	})

	os.Exit(m.Run())
}

func TestPropagation(t *testing.T) {
	traceIDHeader := "0000000000001234"
	spanIDHeader := "0000000000004567"

	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(nil, recorder)
	sensor := instana.NewSensorWithTracer(tracer)

	r := mux.NewRouter()
	r.HandleFunc("/foo/{id}", func(w http.ResponseWriter, r *http.Request) {
		parent, ok := instana.SpanFromContext(r.Context())
		assert.True(t, ok)

		sp := parent.Tracer().StartSpan("sub-call", opentracing.ChildOf(parent.Context()))
		sp.Finish()
		w.Header().Add("x-custom-header-2", "response")
		w.WriteHeader(http.StatusOK)
	})

	instamux.AddMiddleware(sensor, r)

	req := httptest.NewRequest("GET", "https://example.com/foo/1?SECRET_VALUE=%3Credacted%3E&myPassword=%3Credacted%3E&q=term&sensitive_key=%3Credacted%3E", nil)

	req.Header.Add(instana.FieldT, traceIDHeader)
	req.Header.Add(instana.FieldS, spanIDHeader)
	req.Header.Add(instana.FieldL, "1")
	req.Header.Set("X-Custom-Header-1", "request")

	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

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
		Method:       "GET",
		Status:       http.StatusOK,
		Path:         "/foo/1",
		PathTemplate: "/foo/{id}",
		URL:          "",
		Host:         "example.com",
		Protocol:     "https",
		Params:       "SECRET_VALUE=%3Credacted%3E&myPassword=%3Credacted%3E&q=term&sensitive_key=%3Credacted%3E",
		Headers: map[string]string{
			"x-custom-header-1": "request",
			"x-custom-header-2": "response",
		},
	}, entrySpanData.Tags)
}
