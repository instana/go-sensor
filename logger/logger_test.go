// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package logger_test

import (
	"os"
	"testing"

	"github.com/instana/go-sensor/logger"
	"github.com/instana/testify/assert"
	"github.com/instana/testify/require"
)

func TestLevel_Less(t *testing.T) {
	levels := []logger.Level{logger.ErrorLevel, logger.WarnLevel, logger.InfoLevel, logger.DebugLevel}
	for i := range levels[:len(levels)-1] {
		assert.True(t, levels[i+1].Less(levels[i]), "%s should be less than %s", levels[i+1], levels[i])
	}
}

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
			originalEnvVal, restoreOriginalVal := os.LookupEnv("INSTANA_DEBUG")
			os.Unsetenv("INSTANA_DEBUG")

			// restore original value
			if restoreOriginalVal {
				defer func() {
					os.Setenv("INSTANA_DEBUG", originalEnvVal)
				}()
			}

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

	for lvl, expected := range examples {
		t.Run(lvl.String()+" INSTANA_LOG_LEVEL env var", func(t *testing.T) {
			p := &printer{}

			t.Setenv("INSTANA_LOG_LEVEL", lvl.String())
			l := logger.New(p)
			l.Debug("debug", "level")
			l.Info("info", "level")
			l.Warn("warn", "level")
			l.Error("error", "level")

			assert.Equal(t, expected, p.Records)
		})
	}

	t.Run("INSTANA_LOG_LEVEL env var replaced by SetLevel", func(t *testing.T) {
		p := &printer{}

		t.Setenv("INSTANA_LOG_LEVEL", "wArn")
		l := logger.New(p)
		l.Debug("debug", "level")
		l.Info("info", "level")
		l.Warn("warn", "level")
		l.Error("error", "level")

		assert.Equal(t, examples[logger.WarnLevel], p.Records)

		p.Records = p.Records[:0]

		l.SetLevel(logger.InfoLevel)
		l.Debug("debug", "level")
		l.Info("info", "level")
		l.Warn("warn", "level")
		l.Error("error", "level")

		assert.Equal(t, examples[logger.InfoLevel], p.Records)
	})
}

func TestLogger_SetLevel_INSTANA_DEBUG(t *testing.T) {
	levels := []logger.Level{
		logger.DebugLevel,
		logger.InfoLevel,
		logger.WarnLevel,
		logger.ErrorLevel,
	}
	for _, lvl := range levels {
		t.Run(lvl.String(), func(t *testing.T) {
			originalEnvVal, restoreOriginalVal := os.LookupEnv("INSTANA_DEBUG")
			os.Setenv("INSTANA_DEBUG", "yes")

			// restore original value
			defer func() {
				os.Unsetenv("INSTANA_DEBUG")
				if !restoreOriginalVal {
					os.Setenv("INSTANA_DEBUG", originalEnvVal)
				}
			}()

			p := &printer{}

			l := logger.New(p)
			l.SetLevel(lvl)

			l.Debug("debug", "level")
			assert.Contains(t, p.Records, []interface{}{"instana: ", "DEBUG", ": ", "debuglevel"})
		})
	}
}

type printer struct {
	Records [][]interface{}
}

func (p *printer) Print(args ...interface{}) {
	p.Records = append(p.Records, args)
}
