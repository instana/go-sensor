// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instana

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/instana/go-sensor/secrets"
)

const (
	// EqualsMatcher matches the string exactly
	EqualsMatcher = "equals"
	// EqualsIgnoreCaseMatcher matches the string exactly ignoring the case
	EqualsIgnoreCaseMatcher = "equals-ignore-case"
	// ContainsMatcher matches the substring in a string
	ContainsMatcher = "contains"
	// ContainsIgnoreCaseMatcher matches the substring in a string ignoring the case
	ContainsIgnoreCaseMatcher = "contains-ignore-case"
	// RegexpMatcher matches the string using a set of regular expressions. Each item in a term list
	// provided to instana.NamedMatcher() must be a valid regular expression that can be compiled using
	// regexp.Compile()
	RegexpMatcher = "regex"
	// NoneMatcher does not match any string
	NoneMatcher = "none"
)

// Matcher verifies whether a string meets predefined conditions
type Matcher interface {
	Match(s string) bool
}

// NamedMatcher returns a secrets matcher supported by Instana host agent configuration
//
// See https://www.instana.com/docs/setup_and_manage/host_agent/configuration/#secrets
func NamedMatcher(name string, list []string) (Matcher, error) {
	switch strings.ToLower(name) {
	case EqualsMatcher:
		return secrets.NewEqualsMatcher(list...), nil
	case EqualsIgnoreCaseMatcher:
		return secrets.NewEqualsIgnoreCaseMatcher(list...), nil
	case ContainsMatcher:
		return secrets.NewContainsMatcher(list...), nil
	case ContainsIgnoreCaseMatcher:
		return secrets.NewContainsIgnoreCaseMatcher(list...), nil
	case RegexpMatcher:
		var exps []*regexp.Regexp
		for _, s := range list {
			ex, err := regexp.Compile(s)
			if err != nil {
				sensor.logger.Warn("ignoring malformed regexp secrets matcher ", s, ": ", err)
				continue
			}

			exps = append(exps, ex)
		}

		return secrets.NewRegexpMatcher(exps...)
	case NoneMatcher:
		return secrets.NoneMatcher{}, nil
	default:
		return nil, fmt.Errorf("unknown secrets matcher type %q", name)
	}
}

// DefaultSecretsMatcher returns the default secrets matcher, that matches strings containing
// "key", "pass" and "secret" ignoring the case
func DefaultSecretsMatcher() Matcher {
	m, _ := NamedMatcher(ContainsIgnoreCaseMatcher, []string{"key", "pass", "secret"})
	return m
}
