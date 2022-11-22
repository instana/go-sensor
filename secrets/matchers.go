// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package secrets

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

// NoneMatcher does not match any string. It's used as a default value for (instana.Options).Secrets
type NoneMatcher struct{}

// Match returns false for any string
func (m NoneMatcher) Match(s string) bool { return false }

// EqualsMatcher matches a string that is contained in the terms list. This is the
// matcher for the 'equals' match type
type EqualsMatcher struct {
	list []string
}

// NewEqualsMatcher returns an EqualsMatcher for a list of terms
func NewEqualsMatcher(terms ...string) EqualsMatcher {
	return EqualsMatcher{terms}
}

// Match returns true if provided value present in matcher's terms list
func (m EqualsMatcher) Match(s string) bool {
	for _, term := range m.list {
		if term == s {
			return true
		}
	}

	return false
}

// EqualsIgnoreCaseMatcher is the case-insensitive version of EqualsMatcher. This
// is the matcher for the 'equals-ignore-case' match type
type EqualsIgnoreCaseMatcher struct {
	m EqualsMatcher
}

// NewEqualsIgnoreCaseMatcher returns an EqualsIgnoreCaseMatcher for a list of terms
func NewEqualsIgnoreCaseMatcher(terms ...string) EqualsIgnoreCaseMatcher {
	for i := range terms {
		terms[i] = strings.ToLower(terms[i])
	}

	return EqualsIgnoreCaseMatcher{
		m: NewEqualsMatcher(terms...),
	}
}

// Match returns true if provided value present in matcher's terms list regardless of the case
func (m EqualsIgnoreCaseMatcher) Match(s string) bool {
	return m.m.Match(strings.ToLower(s))
}

// ContainsMatcher matches a string if it contains at any of the terms in the matcher's list.
// This is the matcher for the 'contains' match type
type ContainsMatcher struct {
	list []string
}

// NewContainsMatcher returns a ContainsMatcher for a list of terms
func NewContainsMatcher(terms ...string) ContainsMatcher {
	return ContainsMatcher{terms}
}

// Match returns true if a string contains any of matcher's terms
func (m ContainsMatcher) Match(s string) bool {
	for _, term := range m.list {
		if strings.Contains(s, term) {
			return true
		}
	}

	return false
}

// ContainsIgnoreCaseMatcher is the case-insensitive version of ContainsMatcher. This
// is the matcher for the 'contains-ignore-case' match type
type ContainsIgnoreCaseMatcher struct {
	m ContainsMatcher
}

// NewContainsIgnoreCaseMatcher returns a ContainsIgnoreCaseMatcher for a list of terms
func NewContainsIgnoreCaseMatcher(terms ...string) ContainsIgnoreCaseMatcher {
	for i := range terms {
		terms[i] = strings.ToLower(terms[i])
	}

	return ContainsIgnoreCaseMatcher{
		m: NewContainsMatcher(terms...),
	}
}

// Match returns true if a string contains any of matcher's terms regardless of the case
func (m ContainsIgnoreCaseMatcher) Match(s string) bool {
	return m.m.Match(strings.ToLower(s))
}

var matchNothingRegexpMatcher = RegexpMatcher{regexp.MustCompile(".^")}

// RegexpMatcher matches a string using a set of regular expressions. This is the matcher
// for the 'regex' match type
type RegexpMatcher struct {
	re *regexp.Regexp
}

// NewRegexpMatcher returns a RegexpMatcher for a list of regular expressions
func NewRegexpMatcher(terms ...*regexp.Regexp) (RegexpMatcher, error) {
	if len(terms) == 0 {
		return matchNothingRegexpMatcher, nil
	}

	// combine expressions into one using OR, i.e.
	// [RE1, RE2, ..., REn] -> (RE1)|(RE2)|...|(REn)
	buf := bytes.NewBuffer([]byte(`(\A`))
	sep := []byte(`\z)|(\A`)

	for _, term := range terms {
		reBytes := []byte(term.String())
		// strip leading beginning-of-line matchers, as they are already included into the combined expression
		reBytes = bytes.TrimPrefix(bytes.TrimLeft(reBytes, "^"), []byte(`\A`))
		// strip trailing end-of-line matchers, as they are already included into the combined expression
		reBytes = bytes.TrimSuffix(bytes.TrimRight(reBytes, "$"), []byte(`\z`))

		buf.Write(reBytes)
		buf.Write(sep)
	}
	buf.Truncate(buf.Len() - len(sep)) // trim trailing separator
	buf.WriteString(`\z)`)

	combined := buf.String()

	re, err := regexp.Compile(combined)
	if err != nil {
		return matchNothingRegexpMatcher, fmt.Errorf("malformed regexp %q: %s", combined, err)
	}

	return RegexpMatcher{re}, nil
}

// Match returns true if a string fully matches any of matcher's regular expessions. If an expression matches only
// a part of string, this method returns false:
//
//	m := NewRegexpMatcher(regexp.MustCompile(`aaa`), regexp.MustCompile(`bbb`))
//	m.Match("aaa") // returns true
//	m.Match("bbb") // returns true
//	m.Match("aaabbb") // returns false, as both regular expressions match only a part of the string
func (m RegexpMatcher) Match(s string) bool {
	return m.re.MatchString(s)
}
