// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package w3ctrace

import (
	"bytes"
	"regexp"
	"strings"
)

// VendorInstana is the Instana vendor key in the `tracestate` list
const VendorInstana = "in"

// Length of entries that should be filtered first in case, if tracestate has more than `MaxStateEntries` items
const thresholdLen = 128

var instanaListMemberRegex = regexp.MustCompile("^\\s*" + VendorInstana + "\\s*=\\s*([^,]*)\\s*$")

// State is list of key=value pairs representing vendor-specific data in the trace context
type State struct {
	// The key-value pairs of other vendors. The "in" list member, if present, will be stored separately
	// (Note: list member is the term the W3C trace context specfication uses for the key-value pairs.)
	listMembers []string

	// The value for the Instana-specific list member ("in=...").
	instanaTraceStateValue string
}

// NewState creates a new State with the given values
func NewState(listMembers []string, instanaTraceStateValue string) State {
	return State{
		listMembers:            listMembers,
		instanaTraceStateValue: instanaTraceStateValue,
	}
}

// FormStateWithInstanaTraceStateValue returns a new state prepended with the provided Instana value. If the original state had an Instana
// list member pair, it is discarded/overwritten.
func FormStateWithInstanaTraceStateValue(st State, instanaTraceStateValue string) State {
	var listMembers []string
	if st.instanaTraceStateValue == "" && instanaTraceStateValue != "" && len(st.listMembers) == MaxStateEntries {
		// The incoming tracestate had the maximum number of list members but no Instana list member, now we are adding an
		// Instana list member, so we would exceed the maximum by one. Hence, we need to discard one of the other list
		// members to stay within the limits mandated by the specification.
		listMembers = st.listMembers[:MaxStateEntries-1]
	} else {
		listMembers = st.listMembers
	}

	return State{listMembers: listMembers, instanaTraceStateValue: instanaTraceStateValue}
}

// ParseState parses the value of `tracestate` header. Empty list items are omitted.
func ParseState(traceStateValue string) State {
	listMembers := filterEmptyItems(strings.Split(traceStateValue, ","))

	// Look for the Instana list member first, before discarding any list members due to length restrictions.
	instanaTraceStateValue, instanaTraceStateIdx := searchInstanaHeader(listMembers)

	if instanaTraceStateIdx >= 0 {
		// remove the entry for instana from the array of list members
		listMembers = append(listMembers[:instanaTraceStateIdx], listMembers[instanaTraceStateIdx+1:]...)
	}

	// Depending on whether we found an Instana list member, we can either allow 31 or 32 list members from other vendors.
	maxListMembers := MaxStateEntries
	if instanaTraceStateValue != "" {
		maxListMembers--
	}

	if len(listMembers) < maxListMembers {
		return State{listMembers: listMembers, instanaTraceStateValue: instanaTraceStateValue}
	}

	itemsToFilter := len(listMembers) - maxListMembers
	filteredListMembers := listMembers[:0]
	i := 0
	for ; itemsToFilter > 0 && i < len(listMembers); i++ {
		if len(listMembers[i]) > thresholdLen {
			itemsToFilter--
			continue
		}
		filteredListMembers = append(filteredListMembers, listMembers[i])
	}
	filteredListMembers = append(filteredListMembers, listMembers[i:]...)

	if len(filteredListMembers) > maxListMembers {
		return State{listMembers: filteredListMembers[:maxListMembers], instanaTraceStateValue: instanaTraceStateValue}
	}

	return State{listMembers: filteredListMembers, instanaTraceStateValue: instanaTraceStateValue}
}

// FetchInstanaTraceStateValue retrieves the value of the Instana tracestate list member, if any.
func (st State) FetchInstanaTraceStateValue() (string, bool) {
	return st.instanaTraceStateValue, st.instanaTraceStateValue != ""
}

// String returns string representation of a trace state. The returned value is compatible with the
// `tracestate` header format. If the state has an Instana-specific list member, that one is always rendered first. This
// is optimized for the use case of injecting the string representation of the tracestate header into downstream
// requests.
func (st State) String() string {
	if len(st.listMembers) == 0 && st.instanaTraceStateValue == "" {
		return ""
	}
	if len(st.listMembers) == 0 {
		return VendorInstana + "=" + st.instanaTraceStateValue
	}

	buf := bytes.NewBuffer(nil)
	if st.instanaTraceStateValue != "" {
		buf.WriteString(VendorInstana)
		buf.WriteString("=")
		buf.WriteString(st.instanaTraceStateValue)
		buf.WriteString(",")
	}
	for _, vd := range st.listMembers {
		buf.WriteString(vd)
		buf.WriteByte(',')
	}
	buf.Truncate(buf.Len() - 1) // remove trailing comma

	return buf.String()
}

func filterEmptyItems(entries []string) []string {
	result := entries[:0]
	for _, v := range entries {
		if v != "" {
			result = append(result, v)
		}
	}

	return result
}

func searchInstanaHeader(listMembers []string) (string, int) {
	var instanaTraceStateValue string
	instanaTraceStateIdx := -1
	for i, vd := range listMembers {
		matchResult := instanaListMemberRegex.FindStringSubmatch(vd)
		if len(matchResult) == 2 {
			instanaTraceStateValue = strings.TrimSpace(matchResult[1])
			instanaTraceStateIdx = i
			break
		}
	}
	return instanaTraceStateValue, instanaTraceStateIdx
}
