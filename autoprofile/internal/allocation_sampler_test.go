package internal_test

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/instana/go-sensor/autoprofile/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var objs []string

func TestCreateAllocationCallGraph(t *testing.T) {
	objs = make([]string, 1000000)
	defer func() { objs = nil }()

	runtime.GC()
	runtime.GC()

	samp := internal.NewAllocationSampler()

	p, err := samp.Profile(500*1e6, 120)
	require.NoError(t, err)

	assert.Contains(t, fmt.Sprintf("%v", p.ToMap()), "TestCreateAllocationCallGraph")
}
