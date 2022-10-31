// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package secrets_test

import (
	"regexp"
	"testing"

	"github.com/instana/go-sensor/secrets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEqualsMatcher(t *testing.T) {
	m := secrets.NewEqualsMatcher("one", "two")
	examples := map[string]bool{
		"":          false,
		"one":       true,
		"two":       true,
		"ONE":       false,
		"tWo":       false,
		"onetwo":    false,
		"two ":      false,
		" one":      false,
		"forty-two": false,
		"hello!":    false,
	}

	for s, expected := range examples {
		t.Run(s, func(t *testing.T) {
			assert.Equal(t, expected, m.Match(s))
		})
	}
}

func TestEqualsIgnoreCaseMatcher(t *testing.T) {
	m := secrets.NewEqualsIgnoreCaseMatcher("One", "TWO")
	examples := map[string]bool{
		"":          false,
		"one":       true,
		"two":       true,
		"ONE":       true,
		"tWo":       true,
		"onetwo":    false,
		"two ":      false,
		" one":      false,
		"forty-two": false,
		"hello!":    false,
	}

	for s, expected := range examples {
		t.Run(s, func(t *testing.T) {
			assert.Equal(t, expected, m.Match(s))
		})
	}
}

func TestContainsMatcher(t *testing.T) {
	m := secrets.NewContainsMatcher("one", "two")
	examples := map[string]bool{
		"":          false,
		"one":       true,
		"two":       true,
		"ONE":       false,
		"tWo":       false,
		"onetwo":    true,
		"two ":      true,
		" one":      true,
		"forty-two": true,
		"hello!":    false,
	}

	for s, expected := range examples {
		t.Run(s, func(t *testing.T) {
			assert.Equal(t, expected, m.Match(s))
		})
	}
}

func TestContainsIgnoreCaseMatcher(t *testing.T) {
	m := secrets.NewContainsIgnoreCaseMatcher("one", "two")
	examples := map[string]bool{
		"":          false,
		"one":       true,
		"two":       true,
		"ONE":       true,
		"tWo":       true,
		"onetwo":    true,
		"two ":      true,
		" one":      true,
		"forty-TWO": true,
		"hello!":    false,
	}

	for s, expected := range examples {
		t.Run(s, func(t *testing.T) {
			assert.Equal(t, expected, m.Match(s))
		})
	}
}

func TestRegexpMatcher(t *testing.T) {
	m, err := secrets.NewRegexpMatcher(regexp.MustCompile(`(?i)\Aone\z`), regexp.MustCompile(`^two$`))
	require.NoError(t, err)

	examples := map[string]bool{
		"":          false,
		"one":       true,
		"two":       true,
		"ONE":       true,
		"tWo":       false,
		"onetwo":    false,
		"two ":      false,
		" one":      false,
		"forty-TWO": false,
		"hello!":    false,
	}

	for s, expected := range examples {
		t.Run(s, func(t *testing.T) {
			assert.Equal(t, expected, m.Match(s))
		})
	}
}
