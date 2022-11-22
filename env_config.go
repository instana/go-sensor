// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instana

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// parseInstanaTags parses the tags string passed via INSTANA_TAGS.
// The tag string is a comma-separated list of keys optionally followed by an '=' character and a string value:
//
//	INSTANA_TAGS := key1[=value1][,key2[=value2],...]
//
// The leading and trailing space is truncated from key names, values are used as-is. If a key does not have
// value associated, it's considered to be nil.
func parseInstanaTags(s string) map[string]interface{} {
	tags := make(map[string]interface{})

	for _, tag := range strings.Split(s, ",") {
		kv := strings.SplitN(tag, "=", 2)

		k := strings.TrimSpace(kv[0])
		if k == "" {
			continue
		}

		var v interface{}
		if len(kv) > 1 {
			v = kv[1]
		}

		tags[k] = v
	}

	if len(tags) == 0 {
		return nil
	}

	return tags
}

// parseInstanaSecrets parses the tags string passed via INSTANA_SECRETS.
// The secrets matcher configuration string is expected to have the following format:
//
//	INSTANA_SECRETS := <matcher>:<secret>[,<secret>]
//
// Where `matcher` is one of:
// * `equals` - matches a string if it's contained in the secrets list
// * `equals-ignore-case` is a case-insensitive version of `equals`
// * `contains` matches a string if it contains any of the secrets list values
// * `contains-ignore-case` is a case-insensitive version of `contains`
// * `regex` matches a string if it fully matches any of the regular expressions provided in the secrets list
//
// This function returns an error if there is no matcher configuration provided.
func parseInstanaSecrets(s string) (Matcher, error) {
	if s == "" {
		return nil, errors.New("empty value for secret matcher configuration")
	}

	ind := strings.Index(s, ":")
	if ind < 0 {
		return nil, fmt.Errorf("malformed secret matcher configuration: %q", s)
	}

	matcher, config := strings.TrimSpace(s[:ind]), strings.Split(s[ind+1:], ",")

	return NamedMatcher(matcher, config)
}

// parseInstanaExtraHTTPHeaders parses the tags string passed via INSTANA_EXTRA_HTTP_HEADERS.
// The header names are expected to come in a semicolon-separated list:
//
//	INSTANA_EXTRA_HTTP_HEADERS := header1[;header2;...]
//
// Any leading and trailing whitespace characters will be trimmed from header names.
func parseInstanaExtraHTTPHeaders(s string) []string {
	var headers []string
	for _, h := range strings.Split(s, ";") {
		h = strings.TrimSpace(h)
		if h == "" {
			continue
		}

		headers = append(headers, h)
	}

	return headers
}

// parseInstanaTimeout parses the Instana backend connection timeout passed via INSTANA_TIMEOUT.
// The value is expected to be an integer number of milliseconds, greate than 0.
// This function returns the default timeout 500ms if provided with an empty string.
func parseInstanaTimeout(s string) (time.Duration, error) {
	if s == "" {
		return defaultServerlessTimeout, nil
	}

	ms, err := strconv.ParseUint(s, 10, 64)
	if err != nil || ms < 1 {
		return 0, fmt.Errorf("invalid timeout value: %q", s)
	}

	return time.Duration(ms) * time.Millisecond, nil
}
