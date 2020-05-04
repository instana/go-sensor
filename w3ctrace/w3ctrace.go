package w3ctrace

import (
	"bytes"
	"errors"
	"net/http"
	"strings"
)

const (
	// The max number of items in `tracestate` as defined by https://www.w3.org/TR/trace-context/#tracestate-header-field-values
	MaxStateEntries = 32

	// W3C trace context header names as defined by https://www.w3.org/TR/trace-context/
	TraceParentHeader = "traceparent"
	TraceStateHeader  = "tracestate"
)

var (
	ErrContextNotFound    = errors.New("no w3c context")
	ErrContextCorrupted   = errors.New("corrupted w3c context")
	ErrUnsupportedVersion = errors.New("unsupported w3c context version")
)

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

// State parses RawState and returns the corresponding list.
// It silently discards malformed state. To check errors use ParseState().
func (trCtx Context) State() State {
	st, err := ParseState(trCtx.RawState)
	if err != nil {
		return State{}
	}

	return st
}

// State is list of key=value pairs representing vendor-specific data in the trace context
type State []string

// ParseState parses the value of `tracestate` header. It strips any optional white-space chararacters
// preceding or following the key=value pairs. Empty list items are omitted.
func ParseState(traceStateValue string) (State, error) {
	var state State

	for _, st := range strings.SplitN(traceStateValue, ",", 32) {
		st = strings.TrimSpace(st)
		if st == "" {
			continue
		}

		state = append(state, st)
	}

	return state, nil
}

// Put returns a new state prepended with provided vendor-specific data. It removes any existing
// entries for this vendor and returns the same state if vendor is empty. If the number of entries
// in a state reaches the MaxStateEntries, rest of the items will be truncated
func (st State) Add(vendor, data string) State {
	if vendor == "" {
		return st
	}

	newSt := make(State, 1, len(st)+1)
	newSt[0] = vendor + "=" + data
	newSt = append(newSt, st.Remove(vendor)...)

	// truncate the state if it reached the max number of entries
	if len(newSt) > MaxStateEntries {
		newSt = newSt[:MaxStateEntries]
	}

	return newSt
}

// Remove returns a new state without data for specified vendor. It returns the same state if vendor is empty
func (st State) Remove(vendor string) State {
	if vendor == "" {
		return st
	}

	prefix := vendor + "="

	var newSt State
	for _, vd := range st {
		if !strings.HasPrefix(vd, prefix) {
			newSt = append(newSt, vd)
		}
	}

	return newSt
}

// String returns string representation of a trace state. The returned value is compatible with the
// `tracestate` header format
func (st State) String() string {
	if len(st) == 0 {
		return ""
	}

	buf := bytes.NewBuffer(nil)
	for _, vd := range st {
		buf.WriteString(vd)
		buf.WriteByte(',')
	}
	buf.Truncate(buf.Len() - 1) // remove trailing comma

	return buf.String()
}
