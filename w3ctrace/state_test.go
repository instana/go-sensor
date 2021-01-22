// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package w3ctrace_test

import (
	"fmt"
	"testing"

	"github.com/instana/go-sensor/w3ctrace"
	"github.com/instana/testify/assert"
	"github.com/instana/testify/require"
)

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

func TestState_Fetch(t *testing.T) {
	st := w3ctrace.State{"rojo=00f067aa0ba902b7", "congo=t61rcWkgMzE"}

	t.Run("existing", func(t *testing.T) {
		if vd, ok := st.Fetch("rojo"); assert.True(t, ok) {
			assert.Equal(t, "00f067aa0ba902b7", vd)
		}

		if vd, ok := st.Fetch("congo"); assert.True(t, ok) {
			assert.Equal(t, "t61rcWkgMzE", vd)
		}
	})

	t.Run("non-existing", func(t *testing.T) {
		_, ok := st.Fetch("vendor")
		assert.False(t, ok)
	})
}

func TestState_Index(t *testing.T) {
	st := w3ctrace.State{"rojo=00f067aa0ba902b7", "congo=t61rcWkgMzE"}

	t.Run("existing", func(t *testing.T) {
		assert.Equal(t, 0, st.Index("rojo"))
		assert.Equal(t, 1, st.Index("congo"))
	})

	t.Run("non-existing", func(t *testing.T) {
		assert.Equal(t, -1, st.Index("vendor"))
	})
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
