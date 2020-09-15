// +build linux

package process_test

import (
	"testing"

	"github.com/instana/go-sensor/process"
	"github.com/instana/testify/assert"
	"github.com/instana/testify/require"
)

func TestStats_Memory(t *testing.T) {
	rdr := process.Stats()
	rdr.ProcPath = "testdata"

	stats, err := rdr.Memory()
	require.NoError(t, err)
	assert.Equal(t, process.MemStats{
		Total:  1 * 4 << 10,
		Rss:    2 * 4 << 10,
		Shared: 3 * 4 << 10,
	}, stats)
}
