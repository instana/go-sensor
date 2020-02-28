package autoprofile_test

import (
	"testing"

	"github.com/instana/go-sensor/autoprofile"
	"github.com/stretchr/testify/assert"
)

func TestCallSite_Increment(t *testing.T) {
	root := autoprofile.NewCallSite("root", "", 0)

	root.Increment(12.3, 1)
	root.Increment(0, 0)
	root.Increment(5, 2)

	m, ns := root.Measurement()
	assert.Equal(t, 17.3, m)
	assert.EqualValues(t, 3, ns)
}
