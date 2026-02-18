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

func TestSendDefaultServiceEvent_WithInitializedSensor(t *testing.T) {
	// Initialize sensor with a service name
	InitSensor(&Options{Service: "test-service"})
	defer ShutdownSensor()

	// Test with initialized sensor - should use service name
	SendDefaultServiceEvent("Test Event", "Event with initialized sensor",
		SeverityWarning, 1000*time.Millisecond)
}

func TestSendDefaultServiceEvent_WithoutInitializedSensor(t *testing.T) {
	// Ensure sensor is not initialized
	ShutdownSensor()

	// Test without initialized sensor - should handle error gracefully
	// This tests the error path in getSensor()
	SendDefaultServiceEvent("Test Event", "Event without initialized sensor",
		SeverityChange, 500*time.Millisecond)

	// Clean up any sensor that may have been initialized by sendEvent
	defer ShutdownSensor()
}

func TestSendDefaultServiceEvent_WithBinaryName(t *testing.T) {
	// Initialize sensor without service name - should use binary name
	InitSensor(&Options{})
	defer ShutdownSensor()

	// Test with binary name fallback
	SendDefaultServiceEvent("Binary Event", "Event using binary name",
		SeverityWarning, 750*time.Millisecond)
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
