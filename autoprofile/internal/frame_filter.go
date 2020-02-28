package internal

import (
	"path/filepath"
	"strings"

	profile "github.com/instana/go-sensor/autoprofile/internal/pprof/profile"
)

var (
	IncludeSensorFrames = true
	sensorPath          = filepath.Join("github.com", "instana", "go-sensor")
)

func shouldSkipStack(sample *profile.Sample) bool {
	return !IncludeSensorFrames && isSensorStack(sample)
}

func shouldSkipFrame(fileName, funcName string) bool {
	return (!IncludeSensorFrames && isSensorFrame(fileName)) || funcName == "runtime.goexit"
}

func isSensorStack(sample *profile.Sample) bool {
	return stackContains(sample, "", sensorPath)
}

func isSensorFrame(fileNameTest string) bool {
	return strings.Contains(fileNameTest, sensorPath)
}

func stackContains(sample *profile.Sample, funcNameTest string, fileNameTest string) bool {
	for i := len(sample.Location) - 1; i >= 0; i-- {
		l := sample.Location[i]
		funcName, fileName, _ := readFuncInfo(l)

		if (funcNameTest == "" || strings.Contains(funcName, funcNameTest)) &&
			(fileNameTest == "" || strings.Contains(fileName, fileNameTest)) {
			return true
		}
	}

	return false
}
