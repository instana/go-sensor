// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2016

package instana

import (
	"bufio"
	"bytes"
	crand "crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

const (
	allowRootExitSpanEnv = "INSTANA_ALLOW_ROOT_EXIT_SPAN"
)

// randomID generates a random ID using crypto/rand package.
func randomID() int64 {
	var sr *big.Int
	var err error

	id := time.Now().UnixNano()
	if sr, err = crand.Int(crand.Reader, big.NewInt(id)); err != nil {
		return id // fallback ID if crypto/rand fails to generate random ID
	}

	id = sr.Int64()

	return id
}

// FormatID converts an Instana ID to a value that can be used in
// context propagation (such as HTTP headers). More specifically,
// this converts a signed 64 bit integer into an unsigned hex string.
// The resulting string is always padded with 0 to be 16 characters long.
func FormatID(id int64) string {
	// FIXME: We're assuming LittleEndian here

	// Convert uint64 to hex string equivalent and return that
	return padHexString(strconv.FormatUint(uint64(id), 16), 64)
}

// FormatLongID converts a 128-bit Instana ID passed in two quad words to an
// unsigned hex string suitable for context propagation.
func FormatLongID(hi, lo int64) string {
	return FormatID(hi) + FormatID(lo)
}

func padHexString(s string, bitSize int) string {
	if len(s) >= bitSize>>2 {
		return s
	}

	return strings.Repeat("0", bitSize>>2-len(s)) + s
}

// ParseID converts an header context value into an Instana ID.  More
// specifically, this converts an unsigned 64 bit hex value into a signed
// 64bit integer.
func ParseID(header string) (int64, error) {
	// FIXME: We're assuming LittleEndian here

	// Parse unsigned 64 bit hex string into unsigned 64 bit base 10 integer
	unsignedID, err := strconv.ParseUint(header, 16, 64)
	if err != nil {
		return 0, fmt.Errorf("context corrupted; could not convert value: %s", err)
	}

	// Write out _unsigned_ 64bit integer to byte buffer
	buf := bytes.NewBuffer(nil)
	if err := binary.Write(buf, binary.LittleEndian, unsignedID); err != nil {
		return 0, fmt.Errorf("context corrupted; could not convert value: %s", err)
	}

	// Read bytes back into _signed_ 64 bit integer
	var signedID int64
	if err := binary.Read(buf, binary.LittleEndian, &signedID); err != nil {
		return 0, fmt.Errorf("context corrupted; could not convert value: %s", err)
	}

	return signedID, nil
}

// ParseLongID converts an header context value into a 128-bit Instana ID. Both high and low
// quad words are returned as signed integers.
func ParseLongID(header string) (hi int64, lo int64, err error) {
	if len(header) > 16 {
		hi, err = ParseID(header[:len(header)-16])
		if err != nil {
			return 0, 0, fmt.Errorf("failed to parse the higher 4 bytes of a 128-bit integer: %s", err)
		}

		header = header[len(header)-16:]
	}

	lo, err = ParseID(header)

	return hi, lo, err
}

// ID2Header calls instana.FormatID() and returns its result and a nil error.
// This is kept here for backward compatibility with go-sensor@v1.x
//
// Deprecated: please use instana.FormatID() instead
func ID2Header(id int64) (string, error) {
	return FormatID(id), nil
}

// Header2ID calls instana.ParseID() and returns its result.
// This is kept here for backward compatibility with go-sensor@v1.x
//
// Deprecated: please use instana.ParseID() instead
func Header2ID(header string) (int64, error) {
	return ParseID(header)
}

func getProcCommandLine() (string, []string, bool) {
	var cmdlinePath string = "/proc/" + strconv.Itoa(os.Getpid()) + "/cmdline"

	cmdline, err := os.ReadFile(cmdlinePath)
	if err != nil {
		return "", nil, false
	}

	parts := strings.FieldsFunc(string(cmdline), func(c rune) bool {
		return c == '\u0000'
	})

	return parts[0], parts[1:], true
}

func getProcessEnv() map[string]string {
	osEnv := os.Environ()

	env := make(map[string]string, len(osEnv))
	for _, envVar := range osEnv {
		idx := strings.Index(envVar, "=")
		if idx < 0 {
			continue
		}

		env[envVar[:idx]] = envVar[idx+1:]
	}

	return env
}

func getDefaultGateway(routeTableFile string) (string, error) {
	routeTable, err := os.Open(routeTableFile)
	if err != nil {
		return "", fmt.Errorf("failed to open %s: %s", routeTableFile, err)
	}
	defer routeTable.Close()

	s := bufio.NewScanner(routeTable)
	for s.Scan() {
		entry := strings.Split(s.Text(), "\t")
		if len(entry) < 3 {
			continue
		}

		destination := entry[1]
		if destination == "00000000" {

			gateway, err := hexGatewayToAddr(entry[2])
			if err != nil {
				return "", err
			}

			return gateway, nil
		}
	}

	if err := s.Err(); err != nil {
		return "", fmt.Errorf("failed to read %s: %s", routeTableFile, err)
	}

	return "", nil
}

// hexGatewayToAddr converts the hex representation of the gateway address to string.
func hexGatewayToAddr(gateway string) (string, error) {
	// gateway address is encoded in reverse order in hex
	if len(gateway) != 8 {
		return "", errors.New("invalid gateway length")
	}

	n, err := strconv.ParseUint(gateway, 16, 32)

	if err != nil {
		return "", err
	}

	first := n >> 0 & 0xff
	second := n >> 8 & 0xff
	third := n >> 16 & 0xff
	fourth := n >> 24 & 0xff

	return fmt.Sprintf("%d.%d.%d.%d", first, second, third, fourth), nil

}

func cloneTags(t ot.Tags) ot.Tags {
	clone := ot.Tags{}

	for k, v := range t {
		clone[k] = v
	}

	return clone
}

func cloneMapStringString(t map[string]string) map[string]string {
	clone := map[string]string{}

	for k, v := range t {
		clone[k] = v
	}

	return clone
}

func isRootExitSpan(kind interface{}, isRootSpan bool) bool {

	switch kind {
	case ext.SpanKindRPCClientEnum, string(ext.SpanKindRPCClientEnum),
		ext.SpanKindProducerEnum, string(ext.SpanKindProducerEnum),
		"exit":
		return isRootSpan

	default:
		return false
	}
}

func allowRootExitSpan() bool {
	return os.Getenv(allowRootExitSpanEnv) == "1"
}
