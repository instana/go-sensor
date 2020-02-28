package autoprofile_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/instana/go-sensor/autoprofile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateBlockProfile(t *testing.T) {
	opts := autoprofile.DefaultOptions()
	opts.IncludeSensorFrames = true
	autoprofile.SetOptions(opts)

	blockSampler := autoprofile.NewBlockSampler()

	blockSampler.Reset()
	blockSampler.Start()

	simulateBlocking(150 * time.Millisecond)

	blockSampler.Stop()

	profile, err := blockSampler.Profile(500*1e6, 120)
	require.NoError(t, err)

	assert.Contains(t, fmt.Sprintf("%v", profile.ToMap()), "simulateBlocking")
}

func simulateBlocking(d time.Duration) {
	wait := make(chan struct{})

	time.AfterFunc(d, func() { wait <- struct{}{} })
	<-wait
}
