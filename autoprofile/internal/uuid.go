// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package internal

import (
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync/atomic"
	"time"
)

var (
	nextID int64
)

// GenerateUUID generates a new UUID string
func GenerateUUID() string {
	n := atomic.AddInt64(&nextID, 1)

	uuid := fmt.Sprintf("%019d%09d%010d",
		time.Now().UnixNano(),
		secureRandom(1_000_000_000),
		n,
	)

	return sha256String(uuid)
}

func secureRandom(max int64) int64 {
	nBig, err := crand.Int(crand.Reader, big.NewInt(max))
	if err != nil {
		panic(err)
	}
	return nBig.Int64()
}

func sha256String(s string) string {
	sh256 := sha256.New()
	sh256.Write([]byte(s))

	return hex.EncodeToString(sh256.Sum(nil))
}

func GenerateUUIDv4() string {
	b := make([]byte, 16) // 128-bit UUID
	_, err := crand.Read(b)
	if err != nil {
		panic(err)
	}

	// Set version (4) and variant bits
	b[6] = (b[6] & 0x0f) | 0x40 // Version 4 (bits 12–15 of byte 6)
	b[8] = (b[8] & 0x3f) | 0x80 // Variant bits 10xx (bits 6–7 of byte 8)

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4],
		b[4:6],
		b[6:8],
		b[8:10],
		b[10:16],
	)
}
