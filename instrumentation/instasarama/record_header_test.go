// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instasarama_test

import (
	"testing"

	"github.com/instana/go-sensor/instrumentation/instasarama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPackUnpackTraceContextHeader(t *testing.T) {
	examples := map[string]struct {
		TraceID, SpanID string
		Expected        [24]byte // using fixed len array here to avoid typos in examples
	}{
		"empty values": {
			Expected: [24]byte{
				// trace id
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				// span id
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			},
		},
		"with short 64-bit trace id, no span id": {
			TraceID: "00000000000000000000000deadbeef1",
			Expected: [24]byte{
				// trace id
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x0d, 0xea, 0xdb, 0xee, 0xf1,
				// span id
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			},
		},
		"with 64-bit trace id, no span id": {
			TraceID: "0000000000000000deadbeefdeadbeef",
			Expected: [24]byte{
				// trace id
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0xde, 0xad, 0xbe, 0xef, 0xde, 0xad, 0xbe, 0xef,
				// span id
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			},
		},
		"no trace id, with short 64-bit span id": {
			SpanID: "00000000deadbeef",
			Expected: [24]byte{
				// trace id
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				// span id
				0x00, 0x00, 0x00, 0x00, 0xde, 0xad, 0xbe, 0xef,
			},
		},
		"no trace id, with 64-bit span id": {
			SpanID: "deadbeefdeadbeef",
			Expected: [24]byte{
				// trace id
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				// span id
				0xde, 0xad, 0xbe, 0xef, 0xde, 0xad, 0xbe, 0xef,
			},
		},
		"with 64-bit trace id and 64-bit span id": {
			TraceID: "0000000000000000000000000000abcd",
			SpanID:  "0000000deadbeef1",
			Expected: [24]byte{
				// trace id
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xab, 0xcd,
				// span id
				0x00, 0x00, 0x00, 0x0d, 0xea, 0xdb, 0xee, 0xf1,
			},
		},
		"with 128-bit trace id and 64-bit span id": {
			TraceID: "000000000000abcd000000000000ef12",
			SpanID:  "0000000deadbeef1",
			Expected: [24]byte{
				// trace id
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xab, 0xcd,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xef, 0x12,
				// span id
				0x00, 0x00, 0x00, 0x0d, 0xea, 0xdb, 0xee, 0xf1,
			},
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, example.Expected[:], instasarama.PackTraceContextHeader(example.TraceID, example.SpanID))

			traceID, spanID, err := instasarama.UnpackTraceContextHeader(example.Expected[:])
			require.NoError(t, err)

			assert.Equal(t, example.TraceID[:], traceID)
			assert.Equal(t, example.SpanID[:], spanID)
		})
	}
}

func TestUnpackTraceContextHeader_WrongBufferSize(t *testing.T) {
	examples := map[string][]byte{
		"nil":       nil,
		"too long":  make([]byte, 23),
		"too short": make([]byte, 25),
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			_, _, err := instasarama.UnpackTraceContextHeader(example)
			assert.Error(t, err)
		})
	}
}

func TestPackUnpackTraceLevelHeader(t *testing.T) {
	// using fixed len arrays here to avoid typos in examples
	examples := map[string][1]byte{
		"0": {0x00},
		"1": {0x01},
	}

	for level, expected := range examples {
		t.Run("X-INSTANA-L="+level, func(t *testing.T) {
			assert.Equal(t, expected[:], instasarama.PackTraceLevelHeader(level))

			val, err := instasarama.UnpackTraceLevelHeader(expected[:])
			require.NoError(t, err)
			assert.Equal(t, level, val)
		})
	}
}

func TestUnpackTraceLevelHeader_WrongBufferSize(t *testing.T) {
	examples := map[string][]byte{
		"nil":       nil,
		"too long":  make([]byte, 2),
		"too short": make([]byte, 0),
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			_, err := instasarama.UnpackTraceLevelHeader(example)
			assert.Error(t, err)
		})
	}
}
