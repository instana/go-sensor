// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

//go:build linux
// +build linux

package process

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

const (
	pageSize = 4 << 10 // standard setting, applicable for most systems
	procPath = "/proc"
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
	fd, err := os.Open(rdr.ProcPath + "/self/statm")
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

// CPU returns CPU stats for current process and the CPU tick they were taken on
func (rdr statsReader) CPU() (CPUStats, int, error) {
	fd, err := os.Open(rdr.ProcPath + "/self/stat")
	if err != nil {
		return CPUStats{}, 0, nil
	}
	defer fd.Close()

	var (
		stats   CPUStats
		skipInt int
		skipCh  byte
	)

	// The command in `/proc/self/stat` output is truncated to 15 bytes (16 including the terminating null byte)
	comm := rdr.Command
	if len(comm) > 15 {
		comm = comm[:15]
	}

	// The fields come in order described in `/proc/[pid]/stat` section
	// of https://man7.org/linux/man-pages/man5/proc.5.html. We skip parsing
	// the `comm` field since it may contain space characters that break fmt.Fscanf format.
	if _, err := fmt.Fscanf(fd, "%d ("+comm+") %c %d %d %d %d %d %d %d %d %d %d %d %d",
		&skipInt,      // pid
		&skipCh,       // state
		&skipInt,      // ppid
		&skipInt,      // pgrp
		&skipInt,      // session
		&skipInt,      // tty_nr
		&skipInt,      // tpgid
		&skipInt,      // flags
		&skipInt,      // minflt
		&skipInt,      // cminflt
		&skipInt,      // majflt
		&skipInt,      // cmajflt
		&stats.User,   // utime
		&stats.System, // stime
		// ... the rest of the fields are not used and thus omitted
	); err != nil {
		return stats, 0, fmt.Errorf("failed to parse %s: %s", fd.Name(), err)
	}

	tick, err := rdr.currentTick()
	if err != nil {
		return stats, 0, fmt.Errorf("failed to get current CPU tick: %s", err)
	}

	return stats, tick, nil
}

// currentTick parses /proc/stat, sums up the total number of ticks spent on each CPU and averages them
// by the number of CPUs
func (rdr statsReader) currentTick() (int, error) {
	fd, err := os.Open(rdr.ProcPath + "/stat")
	if err != nil {
		return 0, nil
	}
	defer fd.Close()

	sc := bufio.NewScanner(fd)
	sc.Split(bufio.ScanLines)

	var (
		ticks, cpuCount                                    int
		user, nice, sys, idle, iowait, irq, softIRQ, steal int
		skipStr                                            string
	)

	for sc.Scan() {
		s := sc.Text()
		if !strings.HasPrefix(s, "cpu") {
			continue
		}

		if strings.HasPrefix(s, "cpu ") { // skip total CPU line
			continue
		}

		// The fields come in order described in `/proc/stat` section
		// of https://man7.org/linux/man-pages/man5/proc.5.html
		if _, err := fmt.Sscanf(s, "%s %d %d %d %d %d %d %d %d",
			&skipStr, // CPU label
			&user,
			&nice,
			&sys,
			&idle,
			&iowait,
			&irq,
			&softIRQ,
			&steal,
			// ... the rest of the fields are not used and thus omitted
		); err != nil {
			return 0, fmt.Errorf("failed to parse %s: %s", fd.Name(), err)
		}

		ticks += user + nice + sys + idle + iowait + irq + softIRQ + steal
		cpuCount++
	}

	if err := sc.Err(); err != nil {
		return 0, fmt.Errorf("failed to read %s: %s", fd.Name(), err)
	}

	if cpuCount < 2 {
		return ticks, nil
	}

	return ticks / cpuCount, nil
}

// Limits returns resource limits configured for current process
func (rdr statsReader) Limits() (ResourceLimits, error) {
	fd, err := os.Open(rdr.ProcPath + "/self/limits")
	if err != nil {
		return ResourceLimits{}, nil
	}
	defer fd.Close()

	sc := bufio.NewScanner(fd)
	sc.Split(bufio.ScanLines)

	var limits ResourceLimits

	for sc.Scan() {
		s := sc.Text()
		if !strings.HasPrefix(s, "Max open files") {
			continue
		}

		s = strings.TrimLeft(s[14:], " \t") // trim the "max open files" prefix along with trailing space
		if !strings.HasPrefix(s, "unlimited") {
			if _, err := fmt.Sscanf(s, "%d", &limits.OpenFiles.Max); err != nil {
				return limits, fmt.Errorf("unexpected %s format: %s", fd.Name(), err)
			}
		}

		break
	}

	if err := sc.Err(); err != nil {
		return limits, fmt.Errorf("failed to read %s: %s", fd.Name(), err)
	}

	fdNum, err := rdr.currentOpenFiles()
	if err != nil {
		return limits, fmt.Errorf("failed to get the number of open files: %s", err)
	}

	limits.OpenFiles.Current = fdNum

	return limits, nil
}

func (rdr statsReader) currentOpenFiles() (int, error) {
	fds, err := ioutil.ReadDir(rdr.ProcPath + "/self/fd/")
	if err != nil {
		return 0, fmt.Errorf("failed to list %s: %s", rdr.ProcPath+"/fd/", err)
	}

	return len(fds), nil
}
