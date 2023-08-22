// (c) Copyright IBM Corp. 2023

//go:build go1.17
// +build go1.17

package instasarama

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"
)

// The following functions perform the packing and unpacking of the trace context
// according to https://github.com/instana/technical-documentation/tree/master/tracing/specification#kafka

// PackTraceContextHeader packs the trace and span ID into a byte slice to be used as (sarama.RecordHeader).Value.
// The returned slice is always 24 bytes long.
func PackTraceContextHeader(traceID, spanID string) []byte {
	buf := make([]byte, 24)

	// hex representation uses 2 bytes to encode one byte of information, which means that
	// the length of both trace and span IDs must be even. instana.FormatID() truncates leading
	// zeroes, which may lead to data corruption as hex.Decode() will ignore the incomplete byte
	// representation at the end
	traceID = strings.Repeat("0", len(traceID)%2) + traceID
	spanID = strings.Repeat("0", len(spanID)%2) + spanID

	// write the trace ID into the first 16 bytes with zero padding at the beginning
	if traceID != "" {
		hex.Decode(buf[16-hex.DecodedLen(len(traceID)):16], []byte(traceID))
	}

	// write the span ID into the last 8 bytes
	if spanID != "" {
		hex.Decode(buf[24-hex.DecodedLen(len(spanID)):], []byte(spanID))
	}

	return buf
}

// UnpackTraceContextHeader unpacks and returns the trace and span ID, padding them with zeroes
// to 32 and 16 characters correspondingly. It expects the provided buffer to have exactly 24 bytes.
func UnpackTraceContextHeader(val []byte) (string, string, error) {
	if len(val) != 24 {
		return "", "", fmt.Errorf("unexpected value length: want 24, got %d", len(val))
	}

	traceID := hex.EncodeToString(bytes.TrimLeft(val[:16], "\000"))
	if traceID != "" && len(traceID) < 32 {
		traceID = strings.Repeat("0", 32-len(traceID)) + traceID
	}

	spanID := hex.EncodeToString(bytes.TrimLeft(val[16:], "\000"))
	if spanID != "" && len(spanID) < 16 {
		spanID = strings.Repeat("0", 16-len(spanID)) + spanID
	}

	return traceID, spanID, nil
}

// PackTraceLevelHeader packs the X-INSTANA-L value into a byte slice to be used as (sarama.RecordHeader).Value.
// It returns a 1-byte slice containing 0x00 if the passed value is "0", and 0x01 otherwise.
func PackTraceLevelHeader(val string) []byte {
	switch val {
	case "0":
		return []byte{0x00}
	default:
		return []byte{0x01}
	}
}

// UnpackTraceLevelHeader returns "1" if the value contains a non-zero byte, and "0" otherwise.
// It expects the provided buffer to have exactly 1 byte.
func UnpackTraceLevelHeader(val []byte) (string, error) {
	if len(val) != 1 {
		return "", fmt.Errorf("unexpected value length: want 1, got %d", len(val))
	}

	switch val[0] {
	case 0x00:
		return "0", nil
	default:
		return "1", nil
	}
}
