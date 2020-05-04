package w3ctrace

// Flags contains the trace flags as defined by https://www.w3.org/TR/trace-context/#trace-flags
type Flags struct {
	Sampled bool
}

// Parent represents trace parent extracted from `traceparent` header
type Parent struct {
	Version  Version
	TraceID  string
	ParentID string
	Flags    Flags
}

// ParseParent parses the value of `traceparent` header according to the version
// defined in the first field
func ParseParent(s string) (Parent, error) {
	ver, err := ParseVersion(s)
	if err != nil {
		return Parent{}, ErrContextCorrupted
	}

	return ver.parseParent(s)
}

// String returns string representation of a trace parent. The returned value is compatible with the
// `traceparent` header format
func (p Parent) String() string {
	return p.Version.formatParent(p)
}
