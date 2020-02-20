package autoprofile

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestCreateBlockProfile(t *testing.T) {
	profiler := newAutoProfiler()
	profiler.IncludeSensorFrames = true

	done := make(chan bool)

	go func() {
		time.Sleep(200 * time.Millisecond)

		wait := make(chan bool)

		go func() {
			time.Sleep(150 * time.Millisecond)

			wait <- true
		}()

		<-wait

		done <- true
	}()

	blockSampler := newBlockSampler(profiler)
	blockSampler.resetSampler()
	blockSampler.startSampler()
	time.Sleep(500 * time.Millisecond)
	blockSampler.stopSampler()
	profile, _ := blockSampler.buildProfile(500*1e6, 120)
	//fmt.Printf("CALL GRAPH: %v\n", profile.toIndentedJson())

	if !strings.Contains(fmt.Sprintf("%v", profile.toMap()), "TestCreateBlockProfile") {
		t.Error("The test function is not found in the profile")
	}

	<-done
}

func waitForServer(url string) {
	for {
		if _, err := http.Get(url); err == nil {
			time.Sleep(100 * time.Millisecond)
			break
		}
	}
}
