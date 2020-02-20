package autoprofile

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateUUID_Unique(t *testing.T) {
	assert.NotEqual(t, generateUUID(), generateUUID())
}
