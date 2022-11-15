// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2017

package instana

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEventBasic(t *testing.T) {
	assert.Equal(t, severity(-1), SeverityChange, "SeverityChange wrong value...")
	assert.Equal(t, severity(5), SeverityWarning, "SeverityWarning wrong value...")
	assert.Equal(t, severity(10), SeverityCritical, "SeverityCritical wrong value...")
}
func TestEventDefault(t *testing.T) {
	SendDefaultServiceEvent("microservice-14c", "These are event details",
		SeverityCritical, 5000*time.Millisecond)
	defer ShutdownSensor()
}

func TestSendServiceEvent(t *testing.T) {
	SendServiceEvent("microservice-14c", "Oh No!", "Pull the cable now!",
		SeverityChange, 1000*time.Millisecond)
	defer ShutdownSensor()
}

func TestSendHostEvent(t *testing.T) {
	SendHostEvent("microservice-14c", "r u listening?",
		SeverityWarning, 500*time.Millisecond)
	defer ShutdownSensor()
}
