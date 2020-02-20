package autoprofile

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateCPUProfile(t *testing.T) {
	profiler := newAutoProfiler()
	profiler.IncludeSensorFrames = true

	go func() {
		done := time.After(1 * time.Second)

		var i int
		for {
			i++

			select {
			case <-done:
				return
			default:
				str := "str" + strconv.Itoa(i)
				str = str + "a"
			}
		}
	}()

	cpuSampler := newCPUSampler(profiler)

	cpuSampler.resetSampler()
	cpuSampler.startSampler()

	time.Sleep(500 * time.Millisecond)
	cpuSampler.stopSampler()

	profile, err := cpuSampler.buildProfile(500*1e6, 120)
	require.NoError(t, err)

	assert.Contains(t, fmt.Sprintf("%v", profile.toMap()), "TestCreateCPUProfile")
}
