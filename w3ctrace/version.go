package w3ctrace

import (
	"fmt"
	"strconv"
)

// Version represents the W3C trace context version. It defines the format of `traceparent` header
type Version uint8

const (
	// Invalid W3C Trace Context version
	Version_Invalid Version = iota
	// Supported versions of W3C Trace Context headers
	Version_0
	// The latest supported version of W3C Trace Context
	Version_Max = Version_0
)

// ParseVersion parses the version part of a `traceparent` header value. It returns ErrContextCorrupted
// if the version is malformed
func ParseVersion(s string) (Version, error) {
	if len(s) < 2 || (len(s) > 2 && s[2] != '-') {
		return Version_Invalid, ErrContextCorrupted
	}
	s = s[:2]

	if s == "ff" {
		return Version_Invalid, nil
	}

	ver, err := strconv.ParseUint(s, 16, 8)
	if err != nil {
		return Version_Invalid, ErrContextCorrupted
	}

	return Version(ver + 1), nil
}

// String returns string representation of a trace parent version. The returned value is compatible with the
// `traceparent` header format. The caller should take care of handling the Version_Unknown, otherwise this
// method will return "ff" which is considered invalid
func (ver Version) String() string {
	if ver == Version_Invalid {
		return "ff"
	}

	return fmt.Sprintf("%02x", uint8(ver)-1)
}
