// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package w3ctrace_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/instana/go-sensor/w3ctrace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTracingHandlerFunc(t *testing.T) {
	h := w3ctrace.TracingHandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("X-Test", "value")
		w.Write([]byte("Ok"))
	})

	resp := httptest.NewRecorder()

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("traceparent", exampleTraceParent)
	req.Header.Set("tracestate", exampleTraceState)

	h(resp, req)
	require.Equal(t, http.StatusOK, resp.Result().StatusCode)
	assert.Equal(t, "Ok", resp.Body.String())

	assert.Equal(t, "value", resp.Header().Get("X-Test"))
	assert.Equal(t, exampleTraceParent, resp.Header().Get(w3ctrace.TraceParentHeader))
	assert.Equal(t, exampleTraceState, resp.Header().Get(w3ctrace.TraceStateHeader))
}

func TestTracingHandlerFunc_NoContext(t *testing.T) {
	h := w3ctrace.TracingHandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("X-Test", "value")
		w.Write([]byte("Ok"))
	})

	resp := httptest.NewRecorder()

	req := httptest.NewRequest("GET", "/", nil)

	h(resp, req)
	require.Equal(t, http.StatusOK, resp.Result().StatusCode)
	assert.Equal(t, "Ok", resp.Body.String())

	assert.Equal(t, "value", resp.Header().Get("X-Test"))
	assert.Empty(t, resp.Header().Get(w3ctrace.TraceParentHeader))
	assert.Empty(t, resp.Header().Get(w3ctrace.TraceStateHeader))
}
