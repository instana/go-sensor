package autoprofile

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestCreateCPUProfile(t *testing.T) {
	profiler := NewAutoProfiler()
	profiler.IncludeSensorFrames = true

	done := make(chan bool)

	go func() {
		// cpu
		//start := time.Now().UnixNano()
		for i := 0; i < 10000000; i++ {
			str := "str" + strconv.Itoa(i)
			str = str + "a"
		}
		//took := time.Now().UnixNano() - start
		//fmt.Printf("TOOK: %v\n", took)

		done <- true
	}()

	cpuSampler := newCPUSampler(profiler)
	cpuSampler.resetSampler()
	cpuSampler.startSampler()

	time.Sleep(500 * time.Millisecond)
	cpuSampler.stopSampler()
	profile, _ := cpuSampler.buildProfile(500*1e6, 120)

	//fmt.Printf("CALL GRAPH: %v\n", profile.toIndentedJson())

	if !strings.Contains(fmt.Sprintf("%v", profile.toMap()), "TestCreateCPUProfile") {
		t.Error("The test function is not found in the profile")
	}

	<-done
}
