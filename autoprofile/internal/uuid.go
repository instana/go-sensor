package internal

import (
	"crypto/sha1"
	"encoding/hex"
	"math/rand"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var (
	randSource = rand.New(rand.NewSource(time.Now().UnixNano()))
	randLock   sync.Mutex
	nextID     int64
)

func GenerateUUID() string {
	n := atomic.AddInt64(&nextID, 1)

	uuid := strconv.FormatInt(time.Now().Unix(), 10) +
		strconv.FormatInt(random(1000000000), 10) +
		strconv.FormatInt(n, 10)

	return sha1String(uuid)
}

func random(max int64) int64 {
	randLock.Lock()
	defer randLock.Unlock()

	return randSource.Int63n(max)
}

func sha1String(s string) string {
	sha1 := sha1.New()
	sha1.Write([]byte(s))

	return hex.EncodeToString(sha1.Sum(nil))
}
