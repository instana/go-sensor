// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2017

package instana

import (
	"errors"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Trace IDs (and Span IDs) are based on Java Signed Long datatype
const (
	MinInt64 int64 = -9223372036854775808
	MaxInt64 int64 = 9223372036854775807
)

func TestGeneratedIDRange(t *testing.T) {
	for index := 0; index < 10000; index++ {
		id := randomID()
		assert.LessOrEqual(t, id, MaxInt64)
		assert.GreaterOrEqual(t, id, MinInt64)
	}
}

func TestParseID(t *testing.T) {
	maxHex := "7fffffffffffffff"

	// maxID (int64) -> header -> int64
	header := FormatID(MaxInt64)
	assert.Equal(t, maxHex, header)

	id, err := ParseID(header)
	require.NoError(t, err)
	assert.Equal(t, MaxInt64, id)
}

func TestFormatID(t *testing.T) {
	minID := int64(MinInt64)
	minHex := "8000000000000000"

	// minHex (unsigned 64bit hex string) -> signed 64bit int -> unsigned 64bit hex string
	id, err := ParseID(minHex)
	require.NoError(t, err)
	assert.Equal(t, minID, id)

	header := FormatID(id)
	assert.Equal(t, minHex, header)
}

func TestParseLongID(t *testing.T) {
	examples := map[string]struct {
		Value                  string
		ExpectedHi, ExpectedLo int64
	}{
		"64-bit short":             {"1234ab", 0x0, 0x1234ab},
		"64-bit padded":            {"00000000001234ab", 0x0, 0x1234ab},
		"64-bit padded to 128-bit": {"000000000000000000000000001234ab", 0x0, 0x1234ab},
		"128-bit short":            {"1c00000000001234ab", 0x1c, 0x1234ab},
		"128-bit padded":           {"000000000000001c00000000001234ab", 0x1c, 0x1234ab},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			hi, lo, err := ParseLongID(example.Value)
			require.NoError(t, err)

			assert.Equal(t, example.ExpectedHi, hi)
			assert.Equal(t, example.ExpectedLo, lo)
		})
	}
}

func TestFormatLongID(t *testing.T) {
	examples := map[string]struct {
		Hi, Lo   int64
		Expected string
	}{
		"64-bit":  {0x0, 0x1234ab, "000000000000000000000000001234ab"},
		"128-bit": {0x1c, 0x1234ab, "000000000000001c00000000001234ab"},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, example.Expected, FormatLongID(example.Hi, example.Lo))
		})
	}
}

func TestBogusValues(t *testing.T) {
	var id int64

	// ParseID with random strings should return 0
	id, err := ParseID("this shouldnt work")
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
		in          string
		expected    string
		expectError bool
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

			gateway, err := getDefaultGateway(tmpFile.Name())
			require.NoError(t, err)
			assert.Equal(t, test.expected, gateway)
		}()
	}
}

func BenchmarkFormatID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		FormatID(int64(i))
	}
}
