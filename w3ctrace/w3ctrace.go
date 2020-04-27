package w3ctrace

import (
	"errors"
	"net/http"
	"strings"
)

const (
	// W3C trace context header names as defined by https://www.w3.org/TR/trace-context/
	TraceParentHeader = "traceparent"
	TraceStateHeader  = "tracestate"
)

var ErrContextNotFound = errors.New("no w3c context")

// Context represents the W3C trace context
type Context struct {
	RawParent string
	RawState  string
}

// Extract extracts the W3C trace context from HTTP headers. Returns ErrContextNotFound if
// provided value doesn't contain traceparent header.
func Extract(headers http.Header) (Context, error) {
	var tr Context

	for k, v := range headers {
		if len(v) == 0 {
			continue
		}

		switch {
		case strings.EqualFold(k, TraceParentHeader):
			tr.RawParent = v[0]
		case strings.EqualFold(k, TraceStateHeader):
			tr.RawState = v[0]
		}
	}

	if tr.RawParent == "" {
		return tr, ErrContextNotFound
	}

	return tr, nil
}

// Inject adds the w3c trace context headers, overriding any previously set values
func Inject(trCtx Context, headers http.Header) {
	// delete existing headers ignoring the header name case
	for k := range headers {
		if strings.EqualFold(k, TraceParentHeader) || strings.EqualFold(k, TraceStateHeader) {
			delete(headers, k)
		}
	}

	headers.Set(TraceParentHeader, trCtx.RawParent)
	headers.Set(TraceStateHeader, trCtx.RawState)
}
