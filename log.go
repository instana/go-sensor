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

var defaultLogger = logger.New(log.New(os.Stderr, "", log.LstdFlags))

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
