package w3ctrace

import (
	"bytes"
	"strings"
)

// Instana vendor key in the `tracestate` list
const VendorInstana = "in"

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

// Fetch retrieves stored vendor-specific data for given vendor
func (st State) Fetch(vendor string) (string, bool) {
	i := st.Index(vendor)
	if i < 0 {
		return "", false
	}

	return strings.TrimPrefix(st[i], vendor+"="), true
}

// Index returns the index of vendor-specific data for given vendor in the state.
// It returns -1 if the state does not contain data for this vendor.
func (st State) Index(vendor string) int {
	prefix := vendor + "="

	for i, vd := range st {
		if strings.HasPrefix(vd, prefix) {
			return i
		}
	}

	return -1
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
