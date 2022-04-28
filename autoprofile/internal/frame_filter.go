// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package internal

import (
	"path/filepath"
	"strings"

	"github.com/instana/go-sensor/autoprofile/internal/pprof/profile"
)

var (
	// IncludeProfilerFrames is a setting for the frame filter whether or not to include the profiler
	// frames into the profile
	IncludeProfilerFrames = false
	autoprofilePath       = filepath.Join("github.com", "instana", "go-sensor", "autoprofile")
)

func shouldSkipStack(sample *profile.Sample) bool {
	return !IncludeProfilerFrames && stackContains(sample, autoprofilePath)
}

func stackContains(sample *profile.Sample, fileNameTest string) bool {
	for i := len(sample.Location) - 1; i >= 0; i-- {
		l := sample.Location[i]
		_, fileName, _ := readFuncInfo(l)

		if strings.Contains(fileName, fileNameTest) {
			return true
		}
	}

	return false
}
