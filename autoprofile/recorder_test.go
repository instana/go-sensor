package autoprofile

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecorder_Flush(t *testing.T) {
	profilesChan := make(chan interface{})

	rec := newRecorder()
	rec.SendProfiles = func(profiles interface{}) error {
		profilesChan <- profiles
		return nil
	}

	profile := map[string]interface{}{
		"a": 1,
	}
	rec.record(profile)

	profile = map[string]interface{}{
		"a": 2,
	}
	rec.record(profile)

	go rec.flush()

	select {
	case profiles := <-profilesChan:
		assert.Empty(t, profiler.profileRecorder.queue)

		require.IsType(t, profiles, []interface{}{})
		assert.Len(t, profiles.([]interface{}), 2)
	case <-time.After(2 * time.Second):
		t.Errorf("(*autoprofile.ProfileRecorder).flush() did not return within 2 seconds")
	}
}

func TestRecorder_Flush_Fail(t *testing.T) {
	rec := newRecorder()
	rec.SendProfiles = func(profiles interface{}) error {
		return errors.New("some error")
	}

	profile := map[string]interface{}{
		"a": 1,
	}
	rec.record(profile)

	profile = map[string]interface{}{
		"a": 2,
	}

	rec.record(profile)
	rec.flush()

	assert.Len(t, rec.queue, 2)
}
