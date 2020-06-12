package internal_test

import (
	"errors"
	"testing"

	"github.com/instana/go-sensor/autoprofile/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecorder_Flush(t *testing.T) {
	var profiles interface{}

	rec := internal.NewRecorder()
	rec.SendProfiles = func(p interface{}) error {
		profiles = p
		return nil
	}

	rec.Record(internal.AgentProfile{ID: "1"})
	rec.Record(internal.AgentProfile{ID: "2"})

	rec.Flush()

	require.IsType(t, []interface{}{}, profiles)
	assert.Len(t, profiles.([]interface{}), 2)

	assert.Equal(t, 0, rec.Size())
}

func TestRecorder_Flush_Fail(t *testing.T) {
	rec := internal.NewRecorder()
	rec.SendProfiles = func(profiles interface{}) error {
		return errors.New("some error")
	}

	rec.Record(internal.AgentProfile{ID: "1"})
	rec.Record(internal.AgentProfile{ID: "2"})

	rec.Flush()

	assert.Equal(t, 2, rec.Size())
}
