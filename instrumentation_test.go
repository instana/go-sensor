package instana_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/w3ctrace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTracingHandlerFunc_Write(t *testing.T) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{
		Service: "go-sensor-test",
	}, recorder))

	h := instana.TracingHandlerFunc(s, "test-handler", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(w, "Ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test?q=classified", nil)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "Ok\n", rec.Body.String())

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	assert.Equal(t, 0, span.Ec)
	assert.EqualValues(t, instana.EntrySpanKind, span.Kind)
	assert.False(t, span.Synthetic)

	assert.Nil(t, span.ForeignParent)

	require.IsType(t, instana.HTTPSpanData{}, span.Data)
	data := span.Data.(instana.HTTPSpanData)

	assert.Equal(t, instana.HTTPSpanTags{
		Host:   "example.com",
		Status: http.StatusOK,
		Method: "GET",
		Path:   "/test",
	}, data.Tags)

	// check whether the trace context has been sent back to the client
	assert.Equal(t, instana.FormatID(span.TraceID), rec.Header().Get(instana.FieldT))
	assert.Equal(t, instana.FormatID(span.SpanID), rec.Header().Get(instana.FieldS))
}

func TestTracingHandlerFunc_WriteHeaders(t *testing.T) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	h := instana.TracingHandlerFunc(s, "test-handler", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/test?q=classified", nil))

	assert.Equal(t, http.StatusNotFound, rec.Code)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	assert.Equal(t, 0, span.Ec)
	assert.EqualValues(t, instana.EntrySpanKind, span.Kind)
	assert.False(t, span.Synthetic)

	require.IsType(t, instana.HTTPSpanData{}, span.Data)
	data := span.Data.(instana.HTTPSpanData)

	assert.Equal(t, instana.HTTPSpanTags{
		Status: http.StatusNotFound,
		Method: "GET",
		Host:   "example.com",
		Path:   "/test",
	}, data.Tags)
}

func TestTracingHandlerFunc_Error(t *testing.T) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	h := instana.TracingHandlerFunc(s, "test-handler", func(w http.ResponseWriter, req *http.Request) {
		http.Error(w, "something went wrong", http.StatusInternalServerError)
	})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/test", nil))

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	assert.Equal(t, 0, span.Ec)
	assert.EqualValues(t, instana.EntrySpanKind, span.Kind)
	assert.False(t, span.Synthetic)

	require.IsType(t, instana.HTTPSpanData{}, span.Data)
	data := span.Data.(instana.HTTPSpanData)

	assert.Equal(t, instana.HTTPSpanTags{
		Status: http.StatusInternalServerError,
		Method: "GET",
		Host:   "example.com",
		Path:   "/test",
	}, data.Tags)
}

func TestTracingHandlerFunc_W3CTraceContext_ForeignParent(t *testing.T) {
	examples := map[string]struct {
		IncomingHeaders map[string]string
		Expected        *instana.ForeignParent
	}{
		"incoming w3c, no instana headers, latest state from instana": {
			IncomingHeaders: map[string]string{
				w3ctrace.TraceParentHeader: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
				w3ctrace.TraceStateHeader:  "in=abc123;def456,vendorname1=opaqueValue1",
			},
			Expected: &instana.ForeignParent{
				TraceID:          "4bf92f3577b34da6a3ce929d0e0e4736",
				ParentID:         "00f067aa0ba902b7",
				LatestTraceState: "in=abc123;def456",
			},
		},
		"incoming w3c, with instana headers, latest state from instana": {
			IncomingHeaders: map[string]string{
				w3ctrace.TraceParentHeader: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
				w3ctrace.TraceStateHeader:  "in=abc123;def456,vendorname1=opaqueValue1",
				instana.FieldT:             "abc123",
				instana.FieldS:             "def456",
			},
		},
		"incoming w3c, with instana headers, latest state not from instana": {
			IncomingHeaders: map[string]string{
				w3ctrace.TraceParentHeader: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
				w3ctrace.TraceStateHeader:  "vendorname1=opaqueValue1,in=abc123;def456",
				instana.FieldT:             "abc123",
				instana.FieldS:             "def456",
			},
			Expected: &instana.ForeignParent{
				TraceID:          "4bf92f3577b34da6a3ce929d0e0e4736",
				ParentID:         "00f067aa0ba902b7",
				LatestTraceState: "vendorname1=opaqueValue1",
			},
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			recorder := instana.NewTestRecorder()
			s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{
				Service: "go-sensor-test",
			}, recorder))

			h := instana.TracingHandlerFunc(s, "test-handler", func(w http.ResponseWriter, req *http.Request) {
				fmt.Fprintln(w, "Ok")
			})

			rec := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			for k, v := range example.IncomingHeaders {
				req.Header.Set(k, v)
			}

			h.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "Ok\n", rec.Body.String())

			spans := recorder.GetQueuedSpans()
			require.Len(t, spans, 1)

			assert.Equal(t, example.Expected, spans[0].ForeignParent)
		})
	}
}

func TestTracingHandlerFunc_SyntheticCall(t *testing.T) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	h := instana.TracingHandlerFunc(s, "test-handler", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(w, "Ok")
	})

	rec := httptest.NewRecorder()

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(instana.FieldSynthetic, "1")

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)
	assert.True(t, spans[0].Synthetic)
}

func TestTracingHandlerFunc_PanicHandling(t *testing.T) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	h := instana.TracingHandlerFunc(s, "test-handler", func(w http.ResponseWriter, req *http.Request) {
		panic("something went wrong")
	})

	rec := httptest.NewRecorder()
	assert.Panics(t, func() {
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/test?q=classified", nil))
	})

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	assert.Equal(t, 1, span.Ec)
	assert.EqualValues(t, instana.EntrySpanKind, span.Kind)
	assert.False(t, span.Synthetic)

	require.IsType(t, instana.HTTPSpanData{}, span.Data)
	data := span.Data.(instana.HTTPSpanData)

	assert.Equal(t, instana.HTTPSpanTags{
		Status: http.StatusInternalServerError,
		Method: "GET",
		Host:   "example.com",
		Path:   "/test",
		Error:  "something went wrong",
	}, data.Tags)
}

func TestRoundTripper(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)
	s := instana.NewSensorWithTracer(tracer)

	parentSpan := tracer.StartSpan("parent")

	var traceIDHeader, spanIDHeader string
	rt := instana.RoundTripper(s, testRoundTripper(func(req *http.Request) (*http.Response, error) {
		traceIDHeader = req.Header.Get(instana.FieldT)
		spanIDHeader = req.Header.Get(instana.FieldS)

		return &http.Response{
			Status:     http.StatusText(http.StatusNotImplemented),
			StatusCode: http.StatusNotImplemented,
		}, nil
	}))

	ctx := instana.ContextWithSpan(context.Background(), parentSpan)
	req := httptest.NewRequest("GET", "http://user:password@example.com/hello", nil)

	_, err := rt.RoundTrip(req.WithContext(ctx))
	require.NoError(t, err)

	parentSpan.Finish()

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 2)

	cSpan, pSpan := spans[0], spans[1]
	assert.Equal(t, 0, cSpan.Ec)
	assert.EqualValues(t, instana.ExitSpanKind, cSpan.Kind)

	assert.Equal(t, pSpan.TraceID, cSpan.TraceID)
	assert.Equal(t, pSpan.SpanID, cSpan.ParentID)

	assert.Equal(t, instana.FormatID(cSpan.TraceID), traceIDHeader)
	assert.Equal(t, instana.FormatID(cSpan.SpanID), spanIDHeader)

	require.IsType(t, instana.HTTPSpanData{}, cSpan.Data)
	data := cSpan.Data.(instana.HTTPSpanData)

	assert.Equal(t, instana.HTTPSpanTags{
		Method: "GET",
		Status: http.StatusNotImplemented,
		URL:    "http://example.com/hello",
	}, data.Tags)
}

func TestRoundTripper_WithoutParentSpan(t *testing.T) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	rt := instana.RoundTripper(s, testRoundTripper(func(req *http.Request) (*http.Response, error) {
		assert.Empty(t, req.Header.Get(instana.FieldT))
		assert.Empty(t, req.Header.Get(instana.FieldS))

		return &http.Response{
			Status:     http.StatusText(http.StatusNotImplemented),
			StatusCode: http.StatusNotImplemented,
		}, nil
	}))

	resp, err := rt.RoundTrip(httptest.NewRequest("GET", "http://example.com/hello", nil))
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotImplemented, resp.StatusCode)

	assert.Empty(t, recorder.GetQueuedSpans())
}

func TestRoundTripper_Error(t *testing.T) {
	serverErr := errors.New("something went wrong")

	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	rt := instana.RoundTripper(s, testRoundTripper(func(req *http.Request) (*http.Response, error) {
		return nil, serverErr
	}))

	ctx := instana.ContextWithSpan(context.Background(), s.Tracer().StartSpan("parent"))
	req := httptest.NewRequest("GET", "http://example.com/hello", nil)

	_, err := rt.RoundTrip(req.WithContext(ctx))
	assert.Error(t, err)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	assert.Equal(t, 1, span.Ec)
	assert.EqualValues(t, instana.ExitSpanKind, span.Kind)

	require.IsType(t, instana.HTTPSpanData{}, span.Data)
	data := span.Data.(instana.HTTPSpanData)

	assert.Equal(t, instana.HTTPSpanTags{
		Method: "GET",
		URL:    "http://example.com/hello",
		Error:  "something went wrong",
	}, data.Tags)
}

func TestRoundTripper_DefaultTransport(t *testing.T) {
	recorder := instana.NewTestRecorder()
	s := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	var numCalls int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		numCalls++

		assert.NotEmpty(t, req.Header.Get(instana.FieldT))
		assert.NotEmpty(t, req.Header.Get(instana.FieldS))

		w.Write([]byte("OK"))
	}))
	defer ts.Close()

	rt := instana.RoundTripper(s, nil)

	ctx := instana.ContextWithSpan(context.Background(), s.Tracer().StartSpan("parent"))
	req := httptest.NewRequest("GET", ts.URL+"/hello", nil)

	resp, err := rt.RoundTrip(req.WithContext(ctx))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	assert.Equal(t, 1, numCalls)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	assert.Equal(t, 0, span.Ec)
	assert.EqualValues(t, instana.ExitSpanKind, span.Kind)

	require.IsType(t, instana.HTTPSpanData{}, span.Data)
	data := span.Data.(instana.HTTPSpanData)

	assert.Equal(t, instana.HTTPSpanTags{
		Status: http.StatusOK,
		Method: "GET",
		URL:    ts.URL + "/hello",
	}, data.Tags)
}

type testRoundTripper func(*http.Request) (*http.Response, error)

func (rt testRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return rt(req)
}
