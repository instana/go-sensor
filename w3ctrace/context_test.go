package w3ctrace_test

import (
	"net/http"
	"testing"

	"github.com/instana/go-sensor/w3ctrace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	exampleTraceParent = "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01"
	exampleTraceState  = "vendorname1=opaqueValue1 , vendorname2=opaqueValue2"
)

func TestExtract(t *testing.T) {
	examples := map[string]struct {
		ParentHeader string
		StateHeader  string
	}{
		"lower case": {"traceparent", "tracestate"},
		"upper case": {"TRACEPARENT", "TRACESTATE"},
		"mixed case": {"Traceparent", "Tracestate"},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			headers := http.Header{}
			// set raw headers to preserve header name case
			headers[example.ParentHeader] = []string{exampleTraceParent}
			headers[example.StateHeader] = []string{exampleTraceState}

			tr, err := w3ctrace.Extract(headers)
			require.NoError(t, err)

			assert.Equal(t, w3ctrace.Context{
				RawParent: exampleTraceParent,
				RawState:  exampleTraceState,
			}, tr)
		})
	}
}

func TestExtract_NoContext(t *testing.T) {
	headers := http.Header{}
	headers.Set(w3ctrace.TraceStateHeader, exampleTraceState)

	_, err := w3ctrace.Extract(headers)
	assert.Equal(t, w3ctrace.ErrContextNotFound, err)
}

func TestInject(t *testing.T) {
	examples := map[string]http.Header{
		"add": {
			"Authorization": []string{"Basic 123"},
		},
		"update": {
			"Authorization": []string{"Basic 123"},
			"traceparent":   []string{"00-abcdef1-01"},
			"TraceState":    []string{"x=y"},
		},
	}

	for name, headers := range examples {
		t.Run(name, func(t *testing.T) {
			w3ctrace.Inject(w3ctrace.Context{
				RawParent: exampleTraceParent,
				RawState:  exampleTraceState,
			}, headers)

			assert.Equal(t, "Basic 123", headers.Get("Authorization"))
			assert.Equal(t, exampleTraceParent, headers.Get(w3ctrace.TraceParentHeader))
			assert.Equal(t, exampleTraceState, headers.Get(w3ctrace.TraceStateHeader))
		})
	}
}

func TestContext_State(t *testing.T) {
	trCtx := w3ctrace.Context{
		RawParent: exampleTraceParent,
		RawState:  exampleTraceState,
	}

	assert.Equal(t, w3ctrace.State{"vendorname1=opaqueValue1", "vendorname2=opaqueValue2"}, trCtx.State())
}

func TestContext_Parent(t *testing.T) {
	trCtx := w3ctrace.Context{
		RawParent: exampleTraceParent,
		RawState:  exampleTraceState,
	}

	assert.Equal(t, w3ctrace.Parent{
		Version:  w3ctrace.Version_0,
		TraceID:  "0af7651916cd43dd8448eb211c80319c",
		ParentID: "b7ad6b7169203331",
		Flags: w3ctrace.Flags{
			Sampled: true,
		},
	}, trCtx.Parent())
}
