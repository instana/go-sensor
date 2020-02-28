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

	cpuSampler.Reset()
	cpuSampler.Start()

	simulateCPULoad(1 * time.Second)

	cpuSampler.Stop()

	profile, err := cpuSampler.Profile(500*1e6, 120)
	require.NoError(t, err)

	assert.Contains(t, fmt.Sprintf("%v", profile.ToMap()), "simulateCPULoad")
}

func simulateCPULoad(d time.Duration) {
	done := time.After(d)

	for {
		select {
		case <-done:
			return
		default:
		}
	}
}
