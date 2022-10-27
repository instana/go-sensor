// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
)

// Valid log levels to be used with (*logger.Logger).SetLevel()
const (
	ErrorLevel Level = iota
	WarnLevel
	InfoLevel
	DebugLevel
)

// DefaultPrefix is the default log prefix used by Logger
const DefaultPrefix = "instana: "

// Level defines the minimum logging level for logger.Log
type Level uint8

// Less returns whether the log level is less than the given one in logical order:
// ErrorLevel > WarnLevel > InfoLevel > DebugLevel
func (lvl Level) Less(other Level) bool {
	return uint8(lvl) > uint8(other)
}

// String returns the log line label for this level
func (lvl Level) String() string {
	switch lvl {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Printer is used by logger.Log instance to print out a log message
type Printer interface {
	Print(a ...interface{})
}

// Logger is a configurable leveled logger used by Instana's Go sensor. It follows the same interface
// as github.com/sirupsen/logrus.Logger and go.uber.org/zap.SugaredLogger
type Logger struct {
	p Printer

	mu     sync.Mutex
	lvl    Level
	prefix string
}

// New initializes a new instance of Logger that uses provided printer as a backend to
// output the log messages. The stdlib log.Logger satisfies logger.Printer interface:
//
//	logger := logger.New(logger.WarnLevel, log.New(os.Stderr, "instana:", log.LstdFlags))
//	logger.SetLevel(logger.WarnLevel)
//
//	logger.Debug("this is a debug message") // won't be printed
//	logger.Error("this is an  message") // ... while this one will
//
// In case  there is no printer provided, logger.Logger will use a new instance of log.Logger
// initialized with log.Lstdflags that writes to os.Stderr:
//
//	log.New(os.Stderr, "", log.Lstdflags)
//
// The default logging level for a new logger instance is logger.ErrorLevel.
func New(printer Printer) *Logger {
	if printer == nil {
		printer = log.New(os.Stderr, "", log.LstdFlags)
	}

	l := &Logger{
		p:      printer,
		prefix: DefaultPrefix,
	}

	setInstanaLogLevel(l)

	return l
}

func setInstanaLogLevel(l *Logger) {
	instanaLogLevelEnvVar := os.Getenv("INSTANA_LOG_LEVEL")

	envVarLogLevel := map[string]Level{
		"error": ErrorLevel,
		"info":  InfoLevel,
		"warn":  WarnLevel,
		"debug": DebugLevel,
	}

	if value, ok := envVarLogLevel[strings.ToLower(instanaLogLevelEnvVar)]; ok {
		l.SetLevel(value)
	} else {
		l.SetLevel(ErrorLevel)
	}
}

// SetLevel changes the log level for this logger instance. In case there is an INSTANA_DEBUG env variable set,
// the provided log level will be overridden with DebugLevel.
func (l *Logger) SetLevel(level Level) {
	if _, ok := os.LookupEnv("INSTANA_DEBUG"); ok {
		if level != DebugLevel {
			defer l.Info(
				"INSTANA_DEBUG env variable is set, the log level has been set to ",
				DebugLevel,
				" instead of requested ",
				level,
			)
		}
		level = DebugLevel
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	l.lvl = level
}

// SetPrefix sets the label that will be used as a prefix for each log line
func (l *Logger) SetPrefix(prefix string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.prefix = prefix
}

// Debug appends a debug message to the log
func (l *Logger) Debug(v ...interface{}) {
	if l.lvl < DebugLevel {
		return
	}

	l.print(DebugLevel, v)
}

// Info appends an info message to the log
func (l *Logger) Info(v ...interface{}) {
	if l.lvl < InfoLevel {
		return
	}

	l.print(InfoLevel, v)
}

// Warn appends a warning message to the log
func (l *Logger) Warn(v ...interface{}) {
	if l.lvl < WarnLevel {
		return
	}

	l.print(WarnLevel, v)
}

// Error appends an error message to the log
func (l *Logger) Error(v ...interface{}) {
	if l.lvl < ErrorLevel {
		return
	}

	l.print(ErrorLevel, v)
}

func (l *Logger) print(lvl Level, v []interface{}) {
	l.p.Print(l.prefix, lvl.String(), ": ", fmt.Sprint(v...))
}
