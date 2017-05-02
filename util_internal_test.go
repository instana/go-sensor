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
	var id = int64(0)
	for index := 0; index < count; index++ {
		id = randomID()

		assert.True(t, id <= 9223372036854775807, "Generated ID is out of bounds (+)")
		assert.True(t, id >= -9223372036854775808, "Generated ID is out of bounds (-)")

	}
}
