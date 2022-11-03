// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package w3ctrace_test

import (
	"testing"

	"github.com/instana/go-sensor/w3ctrace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseVersion(t *testing.T) {
	examples := map[string]struct {
		Header   string
		Expected w3ctrace.Version
	}{
		"v0": {
			Header:   "00",
			Expected: w3ctrace.Version_0,
		},
		"vff": {
			Header:   "ff",
			Expected: w3ctrace.Version_Invalid,
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			st, err := w3ctrace.ParseVersion(example.Header)
			require.NoError(t, err)
			assert.Equal(t, example.Expected, st)
		})
	}
}

func TestParseVersion_Malformed(t *testing.T) {
	examples := map[string]struct {
		Header string
	}{
		"empty": {
			Header: "",
		},
		"too short": {
			Header: "f",
		},
		"too long": {
			Header: "abc",
		},
		"non hex": {
			Header: "xy",
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			_, err := w3ctrace.ParseVersion(example.Header)
			assert.Equal(t, w3ctrace.ErrContextCorrupted, err)
		})
	}
}

func TestVersion_String(t *testing.T) {
	examples := map[string]struct {
		Version  w3ctrace.Version
		Expected string
	}{
		"v0": {
			Version:  w3ctrace.Version_0,
			Expected: "00",
		},
		"vff": {
			Version:  w3ctrace.Version_Invalid,
			Expected: "ff",
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, example.Expected, example.Version.String())
		})
	}
}
