// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instana_test

import (
	"testing"

	instana "github.com/instana/go-sensor"
	"github.com/stretchr/testify/assert"
)

func TestDefaultTracerOptions(t *testing.T) {
	opts := instana.DefaultTracerOptions()
	assert.Equal(t, 2, opts.MaxLogsPerSpan)
	assert.Equal(t, instana.DefaultSecretsMatcher(), opts.Secrets)
	assert.False(t, opts.DropAllLogs)
	assert.Nil(t, opts.CollectableHTTPHeaders)
	assert.Nil(t, opts.DisableSpans)
	assert.False(t, opts.HTTP.Exit.ClassifyAll4xxAsErrors)
	assert.Nil(t, opts.HTTP.Exit.ClassifyAsErrors)
}
