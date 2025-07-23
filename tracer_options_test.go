// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instana_test

import (
	"testing"

	instana "github.com/instana/go-sensor"
	"github.com/stretchr/testify/assert"
)

func TestDefaultTracerOptions(t *testing.T) {
	assert.Equal(t, instana.TracerOptions{
		MaxLogsPerSpan: 2,
		Secrets:        instana.DefaultSecretsMatcher(),
	}, instana.DefaultTracerOptions())
}
