package autoprofile

import (
	l "log"
)

const (
	Error = 0
	Warn  = 1
	Info  = 2
	Debug = 3
)

var log *LogWrapper = &LogWrapper{Debug}

type LogWrapper struct {
	logLevel int
}

func (lw *LogWrapper) makeV(prefix string, v ...interface{}) []interface{} {
	return append([]interface{}{prefix}, v...)
}

func (lw *LogWrapper) debug(v ...interface{}) {
	if lw.logLevel >= Debug {
		l.Println(lw.makeV("DEBUG: instana:", v...)...)
	}
}

func (lw *LogWrapper) info(v ...interface{}) {
	if lw.logLevel >= Info {
		l.Println(lw.makeV("INFO: instana:", v...)...)
	}
}

func (lw *LogWrapper) warn(v ...interface{}) {
	if lw.logLevel >= Warn {
		l.Println(lw.makeV("WARN: instana:", v...)...)
	}
}

func (lw *LogWrapper) error(v ...interface{}) {
	if lw.logLevel >= Error {
		l.Println(lw.makeV("ERROR: instana:", v...)...)
	}
}
