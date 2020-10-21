package acceptor_test

import (
	"testing"

	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/testify/assert"
)

func TestNewGoProcessPluginPayload(t *testing.T) {
	data := acceptor.GoProcessData{
		PID: 42,
	}

	assert.Equal(t, acceptor.PluginPayload{
		Name:     "com.instana.plugin.golang",
		EntityID: "42",
		Data:     data,
	}, acceptor.NewGoProcessPluginPayload(data))
}
