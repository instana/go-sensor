// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2016

package instana

import (
	"github.com/instana/go-sensor/logger"
)

// Valid log levels
const (
	Error = 0
	Warn  = 1
	Info  = 2
	Debug = 3
)

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

var defaultLogger LeveledLogger = logger.New(nil)

// SetLogger configures the default logger to be used by Instana go-sensor. Note that changing the logger
// will not affect already initialized instana.Sensor instances. To make them use the new logger please call
// (instana.TracerLogger).SetLogger() explicitly.
func SetLogger(l LeveledLogger) {
	defaultLogger = l

	// if the sensor has already been initialized, we need to update its logger too
	muSensor.RLock()
	defer muSensor.RUnlock()

	if sensor != nil {
		sensor.setLogger(l)
	}
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
