// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instana

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/instana/go-sensor/secrets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseInstanaTags(t *testing.T) {
	examples := map[string]struct {
		Value    string
		Expected map[string]interface{}
	}{
		"empty":                   {"", nil},
		"single tag, empty key":   {"=value", nil},
		"single tag, no value":    {"key", map[string]interface{}{"key": nil}},
		"single tag, empty value": {"key=", map[string]interface{}{"key": ""}},
		"single tag, with value":  {"key=value", map[string]interface{}{"key": "value"}},
		"multiple tags, mixed": {
			`key1,  key2=  , key3   ="",key4=42`,
			map[string]interface{}{
				"key1": nil,
				"key2": "  ",
				"key3": `""`,
				"key4": "42",
			},
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, example.Expected, parseInstanaTags(example.Value))
		})
	}
}
func TestParseInstanaEmptySecrets(t *testing.T) {
	examples := map[string]struct {
		Value    string
		Expected error
	}{
		"empty": {"", errors.New("empty value for secret matcher configuration")},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			_, err := parseInstanaSecrets(example.Value)
			require.Error(t, err)
			assert.Equal(t, example.Expected, err)
		})
	}
}

func TestParseInstanaSecrets(t *testing.T) {
	regexMatcher, err := secrets.NewRegexpMatcher(regexp.MustCompile("a|b|c"), regexp.MustCompile("d"))
	require.NoError(t, err)

	examples := map[string]struct {
		Value    string
		Expected Matcher
	}{
		"equals":               {"equals:a,b,c", secrets.NewEqualsMatcher("a", "b", "c")},
		"equals-ignore-case":   {"equals-ignore-case:a,b,c", secrets.NewEqualsIgnoreCaseMatcher("a", "b", "c")},
		"contains":             {"contains:a,b,c", secrets.NewContainsMatcher("a", "b", "c")},
		"contains-ignore-case": {"contains-ignore-case:a,b,c", secrets.NewContainsIgnoreCaseMatcher("a", "b", "c")},
		"regexp":               {"regex:a|b|c,d", regexMatcher},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			m, err := parseInstanaSecrets(example.Value)
			require.NoError(t, err)
			assert.Equal(t, example.Expected, m)
		})
	}
}

func TestParseInstanaSecrets_Error(t *testing.T) {
	examples := map[string]string{
		"unknown matcher": "magic:pew,pew",
		"malformed":       "equals;a,b,c",
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			_, err := parseInstanaSecrets(example)
			assert.Error(t, err)
		})
	}
}

func TestParseInstanaExtraHTTPHeaders(t *testing.T) {
	examples := map[string]struct {
		Value    string
		Expected []string
	}{
		"empty":    {"", nil},
		"one":      {"a", []string{"a"}},
		"multiple": {"a; ;  b  ;c;", []string{"a", "b", "c"}},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, example.Expected, parseInstanaExtraHTTPHeaders(example.Value))
		})
	}
}

func TestParseInstanaTimeout(t *testing.T) {
	examples := map[string]struct {
		Value    string
		Expected time.Duration
	}{
		"empty":            {"", defaultServerlessTimeout},
		"positive integer": {"123", 123 * time.Millisecond},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			d, err := parseInstanaTimeout(example.Value)
			require.NoError(t, err)
			assert.Equal(t, example.Expected, d)
		})
	}
}

func TestParseInstanaTimeout_Error(t *testing.T) {
	examples := map[string]string{
		"non-number":       "twenty",
		"non-integer":      "12.5:",
		"zero":             "0",
		"negative integer": "-100",
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			_, err := parseInstanaTimeout(example)
			assert.Error(t, err)
		})
	}
}
