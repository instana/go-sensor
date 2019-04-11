package instana

import (
	"errors"
	"io/ioutil"
	"os"
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

func TestIDConversion(t *testing.T) {
	// Place holders
	var header string
	var id int64

	header, _ = ID2Header(-7815363404733516491)
	assert.Equal(t, "938a406416457535", header, "ID2Header incorrect result.")
	id, _ = Header2ID("938a406416457535")
	assert.Equal(t, int64(-7815363404733516491), id, "Header2ID incorrect result")

	header, _ = ID2Header(307170163380978816)
	assert.Equal(t, "44349a2d9ec0480", header, "ID2Header incorrect result.")
	id, _ = Header2ID("44349a2d9ec0480") // Without a leading zero
	assert.Equal(t, int64(307170163380978816), id, "Header2ID incorrect result")
	id, _ = Header2ID("044349a2d9ec0480") // Try with a leading zero
	assert.Equal(t, int64(307170163380978816), id, "Header2ID incorrect result")

	header, _ = ID2Header(2920004540187184976)
	assert.Equal(t, "2885f0a890628f50", header, "ID2Header incorrect result.")
	id, _ = Header2ID("2885f0a890628f50")
	assert.Equal(t, int64(2920004540187184976), id, "Header2ID incorrect result")

	header, _ = ID2Header(16)
	assert.Equal(t, "10", header, "ID2Header should drop leading zeros")
	id, _ = Header2ID("0000000000000010")
	assert.Equal(t, int64(16), id, "Header2ID should stll work with leading zeros")
	id, _ = Header2ID("10")
	assert.Equal(t, int64(16), id, "Header2ID should convert <16 char strings")

	count := 10000
	for index := 0; index < count; index++ {
		generatedID := randomID()
		header, _ := ID2Header(generatedID)
		id, _ := Header2ID(header)
		assert.Equal(t, generatedID, id, "Original ID does not match converted back ID")
	}
}

func TestBogusValues(t *testing.T) {
	var id int64

	// Header2ID with random strings should return 0
	id, err := Header2ID("this shouldnt work")
	assert.Equal(t, int64(0), id, "Bad input should return 0")
	assert.NotNil(t, err, "An error should be returned")
}

func TestHexGatewayToAddr(t *testing.T) {
	tests := []struct {
		in          string
		expected    string
		expectedErr error
	}{
		{
			in:          "0101FEA9",
			expected:    "169.254.1.1",
			expectedErr: nil,
		},
		{
			in:          "0101FEAC",
			expected:    "172.254.1.1",
			expectedErr: nil,
		},
		{
			in:          "0101FEA",
			expected:    "",
			expectedErr: errors.New("invalid gateway length"),
		},
	}

	for _, test := range tests {
		gatewayHex := []rune(test.in)
		gateway, err := hexGatewayToAddr(gatewayHex)
		assert.Equal(t, test.expectedErr, err)
		assert.Equal(t, test.expected, gateway)
	}
}

func TestGetDefaultGateway(t *testing.T) {

	tests := []struct {
		in       string
		expected string
	}{
		{
			in: `Iface	Destination	Gateway 	Flags	RefCnt	Use	Metric	Mask		MTU	Window	IRTT

eth0	00000000	0101FEA9	0003	0	0	0	00000000	0	0	0

eth0	0101FEA9	00000000	0005	0	0	0	FFFFFFFF	0	0	0

`,
			expected: "169.254.1.1",
		},
		{
			in: `Iface	Destination	Gateway 	Flags	RefCnt	Use	Metric	Mask		MTU	Window	IRTT
										 
eth0	000011AC	00000000	0001	0	0	0	0000FFFF	0	0	0

eth0	00000000	010011AC	0003	0	0	0	00000000	0	0	0
                                                                               
`,
			expected: "172.17.0.1",
		},
	}

	for _, test := range tests {
		func() {
			tmpFile, err := ioutil.TempFile("", "getdefaultgateway")
			if err != nil {
				t.Fatal(err)
			}

			defer os.Remove(tmpFile.Name())

			_, err = tmpFile.WriteString(test.in)

			if err != nil {
				t.Fatal(err)
			}

			gateway := getDefaultGateway(tmpFile.Name())

			assert.Equal(t, test.expected, gateway)
		}()
	}
}
