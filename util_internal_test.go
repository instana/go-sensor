package instana

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Trace IDs (and Span IDs) are based on Java Signed Long datatype
const MinUint64 = uint64(0)
const MaxUint64 = uint64(18446744073709551615)
const MinInt64 = int64(-9223372036854775808)
const MaxInt64 = int64(9223372036854775807)

func TestGeneratedIDRange(t *testing.T) {
	var count = 10000
	for index := 0; index < count; index++ {
		id := randomID()
		assert.True(t, id <= 9223372036854775807, "Generated ID is out of bounds (+)")
		assert.True(t, id >= -9223372036854775808, "Generated ID is out of bounds (-)")
	}
}

func TestIDConversionBackForth(t *testing.T) {
	maxID := int64(9223372036854775807)
	minID := int64(-9223372036854775808)
	maxHex := "7fffffffffffffff"
	minHex := "8000000000000000"

	// Place holders
	var header string
	var id int64

	// maxID (int64) -> header -> int64
	header, _ = ID2Header(maxID)
	id, _ = Header2ID(header)
	assert.Equal(t, maxHex, header, "ID2Header incorrect result.")
	assert.Equal(t, maxID, id, "Convert back into original is wrong")

	// minHex (unsigned 64bit hex string) -> signed 64bit int -> unsigned 64bit hex string
	id, _ = Header2ID(minHex)
	header, _ = ID2Header(id)
	assert.Equal(t, minID, id, "Header2ID incorrect result")
	assert.Equal(t, minHex, header, "Convert back into original is wrong")
}

func TestBogusValues(t *testing.T) {
	var id int64

	// Header2ID with random strings should return 0
	id, err := Header2ID("this shouldnt work")
	assert.Equal(t, int64(0), id, "Bad input should return 0")
	assert.NotNil(t, err, "An error should be returned")
}
