// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package internal_test

import (
	"errors"
	"testing"

	"github.com/instana/go-sensor/autoprofile/internal"
	"github.com/stretchr/testify/assert"
)

func TestGenerateUUID_Unique(t *testing.T) {
	assert.NotEqual(t, internal.GenerateUUID(), internal.GenerateUUID())
}

func TestGenerateUUID_PrimaryMethod(t *testing.T) {
	uuid := internal.SecureUUID(nil)
	assert.Equal(t, 40, len(uuid))
}

func TestGenerateUUID_FallbackMethod(t *testing.T) {
	var ir InvalidReader
	uuid := internal.SecureUUID(ir)
	assert.Equal(t, 40, len(uuid))
}

func BenchmarkGenerateUUID_PrimaryMethod(b *testing.B) {
	for range b.N {
		_ = internal.SecureUUID(nil)
	}
}

func BenchmarkGenerateUUID_FallbackMethod(b *testing.B) {
	var ir InvalidReader
	for range b.N {
		_ = internal.SecureUUID(ir)
	}
}

type InvalidReader struct{}

func (r InvalidReader) Read(_ []byte) (n int, err error) {
	return 0, errors.New("failed to generate random values")
}
