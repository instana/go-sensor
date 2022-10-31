// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package internal_test

import (
	"errors"
	"testing"

	"github.com/instana/go-sensor/autoprofile/internal"
	"github.com/stretchr/testify/assert"
)

func TestRecorder_Flush(t *testing.T) {
	var profiles []internal.AgentProfile

	rec := internal.NewRecorder()
	rec.SendProfiles = func(p []internal.AgentProfile) error {
		profiles = p
		return nil
	}

	rec.Record(internal.AgentProfile{ID: "1"})
	rec.Record(internal.AgentProfile{ID: "2"})

	rec.Flush()

	assert.Len(t, profiles, 2)

	assert.Equal(t, 0, rec.Size())
}

func TestRecorder_Flush_Fail(t *testing.T) {
	rec := internal.NewRecorder()
	rec.SendProfiles = func(profiles []internal.AgentProfile) error {
		return errors.New("some error")
	}

	rec.Record(internal.AgentProfile{ID: "1"})
	rec.Record(internal.AgentProfile{ID: "2"})

	rec.Flush()

	assert.Equal(t, 2, rec.Size())
}
