package instana

import (
	l "log"
)

const (
	ERROR = 0
	WARN  = 1
	INFO  = 2
	DEBUG = 3
)

type logS struct {
	sensor *sensorS
}

var log *logS

func (r *logS) makeV(prefix string, v ...interface{}) []interface{} {
	return append([]interface{}{prefix}, v...)
}

func (r *logS) debug(v ...interface{}) {
	if r.sensor.options.LogLevel >= DEBUG {
		l.Println(r.makeV("DEBUG: instana:", v...)...)
	}
}

func (r *logS) info(v ...interface{}) {
	if r.sensor.options.LogLevel >= INFO {
		l.Println(r.makeV("INFO: instana:", v...)...)
	}
}

func (r *logS) warn(v ...interface{}) {
	if r.sensor.options.LogLevel >= WARN {
		l.Println(r.makeV("WARN: instana:", v...)...)
	}
}

func (r *logS) error(v ...interface{}) {
	if r.sensor.options.LogLevel >= ERROR {
		l.Println(r.makeV("ERROR: instana:", v...)...)
	}
}

func (r *sensorS) initLog() {
	log = new(logS)
	log.sensor = r
}
