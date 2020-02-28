package autoprofile_test

import (
	"testing"

	"github.com/instana/go-sensor/autoprofile"
	"github.com/stretchr/testify/assert"
)

func TestGenerateUUID_Unique(t *testing.T) {
	assert.NotEqual(t, autoprofile.GenerateUUID(), autoprofile.GenerateUUID())
}
