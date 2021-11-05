// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

package instalogrus

import (
	instana "github.com/instana/go-sensor"
	"github.com/sirupsen/logrus"
)

type hook struct {
	sensor *instana.Sensor
}

// NewHook returns a new logrus.Hook to instrument logger with Instana
func NewHook(sensor *instana.Sensor) *hook {
	return &hook{
		sensor: sensor,
	}
}

// Levels returns the list of log levels to be sent to Instana
func (h *hook) Levels() []logrus.Level {
	return nil
}

// Fire forwards the logrus.Entry to Instana
func (h *hook) Fire(entry *logrus.Entry) error {
	return nil
}
