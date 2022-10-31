// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

//go:build linux
// +build linux

package process_test

import (
	"testing"

	"github.com/instana/go-sensor/process"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestStats_CPU(t *testing.T) {
	rdr := process.Stats()
	rdr.ProcPath = "testdata"
	rdr.Command = "Hello, brave new world"

	stats, tick, err := rdr.CPU()
	require.NoError(t, err)
	assert.Equal(t, 71695348, tick)
	assert.Equal(t, process.CPUStats{
		User:   14,
		System: 15,
	}, stats)
}

func TestStats_Limits(t *testing.T) {
	rdr := process.Stats()
	rdr.ProcPath = "testdata"

	limits, err := rdr.Limits()
	require.NoError(t, err)
	assert.Equal(t, process.ResourceLimits{
		OpenFiles: process.LimitedResource{
			Current: 4,
			Max:     1024,
		},
	}, limits)
}
