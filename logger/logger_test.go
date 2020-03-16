package logger_test

import (
	"testing"

	"github.com/instana/go-sensor/logger"
	"github.com/stretchr/testify/assert"
)

func TestLogger_SetLevel(t *testing.T) {
	examples := map[logger.Level][][]interface{}{
		logger.DebugLevel: {
			{"DEBUG", ": ", "debuglevel"},
		},
		logger.InfoLevel: {
			{"DEBUG", ": ", "debuglevel"},
			{"INFO", ": ", "infolevel"},
		},
		logger.WarnLevel: {
			{"DEBUG", ": ", "debuglevel"},
			{"INFO", ": ", "infolevel"},
			{"WARN", ": ", "warnlevel"},
		},
		logger.ErrorLevel: {
			{"DEBUG", ": ", "debuglevel"},
			{"INFO", ": ", "infolevel"},
			{"WARN", ": ", "warnlevel"},
			{"ERROR", ": ", "errorlevel"},
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
