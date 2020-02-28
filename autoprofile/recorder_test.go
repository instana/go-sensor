package autoprofile_test

import (
	"errors"
	"testing"
	"time"

	"github.com/instana/go-sensor/autoprofile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecorder_Flush(t *testing.T) {
	profilesChan := make(chan interface{})

	rec := autoprofile.NewRecorder()
	rec.SendProfiles = func(profiles interface{}) error {
		profilesChan <- profiles
		return nil
	}

	profile := map[string]interface{}{
		"a": 1,
	}
	rec.Record(profile)

	profile = map[string]interface{}{
		"a": 2,
	}
	rec.Record(profile)

	go rec.Flush()

	select {
	case profiles := <-profilesChan:
		assert.Equal(t, 0, rec.Size())

		require.IsType(t, profiles, []interface{}{})
		assert.Len(t, profiles.([]interface{}), 2)
	case <-time.After(2 * time.Second):
		t.Errorf("(*autoprofile.ProfileRecorder).Flush() did not return within 2 seconds")
	}
}

func TestRecorder_Flush_Fail(t *testing.T) {
	rec := autoprofile.NewRecorder()
	rec.SendProfiles = func(profiles interface{}) error {
		return errors.New("some error")
	}

	profile := map[string]interface{}{
		"a": 1,
	}
	rec.Record(profile)

	profile = map[string]interface{}{
		"a": 2,
	}

	rec.Record(profile)
	rec.Flush()

	assert.Equal(t, 2, rec.Size())
}
