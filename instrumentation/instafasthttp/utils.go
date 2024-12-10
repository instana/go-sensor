// (c) Copyright IBM Corp. 2024

package instafasthttp

import (
	"net/url"

	ot "github.com/opentracing/opentracing-go"
)

func setHeadersAndParamsToSpan(span ot.Span, headers map[string]string, params url.Values) {
	if len(headers) > 0 {
		span.SetTag("http.header", headers)
	}
	if len(params) > 0 {
		span.SetTag("http.params", params.Encode())
	}
}
