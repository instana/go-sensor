package autoprofile

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
)

var objs []string

func TestCreateAllocationCallGraph(t *testing.T) {
	profiler := newAutoProfiler()
	profiler.IncludeSensorFrames = true

	objs = make([]string, 0)
	for i := 0; i < 1000000; i++ {
		objs = append(objs, string(i))
	}

	runtime.GC()
	runtime.GC()

	allocationSensor := newAllocationSampler(profiler)

	p, _ := allocationSensor.readHeapProfile()

	// size
	callGraph, err := allocationSensor.createAllocationCallGraph(p)
	if err != nil {
		t.Error(err)
		return
	}
	//fmt.Printf("CALL GRAPH: %v\n", callGraph.printLevel(0))

	if !strings.Contains(fmt.Sprintf("%v", callGraph.toMap()), "TestCreateAllocationCallGraph") {
		t.Error("The test function is not found in the profile")
	}

	objs = nil
}
