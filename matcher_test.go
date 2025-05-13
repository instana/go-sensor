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

func TestDefaultSecretsMatcher(t *testing.T) {
	m := instana.DefaultSecretsMatcher()

	// Test default matcher - match
	assert.True(t, m.Match("key"))
	assert.True(t, m.Match("pass"))
	assert.True(t, m.Match("secret"))

	assert.True(t, m.Match("KEY"))
	assert.True(t, m.Match("PASS"))
	assert.True(t, m.Match("SECRET"))

	assert.True(t, m.Match("key123"))
	assert.True(t, m.Match("pass123"))
	assert.True(t, m.Match("secret123"))

	assert.True(t, m.Match("123key"))
	assert.True(t, m.Match("123pass"))
	assert.True(t, m.Match("123secret"))

	assert.True(t, m.Match("123key123"))
	assert.True(t, m.Match("123pass123"))
	assert.True(t, m.Match("123secret123"))

	// Test default matcher - no match
	assert.False(t, m.Match("ke"))
	assert.False(t, m.Match("pas"))
	assert.False(t, m.Match("secre"))

	assert.False(t, m.Match("ke123y"))
	assert.False(t, m.Match("pas123s"))
	assert.False(t, m.Match("sec123ret"))

	assert.True(t, m.Match("password"))
	assert.True(t, m.Match("PASSWORD"))
	assert.True(t, m.Match("123password123"))
	assert.True(t, m.Match("pass123word"))
	assert.False(t, m.Match("pas123sword"))

}
