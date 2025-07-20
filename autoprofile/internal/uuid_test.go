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
	assert.NotEqual(t, internal.GenerateUUID(nil), internal.GenerateUUID(nil))
}

func TestGenerateUUID_PrimaryMethod(t *testing.T) {
	uuid := internal.GenerateUUID(nil)
	assert.Equal(t, 40, len(uuid))
}

func TestGenerateUUID_FallbackMethod(t *testing.T) {
	var ir InvalidReader
	uuid := internal.GenerateUUID(ir)
	assert.Equal(t, 40, len(uuid))
}

func BenchmarkGenerateUUID_PrimaryMethod(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = internal.GenerateUUID(nil)
	}
}

func BenchmarkGenerateUUID_FallbackMethod(b *testing.B) {
	var ir InvalidReader
	for i := 0; i < b.N; i++ {
		_ = internal.GenerateUUID(ir)
	}
}

type InvalidReader struct{}

func (r InvalidReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("failed to generate random values")
}
