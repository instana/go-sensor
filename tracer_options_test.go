package instana_test

import (
	"testing"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/secrets"
	"github.com/instana/testify/assert"
)

func TestDefaultTracerOptions(t *testing.T) {
	assert.Equal(t, instana.TracerOptions{
		MaxLogsPerSpan: 2,
		Secrets:        secrets.NoneMatcher{},
	}, instana.DefaultTracerOptions())
}
