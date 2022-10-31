// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package w3ctrace_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/instana/go-sensor/w3ctrace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const maxKVPairs = 32

func TestParseState(t *testing.T) {

	examples := map[string]struct {
		Header   string
		Expected w3ctrace.State
	}{
		"empty": {
			Expected: w3ctrace.NewState([]string{}, ""),
		},
		"single tracing system": {
			Header:   "rojo=00f067aa0ba902b7",
			Expected: w3ctrace.NewState([]string{"rojo=00f067aa0ba902b7"}, ""),
		},
		"multiple tracing systems": {
			Header:   "rojo=00f067aa0ba902b7,congo=t61rcWkgMzE",
			Expected: w3ctrace.NewState([]string{"rojo=00f067aa0ba902b7", "congo=t61rcWkgMzE"}, ""),
		},
		"only an Instana list member": {
			Header:   "in=fa2375d711a4ca0f;02468acefdb97531",
			Expected: w3ctrace.NewState([]string{}, "fa2375d711a4ca0f;02468acefdb97531"),
		},
		"one other and one Instana list member": {
			Header:   "rojo=00f067aa0ba902b7,in=fa2375d711a4ca0f;02468acefdb97531",
			Expected: w3ctrace.NewState([]string{"rojo=00f067aa0ba902b7"}, "fa2375d711a4ca0f;02468acefdb97531"),
		},
		"one Instana and one other list member": {
			Header:   "in=fa2375d711a4ca0f;02468acefdb97531,rojo=00f067aa0ba902b7",
			Expected: w3ctrace.NewState([]string{"rojo=00f067aa0ba902b7"}, "fa2375d711a4ca0f;02468acefdb97531"),
		},
		"one Instana list member between two others": {
			Header:   "rojo=00f067aa0ba902b7,in=fa2375d711a4ca0f;02468acefdb97531,congo=t61rcWkgMzE",
			Expected: w3ctrace.NewState([]string{"rojo=00f067aa0ba902b7", "congo=t61rcWkgMzE"}, "fa2375d711a4ca0f;02468acefdb97531"),
		},
		"with empty list items": {
			Header:   "rojo=00f067aa0ba902b7,    ,,congo=t61rcWkgMzE",
			Expected: w3ctrace.NewState([]string{"rojo=00f067aa0ba902b7", "    ", "congo=t61rcWkgMzE"}, ""),
		},
		"with whitespace around the Instana list member": {
			Header:   "rojo=00f067aa0ba902b7,  in   =   fa2375d711a4ca0f;02468acefdb97531    ,congo=t61rcWkgMzE",
			Expected: w3ctrace.NewState([]string{"rojo=00f067aa0ba902b7", "congo=t61rcWkgMzE"}, "fa2375d711a4ca0f;02468acefdb97531"),
		},
		"with 33 list items": {
			Header:   strings.TrimRight(strings.Repeat("rojo=00f067aa0ba902b7,", maxKVPairs+1), ","),
			Expected: w3ctrace.NewState(strings.Split(strings.TrimRight(strings.Repeat("rojo=00f067aa0ba902b7,", maxKVPairs), ","), ","), "")},
		"with too many list members and an Instana list member at the end": {
			Header:   strings.TrimRight(strings.Repeat("rojo=00f067aa0ba902b7,", maxKVPairs+1), ",") + ",in=fa2375d711a4ca0f;02468acefdb97531",
			Expected: w3ctrace.NewState(strings.Split(strings.TrimRight(strings.Repeat("rojo=00f067aa0ba902b7,", maxKVPairs-1), ","), ","), "fa2375d711a4ca0f;02468acefdb97531"),
		},
		"with 34 list items, with long one at the beginning": {
			Header: "rojo=" + strings.Repeat("a", 129) + "," + strings.TrimRight(strings.Repeat("rojo=00f067aa0ba902b7,", maxKVPairs+1), ","),
			Expected: w3ctrace.NewState(
				strings.Split(strings.TrimRight(strings.Repeat("rojo=00f067aa0ba902b7,", maxKVPairs), ","), ","),
				"",
			),
		},
		"with 33 list items, each is more then 128 char long": {
			Header: strings.TrimRight(strings.Repeat("rojo="+strings.Repeat("a", 129)+",", maxKVPairs+1), ","),
			Expected: w3ctrace.NewState(
				strings.Split(strings.TrimRight(strings.Repeat("rojo="+strings.Repeat("a", 129)+",", maxKVPairs), ","), ","),
				"",
			),
		},
		"with 34 list items: one short and 33 long": {
			Header: "rojo=00f067aa0ba902b7," + strings.TrimRight(strings.Repeat("rojo="+strings.Repeat("a", 129)+",", maxKVPairs+1), ","),
			Expected: w3ctrace.NewState(
				strings.Split("rojo=00f067aa0ba902b7,"+strings.TrimRight(strings.Repeat("rojo="+strings.Repeat("a", 129)+",", maxKVPairs-1), ","), ","),
				"",
			),
		},
		"with 64 list items, mixed long and short values": {
			Header: strings.TrimRight(strings.Repeat("short="+strings.Repeat("b", 10)+","+"long="+strings.Repeat("a", 129)+",", maxKVPairs), ","),
			Expected: w3ctrace.NewState(
				strings.Split(strings.TrimRight(strings.Repeat("short="+strings.Repeat("b", 10)+",", maxKVPairs), ","), ","),
				"",
			),
		},
		"with empty header value": {
			Header:   "",
			Expected: w3ctrace.NewState([]string{}, ""),
		},
		"with a lot of comas": {
			Header:   strings.Repeat(",", 1024),
			Expected: w3ctrace.NewState([]string{}, ""),
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			st := w3ctrace.ParseState(example.Header)
			assert.Equal(t, example.Expected, st)
		})
	}
}

func TestState_FormStateWithInstanaTraceStateValueIntoEmptyTraceState(t *testing.T) {
	st := w3ctrace.NewState([]string{}, "")
	st = w3ctrace.FormStateWithInstanaTraceStateValue(st, "fa2375d711a4ca0f;02468acefdb97531")
	require.Equal(t, w3ctrace.NewState([]string{}, "fa2375d711a4ca0f;02468acefdb97531"), st)
}

func TestState_FormStateWithInstanaTraceStateValueIntoNonEmptyTraceState(t *testing.T) {
	st := w3ctrace.NewState([]string{"key1=value1", "key2=value"}, "")
	st = w3ctrace.FormStateWithInstanaTraceStateValue(st, "fa2375d711a4ca0f;02468acefdb97531")
	require.Equal(t, w3ctrace.NewState([]string{"key1=value1", "key2=value"}, "fa2375d711a4ca0f;02468acefdb97531"), st)
}

func TestState_FormStateWithInstanaTraceStateValueOverwriteExistingValue(t *testing.T) {
	st := w3ctrace.NewState([]string{}, "fa2375d711a4ca0f;02468acefdb97531")
	st = w3ctrace.FormStateWithInstanaTraceStateValue(st, "aaabbccddeeff012;123456789abcdef0")
	require.Equal(t, w3ctrace.NewState([]string{}, "aaabbccddeeff012;123456789abcdef0"), st)
}

func TestState_FormStateWithInstanaTraceStateValueResetExistingValue(t *testing.T) {
	st := w3ctrace.NewState([]string{}, "fa2375d711a4ca0f;02468acefdb97531")
	st = w3ctrace.FormStateWithInstanaTraceStateValue(st, "")
	require.Equal(t, w3ctrace.NewState([]string{}, ""), st)
}

func TestState_FormStateWithInstanaTraceStateValueAddToTraceStateWithMaxNumberOfListMembers(t *testing.T) {
	var listMembers []string
	for i := 0; i < maxKVPairs; i++ {
		listMembers = append(listMembers, fmt.Sprintf("key%d=value%d", i, i))
	}

	// initially, we are just under the allowed number of list members
	st := w3ctrace.NewState(listMembers, "")
	require.Equal(t, w3ctrace.NewState(listMembers, ""), st)
	// now we also add an Instana list member, which brings us over the limit
	st = w3ctrace.FormStateWithInstanaTraceStateValue(st, "fa2375d711a4ca0f;02468acefdb97531")
	// so we expect the right-most list member to be dropped
	require.Equal(t, w3ctrace.NewState(listMembers[:maxKVPairs-1], "fa2375d711a4ca0f;02468acefdb97531"), st)
}

func TestState_FetchInstanaTraceStateValueNotPresent(t *testing.T) {
	st := w3ctrace.NewState([]string{}, "")
	instanaTraceStateValue, ok := st.FetchInstanaTraceStateValue()
	require.False(t, ok)
	require.Equal(t, "", instanaTraceStateValue)
}

func TestState_FetchInstanaTraceStateValuePresent(t *testing.T) {
	st := w3ctrace.NewState([]string{}, "fa2375d711a4ca0f;02468acefdb97531")
	instanaTraceStateValue, ok := st.FetchInstanaTraceStateValue()
	require.True(t, ok)
	require.Equal(t, "fa2375d711a4ca0f;02468acefdb97531", instanaTraceStateValue)
}

func TestState_String(t *testing.T) {
	examples := map[string]struct {
		State    w3ctrace.State
		Expected string
	}{
		"empty": {},
		"single tracing system": {
			State:    w3ctrace.NewState([]string{"rojo=00f067aa0ba902b7"}, ""),
			Expected: "rojo=00f067aa0ba902b7",
		},
		"only an Instana list member": {
			State:    w3ctrace.NewState([]string{}, "fa2375d711a4ca0f;02468acefdb97531"),
			Expected: "in=fa2375d711a4ca0f;02468acefdb97531",
		},
		"multiple tracing systems, without Instana list member": {
			State:    w3ctrace.NewState([]string{"rojo=00f067aa0ba902b7", "congo=t61rcWkgMzE"}, ""),
			Expected: "rojo=00f067aa0ba902b7,congo=t61rcWkgMzE",
		},
		"multiple tracing systems plus an Instana list member": {
			State:    w3ctrace.NewState([]string{"rojo=00f067aa0ba902b7", "congo=t61rcWkgMzE"}, "fa2375d711a4ca0f;02468acefdb97531"),
			Expected: "in=fa2375d711a4ca0f;02468acefdb97531,rojo=00f067aa0ba902b7,congo=t61rcWkgMzE",
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, example.Expected, example.State.String())
		})
	}
}
