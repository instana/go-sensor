package autoprofile

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var objs []string

func TestCreateAllocationCallGraph(t *testing.T) {
	opts := DefaultOptions()
	opts.IncludeSensorFrames = true
	SetOptions(opts)

	objs = make([]string, 1000000)
	defer func() { objs = nil }()

	runtime.GC()
	runtime.GC()

	samp := newAllocationSampler()

	p, err := samp.readHeapProfile()
	require.NoError(t, err)

	callGraph, err := samp.createAllocationCallGraph(p)
	require.NoError(t, err)
	//fmt.Printf("CALL GRAPH: %v\n", callGraph.printLevel(0))

	assert.Contains(t, fmt.Sprintf("%v", callGraph.toMap()), "TestCreateAllocationCallGraph")
}
