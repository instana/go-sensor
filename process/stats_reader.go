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

// CPU returns CPU stats for current process
func (statsReader) CPU() (CPUStats, error) {
	return CPUStats{}, nil
}
