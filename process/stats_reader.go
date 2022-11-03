// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

//go:build !linux
// +build !linux

package process

type statsReader struct{}

// Stats returns a process resource stats reader for current process
func Stats() statsReader {
	return statsReader{}
}

// Memory returns memory stats for current process
func (statsReader) Memory() (MemStats, error) {
	return MemStats{}, nil
}

// CPU returns CPU stats for current process and the CPU tick they were taken on
func (statsReader) CPU() (CPUStats, int, error) {
	return CPUStats{}, 0, nil
}

// Limits returns resource limits configured for current process
func (statsReader) Limits() (ResourceLimits, error) {
	return ResourceLimits{}, nil
}
