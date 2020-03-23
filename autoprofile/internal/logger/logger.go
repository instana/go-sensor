package logger

import (
	instalogger "github.com/instana/go-sensor/logger"
)

// Level is the log level
type Level uint8

// Valid log levels compatible with github.com/instana/go-sensor log level values
const (
	ErrorLevel Level = iota
	WarnLevel
	InfoLevel
	DebugLevel
)

var defaultLogger LeveledLogger = instalogger.New(nil)

// LeveledLogger is an interface of a generic logger that support different message levels.
// By default instana.Sensor uses logger.Logger with log.Logger as an output, however this
// interface is also compatible with such popular loggers as github.com/sirupsen/logrus.Logger
// and go.uber.org/zap.SugaredLogger
type LeveledLogger interface {
	Debug(v ...interface{})
	Info(v ...interface{})
	Warn(v ...interface{})
	Error(v ...interface{})
}

// SetLogLevel configures the min log level for autoprofile defaultLogger
func SetLogLevel(level instalogger.Level) {
	l, ok := defaultLogger.(*instalogger.Logger)
	if !ok {
		return
	}

	l.SetLevel(level)
}

// SetLogger sets the leveled logger to use to output the diagnostic messages and errors
func SetLogger(l LeveledLogger) {
	defaultLogger = l
}

// Debug writes log message with defaultLogger.Debug level
func Debug(v ...interface{}) {
	defaultLogger.Debug(v...)
}

// Info writes log message with defaultLogger.Info level
func Info(v ...interface{}) {
	defaultLogger.Info(v...)
}

// Warn writes log message with defaultLogger.Warn level
func Warn(v ...interface{}) {
	defaultLogger.Warn(v...)
}

// Error writes log message with defaultLogger.Error level
func Error(v ...interface{}) {
	defaultLogger.Error(v...)
}
