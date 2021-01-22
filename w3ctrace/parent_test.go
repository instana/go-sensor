// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package w3ctrace_test

import (
	"testing"

	"github.com/instana/go-sensor/w3ctrace"
	"github.com/instana/testify/assert"
	"github.com/instana/testify/require"
)

func TestParseParent(t *testing.T) {
	examples := map[string]struct {
		Header   string
		Expected w3ctrace.Parent
	}{
		"v0, valid sampled": {
			Header: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
			Expected: w3ctrace.Parent{
				Version:  w3ctrace.Version_0,
				TraceID:  "4bf92f3577b34da6a3ce929d0e0e4736",
				ParentID: "00f067aa0ba902b7",
				Flags: w3ctrace.Flags{
					Sampled: true,
				},
			},
		},
		"v0, valid not sampled": {
			Header: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-00",
			Expected: w3ctrace.Parent{
				Version:  w3ctrace.Version_0,
				TraceID:  "4bf92f3577b34da6a3ce929d0e0e4736",
				ParentID: "00f067aa0ba902b7",
			},
		},
		"future": {
			Header: "fe-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01-hello future",
			Expected: w3ctrace.Parent{
				Version:  w3ctrace.Version_0,
				TraceID:  "4bf92f3577b34da6a3ce929d0e0e4736",
				ParentID: "00f067aa0ba902b7",
				Flags: w3ctrace.Flags{
					Sampled: true,
				},
			},
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			st, err := w3ctrace.ParseParent(example.Header)
			require.NoError(t, err)
			assert.Equal(t, example.Expected, st)
		})
	}
}

func TestParseParent_Malformed(t *testing.T) {
	examples := map[string]string{
		"invalid version":            "ff-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
		"malformed version":          "xx-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
		"v0, no version separator":   "00@4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
		"v0, no trace id separator":  "00-4bf92f3577b34da6a3ce929d0e0e4736@00f067aa0ba902b7-01",
		"v0, no parent id separator": "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7@01",
		"v0, malformed flags":        "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-xx",
		"future, no flags separator": "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01@hello future",
	}

	for name, header := range examples {
		t.Run(name, func(t *testing.T) {
			_, err := w3ctrace.ParseParent(header)
			assert.Equal(t, w3ctrace.ErrContextCorrupted, err)
		})
	}
}

func TestParent_String(t *testing.T) {
	examples := map[string]struct {
		Parent   w3ctrace.Parent
		Expected string
	}{
		"v0, valid sampled": {
			Parent: w3ctrace.Parent{
				Version:  w3ctrace.Version_0,
				TraceID:  "1234",
				ParentID: "56789",
				Flags: w3ctrace.Flags{
					Sampled: true,
				},
			},
			Expected: "00-00000000000000000000000000001234-0000000000056789-01",
		},
		"v0, valid not sampled": {
			Parent: w3ctrace.Parent{
				Version:  w3ctrace.Version_0,
				TraceID:  "4bf92f3577b34da6a3ce929d0e0e4736",
				ParentID: "00f067aa0ba902b7",
			},
			Expected: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-00",
		},
		"future": {
			Parent: w3ctrace.Parent{
				Version:  w3ctrace.Version(uint8(w3ctrace.Version_Max) + 1),
				TraceID:  "4bf92f3577b34da6a3ce929d0e0e4736",
				ParentID: "00f067aa0ba902b7",
				Flags: w3ctrace.Flags{
					Sampled: true,
				},
			},
			Expected: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, example.Expected, example.Parent.String())
		})
	}
}
