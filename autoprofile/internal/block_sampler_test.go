// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package internal_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/instana/go-sensor/autoprofile/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateBlockProfile(t *testing.T) {
	blockSampler := internal.NewBlockSampler()
	internal.IncludeProfilerFrames = true

	blockSampler.Reset()
	blockSampler.Start()

	simulateBlocking(150 * time.Millisecond)

	blockSampler.Stop()

	profile, err := blockSampler.Profile(500*1e6, 120)
	require.NoError(t, err)

	assert.Contains(t, fmt.Sprintf("%v", internal.NewAgentProfile(profile)), "simulateBlocking")
}

func simulateBlocking(d time.Duration) {
	wait := make(chan struct{})

	time.AfterFunc(d, func() { wait <- struct{}{} })
	<-wait
}
