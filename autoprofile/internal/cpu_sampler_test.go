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

func TestCreateCPUProfile(t *testing.T) {
	cpuSampler := internal.NewCPUSampler()
	internal.IncludeProfilerFrames = true

	cpuSampler.Reset()
	cpuSampler.Start()

	simulateCPULoad(1 * time.Second)

	cpuSampler.Stop()

	profile, err := cpuSampler.Profile(500*1e6, 120)
	require.NoError(t, err)

	assert.Contains(t, fmt.Sprintf("%v", internal.NewAgentProfile(profile)), "simulateCPULoad")
}

func simulateCPULoad(d time.Duration) {
	done := time.After(d)

	for {
		select {
		case <-done:
			return
		default: //nolint:staticcheck
		}
	}
}
