package autoprofile

import (
	"errors"
	"testing"
)

func TestFlush(t *testing.T) {
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

	profiles := <-profilesChan

	if len(profiler.profileRecorder.queue) > 0 {
		t.Errorf("Queue should be empty, but have %v profiles", len(profiler.profileRecorder.queue))
	}

	if len(profiles.([]interface{})) < 2 {
		t.Errorf("Received %v profiles", len(profiles.([]interface{})))
	}
}

func TestFlushFail(t *testing.T) {
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

	if len(profiler.profileRecorder.queue) < 2 {
		t.Errorf("Queue contains %v profiles", len(profiler.profileRecorder.queue))
	}
}
