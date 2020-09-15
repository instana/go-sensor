// +build linux

package process

import (
	"fmt"
	"os"
	"path"
)

const (
	pageSize = 4 << 10 // standard setting, applicable for most systems
	procPath = "/proc/self"
)

type statsReader struct {
	ProcPath string
	Command  string
}

// Stats returns a process resource stats reader for current process
func Stats() statsReader {
	return statsReader{
		ProcPath: procPath,
		Command:  path.Base(os.Args[0]),
	}
}

// Memory returns memory stats for current process
func (rdr statsReader) Memory() (MemStats, error) {
	fd, err := os.Open(rdr.ProcPath + "/statm")
	if err != nil {
		return MemStats{}, nil
	}
	defer fd.Close()

	var total, rss, shared int

	// The fields come in order described in `/proc/[pid]/statm` section
	// of https://man7.org/linux/man-pages/man5/proc.5.html
	if _, err := fmt.Fscanf(fd, "%d %d %d",
		&total,  // size
		&rss,    // resident
		&shared, // shared
		// ... the rest of the fields are not used and thus omitted
	); err != nil {
		return MemStats{}, fmt.Errorf("failed to parse %s: %s", fd.Name(), err)
	}

	return MemStats{
		Total:  total * pageSize,
		Rss:    rss * pageSize,
		Shared: shared * pageSize,
	}, nil
}
