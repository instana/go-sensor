package logger

import l "log"

// Level is the log level
type Level uint8

// Valid log levels compatible with github.com/instana/go-sensor log level values
const (
	ErrorLevel Level = iota
	WarnLevel
	InfoLevel
	DebugLevel
)

var defaultLogger *logWrapper = &logWrapper{DebugLevel}

// SetLogLevel configures the min log level for autoprofile defaultLogger
func SetLogLevel(level Level) {
	defaultLogger.logLevel = level
}

// Debug writes log message with defaultLogger.Debug level
func Debug(v ...interface{}) {
	defaultLogger.debug(v...)
}

// Info writes log message with defaultLogger.Info level
func Info(v ...interface{}) {
	defaultLogger.info(v...)
}

// Warn writes log message with defaultLogger.Warn level
func Warn(v ...interface{}) {
	defaultLogger.warn(v...)
}

// Error writes log message with defaultLogger.Error level
func Error(v ...interface{}) {
	defaultLogger.error(v...)
}

type logWrapper struct {
	logLevel Level
}

func (lw *logWrapper) makeV(prefix string, v ...interface{}) []interface{} {
	return append([]interface{}{prefix}, v...)
}

func (lw *logWrapper) debug(v ...interface{}) {
	if lw.logLevel >= DebugLevel {
		l.Println(lw.makeV("DEBUG: instana:", v...)...)
	}
}

func (lw *logWrapper) info(v ...interface{}) {
	if lw.logLevel >= InfoLevel {
		l.Println(lw.makeV("INFO: instana:", v...)...)
	}
}

func (lw *logWrapper) warn(v ...interface{}) {
	if lw.logLevel >= WarnLevel {
		l.Println(lw.makeV("WARN: instana:", v...)...)
	}
}

func (lw *logWrapper) error(v ...interface{}) {
	if lw.logLevel >= ErrorLevel {
		l.Println(lw.makeV("ERROR: instana:", v...)...)
	}
}
