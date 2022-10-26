// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instana_test

import (
	"testing"

	instana "github.com/instana/go-sensor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNamedMatcher(t *testing.T) {
	examples := map[string]struct {
		List               []string
		MatchingStrings    []string
		NonMatchingStrings []string
	}{
		"equals": {
			List:               []string{"foo", "bar"},
			MatchingStrings:    []string{"foo", "bar"},
			NonMatchingStrings: []string{"Foo", "foobar", "baz"},
		},
		"equals-ignore-case": {
			List:               []string{"foo", "bar"},
			MatchingStrings:    []string{"foo", "bar", "Foo"},
			NonMatchingStrings: []string{"foobar", "baz"},
		},
		"contains": {
			List:               []string{"foo", "bar"},
			MatchingStrings:    []string{"foo", "foobar", "Foobar"},
			NonMatchingStrings: []string{"baz", "FooBar"},
		},
		"contains-ignore-case": {
			List:               []string{"foo", "bar"},
			MatchingStrings:    []string{"foo", "foobar", "Foobar", "FooBar"},
			NonMatchingStrings: []string{"baz"},
		},
		"regex": {
			List:               []string{"(?i)foo.+", "ba[rz]"},
			MatchingStrings:    []string{"foobar", "Foobar", "FooBar", "baz"},
			NonMatchingStrings: []string{"foo", "bap"},
		},
		"none": {
			List:               []string{"foo", "bar"},
			NonMatchingStrings: []string{"foo", "bar", "baz"},
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			m, err := instana.NamedMatcher(name, example.List)
			require.NoError(t, err)

			for _, s := range example.MatchingStrings {
				t.Run(s, func(t *testing.T) {
					assert.True(t, m.Match(s))
				})
			}

			for _, s := range example.NonMatchingStrings {
				t.Run(s, func(t *testing.T) {
					assert.False(t, m.Match(s))
				})
			}
		})
	}
}

func TestNamedMatcher_Unsupported(t *testing.T) {
	_, err := instana.NamedMatcher("custom", []string{"foo", "bar"})
	assert.Error(t, err)
}
