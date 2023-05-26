// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

package instalogrus

import (
	instana "github.com/instana/go-sensor"
	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
)

type hook struct {
	sensor instana.TracerLogger
}

// NewHook returns a new logrus.Hook to instrument logger with Instana
func NewHook(sensor instana.TracerLogger) *hook {
	return &hook{
		sensor: sensor,
	}
}

// Levels returns the list of log levels to be sent to Instana
func (h *hook) Levels() []logrus.Level {
	return []logrus.Level{logrus.ErrorLevel, logrus.WarnLevel}
}

// Fire forwards the logrus.Entry to Instana
func (h *hook) Fire(entry *logrus.Entry) error {
	if entry.Context == nil {
		h.sensor.Logger().Debug("ignoring logrus.Entry without context.Context")
		return nil
	}

	parent, ok := instana.SpanFromContext(entry.Context)
	if !ok {
		return nil
	}

	msg, err := entry.String()
	if err != nil {
		h.sensor.Logger().Error("failed to obtain logrus.Entry data:", err)
		return nil
	}

	h.sensor.Tracer().StartSpan("log.go",
		opentracing.ChildOf(parent.Context()),
		opentracing.StartTime(entry.Time),
		opentracing.Tags{
			"log.level":   convertLevel(entry.Level),
			"log.message": string(msg),
		},
	).FinishWithOptions(opentracing.FinishOptions{
		FinishTime: entry.Time,
	})

	return nil
}

func convertLevel(lvl logrus.Level) string {
	switch lvl {
	case logrus.ErrorLevel:
		return "ERROR"
	case logrus.WarnLevel:
		return "WARN"
	case logrus.InfoLevel:
		return "INFO"
	default:
		return "DEBUG"
	}
}
