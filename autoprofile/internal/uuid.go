// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package internal

import (
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"sync/atomic"
	"time"
)

var (
	nextID atomic.Int64
)

// GenerateUUID generates a UUID of length 40 characters
func GenerateUUID() string {
	return SecureUUID(crand.Reader)
}

func SecureUUID(r io.Reader) string {
	const byteLength = 20
	uuidBytes := make([]byte, 20)

	if r == nil {
		r = crand.Reader
	}

	if _, err := io.ReadFull(r, uuidBytes); err != nil {
		//fallback mechanism if crypto/rand fails to generate random data
		now := time.Now().UnixNano()
		next := nextID.Add(1)

		data := fmt.Sprintf("%d%d%d", now, os.Getpid(), next)
		hash := sha256.Sum256([]byte(data))

		copy(uuidBytes, hash[:byteLength])
	}

	return hex.EncodeToString(uuidBytes)
}
