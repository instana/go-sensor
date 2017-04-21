package instana

import "testing"

// Trace IDs (and Span IDs) are based on Java Signed Long datatype
const MinUint64 = uint64(0)
const MaxUint64 = uint64(18446744073709551615)
const MinInt64 = int64(-9223372036854775808)
const MaxInt64 = int64(9223372036854775807)

func TestIDGeneration(t *testing.T) {
}
