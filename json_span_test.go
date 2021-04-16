// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instana_test

import (
	"testing"

	instana "github.com/instana/go-sensor"
	"github.com/instana/testify/assert"
)

func TestSpanKind_String(t *testing.T) {
	examples := map[string]struct {
		Kind     instana.SpanKind
		Expected string
	}{
		"entry": {
			Kind:     instana.EntrySpanKind,
			Expected: "entry",
		},
		"exit": {
			Kind:     instana.ExitSpanKind,
			Expected: "exit",
		},
		"intermediate": {
			Kind:     instana.IntermediateSpanKind,
			Expected: "intermediate",
		},
		"unknown": {
			Kind:     instana.SpanKind(0),
			Expected: "intermediate",
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, example.Expected, example.Kind.String())
		})
	}
}
