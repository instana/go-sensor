package autoprofile_test

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/instana/go-sensor/autoprofile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var objs []string

func TestCreateAllocationCallGraph(t *testing.T) {
	opts := autoprofile.DefaultOptions()
	opts.IncludeSensorFrames = true
	autoprofile.SetOptions(opts)

	objs = make([]string, 1000000)
	defer func() { objs = nil }()

	runtime.GC()
	runtime.GC()

	samp := autoprofile.NewAllocationSampler()

	p, err := samp.Profile(500*1e6, 120)
	require.NoError(t, err)

	assert.Contains(t, fmt.Sprintf("%v", p.ToMap()), "TestCreateAllocationCallGraph")
}
