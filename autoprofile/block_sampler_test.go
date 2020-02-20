package autoprofile

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateBlockProfile(t *testing.T) {
	profiler := newAutoProfiler()
	profiler.IncludeSensorFrames = true

	ready := make(chan struct{})
	go func() {
		wait := make(chan struct{})

		<-ready
		go func() {
			time.Sleep(150 * time.Millisecond)

			wait <- struct{}{}
		}()

		<-wait
	}()

	blockSampler := newBlockSampler(profiler)

	blockSampler.resetSampler()
	blockSampler.startSampler()

	ready <- struct{}{}

	time.Sleep(500 * time.Millisecond)

	blockSampler.stopSampler()

	profile, err := blockSampler.buildProfile(500*1e6, 120)
	require.NoError(t, err)

	assert.Contains(t, fmt.Sprintf("%v", profile.toMap()), "TestCreateBlockProfile")
}
