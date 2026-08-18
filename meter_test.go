// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2016

package instana

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestNewMeter verifies the meter is properly initialised.
func TestNewMeter(t *testing.T) {
	m := newMeter(defaultLogger)

	assert.NotNil(t, m)
	assert.NotNil(t, m.done)
	assert.Equal(t, uint32(0), m.numGC.Load())
}

// TestMeterRun_StartsOnce ensures the collection goroutine is only started once
// regardless of how many times Run is called — matching the agent-reconnect use case
// where the FSM calls Run again but the loop must continue uninterrupted.
func TestMeterRun_StartsOnce(t *testing.T) {
	m := newMeter(defaultLogger)

	m.Run(50 * time.Millisecond)
	m.Run(50 * time.Millisecond)
	m.Run(50 * time.Millisecond)

	time.Sleep(80 * time.Millisecond)

	assert.NotPanics(t, m.Stop)

	select {
	case <-m.done:
		// expected: channel is closed
	default:
		t.Fatal("done channel should be closed after Stop()")
	}
}

// TestMeterRun_LoopExitsOnStop verifies the collection goroutine stops when Stop is called.
func TestMeterRun_LoopExitsOnStop(t *testing.T) {
	m := newMeter(defaultLogger)
	m.Run(50 * time.Millisecond)
	time.Sleep(80 * time.Millisecond)

	stopped := make(chan struct{})
	go func() {
		m.Stop()
		close(stopped)
	}()

	select {
	case <-stopped:
		// expected
	case <-time.After(2 * time.Second):
		t.Fatal("Stop() did not return in time")
	}
}

// TestMeterRun_ConcurrentCallsSafe verifies concurrent calls to Run are race-free.
func TestMeterRun_ConcurrentCallsSafe(t *testing.T) {
	m := newMeter(defaultLogger)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.Run(100 * time.Millisecond)
		}()
	}
	wg.Wait()

	assert.NotPanics(t, m.Stop)
}

// TestMeterStop covers idempotency and nil-receiver safety of Stop, and calling
// Stop before Run has been called.
func TestMeterStop(t *testing.T) {
	tests := []struct {
		name  string
		setup func() *meterS
		stop  func(*meterS)
	}{
		{
			name:  "idempotent: multiple Stop calls do not panic",
			setup: func() *meterS { m := newMeter(defaultLogger); m.Run(100 * time.Millisecond); return m },
			stop:  func(m *meterS) { m.Stop(); m.Stop(); m.Stop() },
		},
		{
			name:  "safe when Run was never called",
			setup: func() *meterS { return newMeter(defaultLogger) },
			stop:  func(m *meterS) { m.Stop() },
		},
		{
			name:  "nil receiver is a no-op",
			setup: func() *meterS { return nil },
			stop:  func(m *meterS) { m.Stop() },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.setup()
			assert.NotPanics(t, func() { tt.stop(m) })
		})
	}
}

// TestMeterRun_NilReceiver verifies Run is a no-op when called on a nil *meterS.
func TestMeterRun_NilReceiver(t *testing.T) {
	var m *meterS
	assert.NotPanics(t, func() { m.Run(50 * time.Millisecond) })
}

// TestMetricsOptions_GetTransmissionInterval_Default verifies that an unconfigured
// MetricsOptions returns zero (callers are responsible for applying the default).
func TestMetricsOptions_GetTransmissionInterval_Default(t *testing.T) {
	opts := &MetricsOptions{}
	assert.Equal(t, time.Duration(0), opts.getTransmissionInterval())
}

// TestMetricsOptions_SetTransmissionInterval verifies that positive values are stored
// as-is and non-positive values fall back to the default (1 second).
func TestMetricsOptions_SetTransmissionInterval(t *testing.T) {
	tests := []struct {
		name     string
		seconds  int
		expected time.Duration
	}{
		{"minimum canonical value (1s)", 1, 1 * time.Second},
		{"canonical 5s", 5, 5 * time.Second},
		{"canonical 60s", 60, 60 * time.Second},
		{"canonical 300s", 300, 300 * time.Second},
		{"canonical 600s", 600, 600 * time.Second},
		{"non-canonical positive value stored as-is (7s)", 7, 7 * time.Second},
		{"large positive value stored as-is (1000s)", 1000, 1000 * time.Second},
		{"zero uses default (1s)", 0, 1 * time.Second},
		{"negative uses default (1s)", -1, 1 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &MetricsOptions{}
			opts.setTransmissionInterval(tt.seconds)
			assert.Equal(t, tt.expected, opts.getTransmissionInterval())
		})
	}
}

// TestMeterCollectMetrics verifies that metric collection returns non-zero values.
func TestMeterCollectMetrics(t *testing.T) {
	m := newMeter(defaultLogger)
	metrics := m.collectMetrics()

	assert.Greater(t, metrics.Goroutine, 0)
	assert.NotZero(t, metrics.MemoryStats.Alloc)
}

// TestMeterCollectMemoryMetrics verifies that memory stats are populated.
func TestMeterCollectMemoryMetrics(t *testing.T) {
	m := newMeter(defaultLogger)
	mem := m.collectMemoryMetrics()

	assert.NotZero(t, mem.Alloc)
	assert.NotZero(t, mem.Sys)
	assert.NotZero(t, mem.HeapAlloc)
}

// TestMeterRun_SendMetrics_SensorNil verifies the meter loop skips SendMetrics when
// the global sensor is nil (agent not ready). The loop stays alive and stops cleanly.
func TestMeterRun_SendMetrics_SensorNil(t *testing.T) {
	m := newMeter(defaultLogger)
	m.Run(20 * time.Millisecond)
	// Give several ticks to fire; none should panic because isAgentReady returns false.
	time.Sleep(80 * time.Millisecond)
	assert.NotPanics(t, m.Stop)
}

// TestMeterRun_SendMetrics_AgentReady verifies that the tick handler enters the
// SendMetrics path when a ready sensor is present, without panic.
func TestMeterRun_SendMetrics_AgentReady(t *testing.T) {
	// Build a sensor with a ready mock agent.
	mock := &sensorS{
		options: DefaultOptions(),
		meter:   newMeter(defaultLogger),
	}
	mock.setLogger(defaultLogger)
	mock.setAgent(alwaysReadyClient{})

	// Protect all writes to sensor with muSensor so isAgentReady() (which holds
	// muSensor.RLock) does not race with this goroutine.
	muSensor.Lock()
	orig := sensor
	sensor = mock
	muSensor.Unlock()

	m := newMeter(defaultLogger)
	m.Run(20 * time.Millisecond)
	time.Sleep(80 * time.Millisecond)

	// Stop the meter BEFORE restoring sensor so the background goroutine is dead
	// before we mutate the shared variable again.
	assert.NotPanics(t, m.Stop)

	muSensor.Lock()
	sensor = orig
	muSensor.Unlock()
}
