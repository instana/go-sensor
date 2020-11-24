package instana

import (
	"errors"
	"io/ioutil"
	"os"
	"testing"

	"github.com/instana/testify/assert"
	"github.com/instana/testify/require"
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

func TestIDConversion(t *testing.T) {
	// Place holders
	var header string
	var id int64

	header = FormatID(-7815363404733516491)
	assert.Equal(t, "938a406416457535", header, "FormatID incorrect result.")

	id, err := ParseID("938a406416457535")
	require.NoError(t, err)
	assert.Equal(t, int64(-7815363404733516491), id, "ParseID incorrect result")

	header = FormatID(307170163380978816)
	assert.Equal(t, "44349a2d9ec0480", header, "FormatID incorrect result.")

	id, err = ParseID("44349a2d9ec0480") // Without a leading zero
	require.NoError(t, err)
	assert.Equal(t, int64(307170163380978816), id, "ParseID incorrect result")

	id, err = ParseID("044349a2d9ec0480") // Try with a leading zero
	require.NoError(t, err)
	assert.Equal(t, int64(307170163380978816), id, "ParseID incorrect result")

	header = FormatID(2920004540187184976)
	assert.Equal(t, "2885f0a890628f50", header, "FormatID incorrect result.")

	id, err = ParseID("2885f0a890628f50")
	require.NoError(t, err)
	assert.Equal(t, int64(2920004540187184976), id, "ParseID incorrect result")

	header = FormatID(16)
	assert.Equal(t, "10", header, "FormatID should drop leading zeros")

	id, err = ParseID("0000000000000010")
	require.NoError(t, err)
	assert.Equal(t, int64(16), id, "ParseID should stll work with leading zeros")

	id, err = ParseID("10")
	require.NoError(t, err)
	assert.Equal(t, int64(16), id, "ParseID should convert <16 char strings")

	count := 10000
	for index := 0; index < count; index++ {
		generatedID := randomID()

		header := FormatID(generatedID)

		id, err := ParseID(header)
		require.NoError(t, err)
		assert.Equal(t, generatedID, id, "Original ID does not match converted back ID")
	}
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
		"64-bit":  {0x0, 0x1234ab, "1234ab"},
		"128-bit": {0x1c, 0x1234ab, "1c00000000001234ab"},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, example.Expected, FormatLongID(example.Hi, example.Lo))
		})
	}
}

func TestParseFormatLongID(t *testing.T) {
	examples := map[string]string{
		"64-bit":  "1234abcd1234abcd",
		"128-bit": "5678defa5678defa1234abcd1234abcd",
	}

	for name, id := range examples {
		t.Run(name, func(t *testing.T) {
			hi, lo, err := ParseLongID(id)
			require.NoError(t, err)

			assert.Equal(t, id, FormatLongID(hi, lo))
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
