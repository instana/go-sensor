// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

package instamongo

import (
	"context"

	instana "github.com/instana/go-sensor"
	"go.mongodb.org/mongo-driver/event"
)

type CommandMonitor struct {
	sensor *instana.Sensor
}

func NewCommandMonitor(sensor *instana.Sensor) *CommandMonitor {
	return &CommandMonitor{
		sensor: sensor,
	}
}

func (*CommandMonitor) Started(ctx context.Context, evt *event.CommandStartedEvent) {}

func (*CommandMonitor) Succeeded(ctx context.Context, evt *event.CommandSucceededEvent) {}

func (*CommandMonitor) Failed(ctx context.Context, evt *event.CommandFailedEvent) {}
