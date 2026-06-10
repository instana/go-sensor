// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instana_test

import (
	"context"
	"errors"
	"testing"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const TestServiceName = "test_service"

// mockAgentClient is a mock implementation of AgentClient for testing
type mockAgentClient struct {
	ready      bool
	flushError error
	flushCalls int
}

func (m *mockAgentClient) Ready() bool {
	return m.ready
}

func (m *mockAgentClient) SendMetrics(data acceptor.Metrics) error {
	return nil
}

func (m *mockAgentClient) SendEvent(event *instana.EventData) error {
	return nil
}

func (m *mockAgentClient) SendSpans(spans []instana.Span) error {
	return nil
}

func (m *mockAgentClient) SendProfiles(profiles []autoprofile.Profile) error {
	return nil
}

func (m *mockAgentClient) Flush(ctx context.Context) error {
	m.flushCalls++
	return m.flushError
}

func TestReady_WithInitializedCollector(t *testing.T) {
	tests := []struct {
		name          string
		agentReady    bool
		expectedReady bool
	}{
		{
			name:          "agent is ready",
			agentReady:    true,
			expectedReady: true,
		},
		{
			name:          "agent is not ready",
			agentReady:    false,
			expectedReady: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock agent
			mockAgent := &mockAgentClient{ready: tt.agentReady}

			// Initialize collector with custom options
			opts := &instana.Options{
				Service:     TestServiceName,
				AgentClient: mockAgent,
			}

			collector := instana.InitCollector(opts)
			require.NotNil(t, collector)
			defer instana.ShutdownCollector()

			// Test Ready() - it should reflect the agent's ready state
			ready := instana.Ready()
			assert.Equal(t, tt.expectedReady, ready, "Ready() should return %v", tt.expectedReady)
		})
	}
}

func TestReady_BeforeInitialization(t *testing.T) {
	// Ensure collector is shut down before test
	instana.ShutdownCollector()

	// Test Ready() when collector is not initialized
	// This should return false
	ready := instana.Ready()
	assert.False(t, ready, "Ready() should return false when collector is not initialized")
}

func TestFlush_WithInitializedCollector(t *testing.T) {
	tests := []struct {
		name          string
		flushError    error
		expectedError bool
	}{
		{
			name:          "successful flush",
			flushError:    nil,
			expectedError: false,
		},
		{
			name:          "flush with error",
			flushError:    errors.New("flush failed"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock agent
			mockAgent := &mockAgentClient{
				ready:      true,
				flushError: tt.flushError,
			}

			// Initialize collector with custom options
			opts := &instana.Options{
				Service:     TestServiceName,
				AgentClient: mockAgent,
			}

			collector := instana.InitCollector(opts)
			require.NotNil(t, collector)
			defer instana.ShutdownCollector()

			// Test Flush()
			ctx := context.Background()
			err := instana.Flush(ctx)

			if tt.expectedError {
				assert.Error(t, err, "Flush() should return error")
				assert.Equal(t, tt.flushError, err, "Error should match expected error")
			} else {
				assert.NoError(t, err, "Flush() should not return error")
			}

			// Verify Flush was called
			assert.Equal(t, 1, mockAgent.flushCalls, "Flush should be called once")
		})
	}
}

func TestFlush_BeforeInitialization(t *testing.T) {
	// Ensure collector is shut down before test
	instana.ShutdownCollector()

	// Test Flush() when collector is not initialized
	ctx := context.Background()
	err := instana.Flush(ctx)
	assert.Error(t, err, "Flush() should return error when collector is not initialized")
}

func TestFlush_WithContext(t *testing.T) {
	tests := []struct {
		name        string
		ctx         context.Context
		setupCancel bool
	}{
		{
			name:        "with background context",
			ctx:         context.Background(),
			setupCancel: false,
		},
		{
			name:        "with cancellable context",
			ctx:         nil, // Will be created in test
			setupCancel: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAgent := &mockAgentClient{ready: true}

			opts := &instana.Options{
				Service:     TestServiceName,
				AgentClient: mockAgent,
			}

			collector := instana.InitCollector(opts)
			require.NotNil(t, collector)
			defer instana.ShutdownCollector()

			ctx := tt.ctx
			if tt.setupCancel {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(context.Background())
				defer cancel()
			}

			err := instana.Flush(ctx)
			assert.NoError(t, err, "Flush() should not return error")
			assert.Equal(t, 1, mockAgent.flushCalls, "Flush should be called once")
		})
	}
}

// TestInitSensor_DoesNotReinitializeWhenAlreadyInitialized validates that
// calling InitSensor multiple times does not reinitialize the sensor
func TestInitSensor_DoesNotReinitializeWhenAlreadyInitialized(t *testing.T) {
	// Ensure clean state
	instana.ShutdownCollector()

	// Create first agent
	firstAgent := &mockAgentClient{ready: true}

	opts1 := &instana.Options{
		Service:     "test-service-1",
		AgentClient: firstAgent,
	}

	// First initialization
	instana.InitSensor(opts1)
	defer instana.ShutdownCollector()

	// Verify sensor is ready
	require.True(t, instana.Ready(), "Sensor should be ready after first initialization")

	// Create second agent with different options
	secondAgent := &mockAgentClient{ready: false} // Different ready state

	opts2 := &instana.Options{
		Service:     "test-service-2", // Different service name
		AgentClient: secondAgent,      // Different agent
	}

	// Second initialization attempt - should return early without reinitializing
	instana.InitSensor(opts2)

	// Verify sensor is still ready with the FIRST agent (not the second)
	// If reinitialization happened, Ready() would return false (secondAgent.ready = false)
	assert.True(t, instana.Ready(), "Sensor should still use first agent (ready=true), not second agent (ready=false)")
}

// TestInitSensor_WithNilOptionsAfterInitialization validates that calling
// InitSensor with nil options after initialization doesn't reinitialize
func TestInitSensor_WithNilOptionsAfterInitialization(t *testing.T) {
	// Ensure clean state
	instana.ShutdownCollector()

	agent := &mockAgentClient{ready: true}

	opts := &instana.Options{
		Service:     "test-service",
		AgentClient: agent,
	}

	// First initialization with options
	instana.InitSensor(opts)
	defer instana.ShutdownCollector()

	require.True(t, instana.Ready(), "Sensor should be ready after initialization")

	// Second call with nil options - should return early without reinitializing
	instana.InitSensor(nil)

	// Sensor should still be ready with original configuration
	assert.True(t, instana.Ready(), "Sensor should still be ready with original config")
}

// TestStartMetrics_AfterInitCollector validates the edge case where
// InitCollector is called followed by StartMetrics (the reported issue)
func TestStartMetrics_AfterInitCollector(t *testing.T) {
	// Ensure clean state
	instana.ShutdownCollector()

	agent := &mockAgentClient{ready: true}

	opts := &instana.Options{
		Service:     "test-service",
		AgentClient: agent,
		Recorder:    instana.NewRecorder(),
	}

	// Initialize collector (which calls StartMetrics internally)
	collector := instana.InitCollector(opts)
	require.NotNil(t, collector)
	defer instana.ShutdownCollector()

	require.True(t, instana.Ready(), "Sensor should be ready after InitCollector")

	// Call StartMetrics again (which calls InitSensor internally)
	// This should return early without reinitializing
	instana.StartMetrics(nil)

	// Verify sensor is still ready and not reinitialized
	assert.True(t, instana.Ready(), "Sensor should still be ready after StartMetrics call")
}
