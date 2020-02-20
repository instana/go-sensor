package autoprofile

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProfileRecorder_Flush(t *testing.T) {
	profilesChan := make(chan interface{})

	profiler := newAutoProfiler()
	profiler.IncludeSensorFrames = true
	profiler.SendProfiles = func(profiles interface{}) error {
		profilesChan <- profiles
		return nil
	}

	profile := map[string]interface{}{
		"a": 1,
	}
	profiler.profileRecorder.record(profile)

	profile = map[string]interface{}{
		"a": 2,
	}
	profiler.profileRecorder.record(profile)

	go profiler.profileRecorder.flush()

	select {
	case profiles := <-profilesChan:
		assert.Empty(t, profiler.profileRecorder.queue)

		require.IsType(t, profiles, []interface{}{})
		assert.Len(t, profiles.([]interface{}), 2)
	case <-time.After(2 * time.Second):
		t.Errorf("(*autoprofile.ProfileRecorder).flush() did not return within 2 seconds")
	}
}

func TestProfileRecorder_Flush_Fail(t *testing.T) {
	profiler := newAutoProfiler()
	profiler.IncludeSensorFrames = true
	profiler.SendProfiles = func(profiles interface{}) error {
		return errors.New("some error")
	}

	profile := map[string]interface{}{
		"a": 1,
	}
	profiler.profileRecorder.record(profile)

	profile = map[string]interface{}{
		"a": 2,
	}

	profiler.profileRecorder.record(profile)
	profiler.profileRecorder.flush()

	assert.Len(t, profiler.profileRecorder.queue, 2)
}
