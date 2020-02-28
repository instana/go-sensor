package internal_test

import (
	"testing"

	"github.com/instana/go-sensor/autoprofile/internal"
	"github.com/stretchr/testify/assert"
)

func TestGenerateUUID_Unique(t *testing.T) {
	assert.NotEqual(t, internal.GenerateUUID(), internal.GenerateUUID())
}
