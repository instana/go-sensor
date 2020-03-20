package logger_test

import (
	"testing"

	"github.com/instana/go-sensor/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogger_SetPrefix(t *testing.T) {
	p := &printer{}

	l := logger.New(p)
	l.SetPrefix("test-logger>>")
	l.Error("error level")

	require.Len(t, p.Records, 1)
	assert.Equal(t, []interface{}{"test-logger>>", "ERROR", ": ", "error level"}, p.Records[0])
}

func TestLogger_SetPrefix_DefaultValue(t *testing.T) {
	p := &printer{}

	l := logger.New(p)
	l.Error("error level")

	require.Len(t, p.Records, 1)
	assert.Equal(t, []interface{}{"instana: ", "ERROR", ": ", "error level"}, p.Records[0])
}

func TestLogger_SetLevel(t *testing.T) {
	examples := map[logger.Level][][]interface{}{
		logger.DebugLevel: {
			{"instana: ", "DEBUG", ": ", "debuglevel"},
			{"instana: ", "INFO", ": ", "infolevel"},
			{"instana: ", "WARN", ": ", "warnlevel"},
			{"instana: ", "ERROR", ": ", "errorlevel"},
		},
		logger.InfoLevel: {
			{"instana: ", "INFO", ": ", "infolevel"},
			{"instana: ", "WARN", ": ", "warnlevel"},
			{"instana: ", "ERROR", ": ", "errorlevel"},
		},
		logger.WarnLevel: {
			{"instana: ", "WARN", ": ", "warnlevel"},
			{"instana: ", "ERROR", ": ", "errorlevel"},
		},
		logger.ErrorLevel: {
			{"instana: ", "ERROR", ": ", "errorlevel"},
		},
	}

	for lvl, expected := range examples {
		t.Run(lvl.String(), func(t *testing.T) {
			p := &printer{}

			l := logger.New(p)
			l.SetLevel(lvl)

			l.Debug("debug", "level")
			l.Info("info", "level")
			l.Warn("warn", "level")
			l.Error("error", "level")

			assert.Equal(t, expected, p.Records)
		})
	}
}

type printer struct {
	Records [][]interface{}
}

func (p *printer) Print(args ...interface{}) {
	p.Records = append(p.Records, args)
}
