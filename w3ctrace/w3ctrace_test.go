package w3ctrace_test

import (
	"fmt"
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

func TestParseState(t *testing.T) {
	examples := map[string]struct {
		Header   string
		Expected w3ctrace.State
	}{
		"empty": {},
		"single tracing system": {
			Header:   "rojo=00f067aa0ba902b7",
			Expected: w3ctrace.State{"rojo=00f067aa0ba902b7"},
		},
		"multiple tracing systems": {
			Header:   "rojo=00f067aa0ba902b7 , congo=t61rcWkgMzE",
			Expected: w3ctrace.State{"rojo=00f067aa0ba902b7", "congo=t61rcWkgMzE"},
		},
		"with empty list items": {
			Header:   "rojo=00f067aa0ba902b7,    ,,congo=t61rcWkgMzE",
			Expected: w3ctrace.State{"rojo=00f067aa0ba902b7", "congo=t61rcWkgMzE"},
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			st, err := w3ctrace.ParseState(example.Header)
			require.NoError(t, err)
			assert.Equal(t, example.Expected, st)
		})
	}
}

func TestState_Add(t *testing.T) {
	var st w3ctrace.State

	st = st.Add("rojo", "00f067aa0ba902b7")
	require.Equal(t, w3ctrace.State{"rojo=00f067aa0ba902b7"}, st)

	st = st.Add("congo", "t61rcWkgMzE")
	require.Equal(t, w3ctrace.State{"congo=t61rcWkgMzE", "rojo=00f067aa0ba902b7"}, st)

	st = st.Add("", "data")
	require.Equal(t, w3ctrace.State{"congo=t61rcWkgMzE", "rojo=00f067aa0ba902b7"}, st)

	st = st.Add("rojo", "updated")
	require.Equal(t, w3ctrace.State{"rojo=updated", "congo=t61rcWkgMzE"}, st)

	st = st.Add("rojo", "updated again")
	require.Equal(t, w3ctrace.State{"rojo=updated again", "congo=t61rcWkgMzE"}, st)
}

func TestState_Add_MaximumReached(t *testing.T) {
	var st w3ctrace.State

	for i := 0; i < w3ctrace.MaxStateEntries; i++ {
		st = st.Add(fmt.Sprintf("vendor%d", i), "data")
	}

	require.Len(t, st, w3ctrace.MaxStateEntries)
	require.Equal(t, st[w3ctrace.MaxStateEntries-1], "vendor0=data")

	st = st.Add("newVendor", "data")
	require.Len(t, st, w3ctrace.MaxStateEntries)
	assert.Equal(t, st[0], "newVendor=data")
	assert.Equal(t, st[w3ctrace.MaxStateEntries-1], "vendor1=data")
}

func TestState_Remove(t *testing.T) {
	st := w3ctrace.State{"rojo=00f067aa0ba902b7", "congo=t61rcWkgMzE"}

	st = st.Remove("congo")
	require.Equal(t, w3ctrace.State{"rojo=00f067aa0ba902b7"}, st)

	st = st.Remove("")
	require.Equal(t, w3ctrace.State{"rojo=00f067aa0ba902b7"}, st)

	st = st.Remove("vendor")
	require.Equal(t, w3ctrace.State{"rojo=00f067aa0ba902b7"}, st)

	st = st.Remove("rojo")
	require.Empty(t, st)
}

func TestState_String(t *testing.T) {
	examples := map[string]struct {
		State    w3ctrace.State
		Expected string
	}{
		"empty": {},
		"single tracing system": {
			State:    w3ctrace.State{"rojo=00f067aa0ba902b7"},
			Expected: "rojo=00f067aa0ba902b7",
		},
		"multiple tracing systems": {
			State:    w3ctrace.State{"rojo=00f067aa0ba902b7", "congo=t61rcWkgMzE"},
			Expected: "rojo=00f067aa0ba902b7,congo=t61rcWkgMzE",
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, example.Expected, example.State.String())
		})
	}
}
