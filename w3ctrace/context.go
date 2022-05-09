// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package w3ctrace

import (
	"errors"
	"net/http"
	"strings"
)

const (
	// MaxStateEntries is the maximum number of items in `tracestate` as defined by
	// https://www.w3.org/TR/trace-context/#tracestate-header-field-values
	MaxStateEntries = 32

	// TraceParentHeader is the W3C trace parent header name as defined by https://www.w3.org/TR/trace-context/
	TraceParentHeader = "traceparent"
	// TraceStateHeader is the W3C trace state header name as defined by https://www.w3.org/TR/trace-context/
	TraceStateHeader = "tracestate"
)

var (
	// ErrContextNotFound is an error retuned by w3ctrace.Extract() if provided HTTP headers does not contain W3C trace context
	ErrContextNotFound = errors.New("no w3c context")
	// ErrContextCorrupted is an error retuned by w3ctrace.Extract() if provided HTTP headers contain W3C trace context in unexpected format
	ErrContextCorrupted = errors.New("corrupted w3c context")
	// ErrUnsupportedVersion is an error retuned by w3ctrace.Extract() if the version of provided W3C trace context is not supported
	ErrUnsupportedVersion = errors.New("unsupported w3c context version")
)

// Context represents the W3C trace context
type Context struct {
	RawParent string
	RawState  string
}

// New initializes a new W3C trace context from given parent
func New(parent Parent) Context {
	return Context{
		RawParent: parent.String(),
	}
}

// IsZero returns whether a context is a zero value
func (c Context) IsZero() bool {
	return c.RawParent == ""
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
	if trCtx.RawState != "" {
		headers.Set(TraceStateHeader, trCtx.RawState)
	}
}

// State parses RawState and returns the corresponding list.
func (c Context) State() State {
	return ParseState(c.RawState)
}

// Parent parses RawParent and returns the corresponding list.
// It silently discards malformed value. To check errors use ParseParent().
func (c Context) Parent() Parent {
	st, err := ParseParent(c.RawParent)
	if err != nil {
		return Parent{}
	}

	return st
}
