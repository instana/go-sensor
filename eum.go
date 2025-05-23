// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2018

package instana

import (
	"bytes"
	"os"
	"sort"
	"strings"
)

const eumTemplate string = "eum.js"

// EumSnippet generates javascript code to initialize JavaScript agent
//
// Deprecated: this snippet is outdated and this method will be removed in
// the next major version. To learn about the way to install Instana EUM snippet
// please refer to https://docs.instana.io/products/website_monitoring/#installation
func EumSnippet(apiKey string, traceID string, meta map[string]string) string {

	if len(apiKey) == 0 || len(traceID) == 0 {
		return ""
	}

	b, err := os.ReadFile(eumTemplate)
	if err != nil {
		return ""
	}

	snippet := strings.Replace(string(b), "$apiKey", apiKey, -1)
	snippet = strings.Replace(snippet, "$traceId", traceID, -1)

	var keys []string
	for k := range meta {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var metaBuffer bytes.Buffer
	for _, k := range keys {
		metaBuffer.WriteString("ineum('meta','" + k + "','" + meta[k] + "');")
	}

	snippet = strings.Replace(snippet, "$meta", metaBuffer.String(), -1)

	return snippet
}
