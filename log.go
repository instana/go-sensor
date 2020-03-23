package instana

import (
	"log"
	"os"

	"github.com/instana/go-sensor/logger"
)

// Valid log levels
const (
	Error = 0
	Warn  = 1
	Info  = 2
	Debug = 3
)

type logS struct {
	sensor *sensorS
}

var (
	instanaLog    *logS
	defaultLogger = logger.New(log.New(os.Stderr, "", log.LstdFlags))
)

func (r *logS) makeV(prefix string, v ...interface{}) []interface{} {
	return append([]interface{}{prefix}, v...)
}

func (r *logS) debug(v ...interface{}) {
	if r.sensor.options.LogLevel >= Debug {
		log.Println(r.makeV("DEBUG: instana:", v...)...)
	}
}

func (r *logS) info(v ...interface{}) {
	if r.sensor.options.LogLevel >= Info {
		log.Println(r.makeV("INFO: instana:", v...)...)
	}
}

func (r *logS) warn(v ...interface{}) {
	if r.sensor.options.LogLevel >= Warn {
		log.Println(r.makeV("WARN: instana:", v...)...)
	}
}

func (r *logS) error(v ...interface{}) {
	if r.sensor.options.LogLevel >= Error {
		log.Println(r.makeV("ERROR: instana:", v...)...)
	}
}

func (r *sensorS) initLog() {
	instanaLog = new(logS)
	instanaLog.sensor = r
}

// setLogLevel translates legacy Instana log levels into logger.Logger levels.
// Any level that is greater than instana.Debug is interpreted as logger.DebugLevel.
func setLogLevel(l *logger.Logger, level int) {
	switch level {
	case Error:
		l.SetLevel(logger.ErrorLevel)
	case Warn:
		l.SetLevel(logger.WarnLevel)
	case Info:
		l.SetLevel(logger.InfoLevel)
	default:
		l.SetLevel(logger.DebugLevel)
	}
}
