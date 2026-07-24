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

	// Call Run three times in quick succession — only the first must start the loop.
	m.Run(50 * time.Millisecond)
	m.Run(50 * time.Millisecond)
	m.Run(50 * time.Millisecond)

	// Give one tick to fire so we know the loop is alive.
	time.Sleep(80 * time.Millisecond)

	// Stop and confirm the channel closes cleanly (not double-closed/panicked).
	assert.NotPanics(t, m.Stop)

	// Confirm done channel is closed (reading from a closed channel returns immediately).
	select {
	case <-m.done:
		// expected
	default:
		t.Fatal("done channel should be closed after Stop()")
	}
}

// TestMeterStop_Idempotent verifies that calling Stop multiple times never panics.
func TestMeterStop_Idempotent(t *testing.T) {
	m := newMeter(defaultLogger)
	m.Run(100 * time.Millisecond)

	assert.NotPanics(t, func() {
		m.Stop()
		m.Stop()
		m.Stop()
	})
}

// TestMeterStop_WithoutRun verifies Stop is safe even when Run was never called.
func TestMeterStop_WithoutRun(t *testing.T) {
	m := newMeter(defaultLogger)

	assert.NotPanics(t, func() {
		m.Stop()
	})
}

// TestMeterRun_LoopExitsOnStop verifies the collection goroutine stops when Stop is called.
func TestMeterRun_LoopExitsOnStop(t *testing.T) {
	m := newMeter(defaultLogger)
	m.Run(50 * time.Millisecond)

	// Let at least one tick fire.
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

	// Should stop cleanly with no panic.
	assert.NotPanics(t, m.Stop)
}

// TestMetricsOptions_GetTransmissionInterval_Default verifies the default interval
// is returned when none has been configured.
func TestMetricsOptions_GetTransmissionInterval_Default(t *testing.T) {
	opts := &MetricsOptions{}
	assert.Equal(t, time.Second, opts.getTransmissionInterval())
}

// TestMetricsOptions_SetTransmissionInterval verifies clamping and valid values.
func TestMetricsOptions_SetTransmissionInterval(t *testing.T) {
	tests := []struct {
		name     string
		seconds  int
		expected time.Duration
	}{
		{"minimum (1s)", 1, 1 * time.Second},
		{"valid 5s", 5, 5 * time.Second},
		{"valid 60s", 60, 60 * time.Second},
		{"valid 300s", 300, 300 * time.Second},
		{"maximum (600s)", 600, 600 * time.Second},
		{"zero clamps to minimum", 0, 1 * time.Second},
		{"negative clamps to minimum", -1, 1 * time.Second},
		{"exceeds maximum clamps to 600s", 1000, 600 * time.Second},
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
