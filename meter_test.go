// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2016

package instana

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMeterS_Stop(t *testing.T) {
	// Create a new meter
	m := newMeter(defaultLogger)

	// Track if Run is still executing
	var wg sync.WaitGroup
	wg.Add(1)

	// Start the meter in a goroutine
	go func() {
		defer wg.Done()
		m.Run(100 * time.Millisecond)
	}()

	// Let it run for a bit
	time.Sleep(300 * time.Millisecond)

	// Stop the meter
	m.Stop()

	// Wait for Run to exit with a timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success - Run exited after Stop was called
	case <-time.After(2 * time.Second):
		t.Fatal("meter.Run() did not exit after Stop() was called")
	}
}

func TestMeterS_Run_StopImmediately(t *testing.T) {
	// Create a new meter
	m := newMeter(defaultLogger)

	// Track if Run is still executing
	var wg sync.WaitGroup
	wg.Add(1)

	// Start the meter in a goroutine
	go func() {
		defer wg.Done()
		m.Run(100 * time.Millisecond)
	}()

	// Stop immediately without waiting
	m.Stop()

	// Wait for Run to exit with a timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success - Run exited after Stop was called
	case <-time.After(2 * time.Second):
		t.Fatal("meter.Run() did not exit after immediate Stop() was called")
	}
}

func TestMeterS_CollectMetrics(t *testing.T) {
	// Create a new meter
	m := newMeter(defaultLogger)

	// Collect metrics
	metrics := m.collectMetrics()

	// Verify metrics are collected
	if metrics.Goroutine <= 0 {
		t.Errorf("Expected positive goroutine count, got %d", metrics.Goroutine)
	}

	if metrics.MemoryStats.Alloc == 0 {
		t.Error("Expected non-zero memory allocation")
	}
}

func TestMeterS_CollectMemoryMetrics(t *testing.T) {
	// Create a new meter
	m := newMeter(defaultLogger)

	// Collect memory metrics
	memStats := m.collectMemoryMetrics()

	// Verify memory stats are collected
	if memStats.Alloc == 0 {
		t.Error("Expected non-zero Alloc")
	}

	if memStats.Sys == 0 {
		t.Error("Expected non-zero Sys")
	}

	if memStats.HeapAlloc == 0 {
		t.Error("Expected non-zero HeapAlloc")
	}
}

func TestMeterS_NewMeter(t *testing.T) {
	// Create a new meter
	m := newMeter(defaultLogger)

	if m == nil {
		t.Fatal("Expected non-nil meter")
	}

	if m.done == nil {
		t.Error("Expected done channel to be initialized")
	}

	if m.numGC != 0 {
		t.Errorf("Expected initial numGC to be 0, got %d", m.numGC)
	}
}

func TestMetricsOptions_GetTransmissionInterval_Default(t *testing.T) {
	opts := &MetricsOptions{}

	interval := opts.getTransmissionInterval()

	assert.Equal(t, 1*time.Second, interval, "Default transmission interval should be 1 second")
}

func TestMetricsOptions_SetTransmissionInterval(t *testing.T) {
	tests := []struct {
		name     string
		seconds  int
		expected time.Duration
	}{
		{
			name:     "Valid 1 second (minimum)",
			seconds:  1,
			expected: 1 * time.Second,
		},
		{
			name:     "Valid 5 seconds",
			seconds:  5,
			expected: 5 * time.Second,
		},
		{
			name:     "Valid 60 seconds",
			seconds:  60,
			expected: 60 * time.Second,
		},
		{
			name:     "Valid 300 seconds",
			seconds:  300,
			expected: 300 * time.Second,
		},
		{
			name:     "Valid 3600 seconds (maximum)",
			seconds:  3600,
			expected: 3600 * time.Second,
		},
		{
			name:     "Zero seconds sets to minimum (1 second)",
			seconds:  0,
			expected: 1 * time.Second,
		},
		{
			name:     "Negative value sets to minimum (1 second)",
			seconds:  -1,
			expected: 1 * time.Second,
		},
		{
			name:     "Negative value -100 sets to minimum (1 second)",
			seconds:  -100,
			expected: 1 * time.Second,
		},
		{
			name:     "Value exceeding maximum (3601) sets to maximum (3600 seconds)",
			seconds:  3601,
			expected: 3600 * time.Second,
		},
		{
			name:     "Value exceeding maximum (5000) sets to maximum (3600 seconds)",
			seconds:  5000,
			expected: 3600 * time.Second,
		},
		{
			name:     "Value exceeding maximum (10000) sets to maximum (3600 seconds)",
			seconds:  10000,
			expected: 3600 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &MetricsOptions{}

			opts.setTransmissionInterval(tt.seconds)

			assert.Equal(t, tt.expected, opts.getTransmissionInterval())
		})
	}
}

func TestMeterS_Reset(t *testing.T) {
	t.Run("reset running meter", func(t *testing.T) {
		m := newMeter(defaultLogger)
		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			defer wg.Done()
			m.Run(100 * time.Millisecond)
		}()

		time.Sleep(150 * time.Millisecond)

		m.mu.Lock()
		assert.True(t, m.running, "Meter should be running before reset")
		m.mu.Unlock()

		m.reset(200 * time.Millisecond)
		time.Sleep(100 * time.Millisecond)

		m.mu.Lock()
		assert.True(t, m.running, "Meter should be running after reset")
		m.mu.Unlock()

		m.Stop()

		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatal("meter.Run() did not exit after Stop() was called")
		}
	})

	t.Run("multiple resets with different intervals", func(t *testing.T) {
		m := newMeter(defaultLogger)

		m.Run(100 * time.Millisecond)
		time.Sleep(150 * time.Millisecond)

		m.reset(50 * time.Millisecond)
		time.Sleep(100 * time.Millisecond)

		m.reset(200 * time.Millisecond)
		time.Sleep(100 * time.Millisecond)

		m.mu.Lock()
		running := m.running
		m.mu.Unlock()

		assert.True(t, running, "Meter should still be running after multiple resets")

		m.Stop()

		m.mu.Lock()
		assert.False(t, m.running, "Meter should be stopped")
		m.mu.Unlock()
	})

	t.Run("reset without initial run", func(t *testing.T) {
		m := newMeter(defaultLogger)

		m.mu.Lock()
		assert.False(t, m.running, "Meter should not be running initially")
		m.mu.Unlock()

		m.reset(100 * time.Millisecond)
		time.Sleep(150 * time.Millisecond)

		m.mu.Lock()
		running := m.running
		m.mu.Unlock()

		assert.True(t, running, "Meter should be running after reset")

		m.Stop()

		m.mu.Lock()
		assert.False(t, m.running, "Meter should be stopped")
		m.mu.Unlock()
	})
}

func TestMeterS_ConcurrentOperations(t *testing.T) {
	t.Run("concurrent stop and run", func(t *testing.T) {
		m := newMeter(defaultLogger)

		m.Run(100 * time.Millisecond)
		time.Sleep(50 * time.Millisecond)

		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(2)
			go func() {
				defer wg.Done()
				m.Stop()
			}()
			go func() {
				defer wg.Done()
				m.Run(100 * time.Millisecond)
			}()
		}

		wg.Wait()
		m.Stop()

		m.mu.Lock()
		assert.False(t, m.running, "Meter should be stopped after concurrent operations")
		m.mu.Unlock()
	})

	t.Run("concurrent reset", func(t *testing.T) {
		m := newMeter(defaultLogger)

		m.Run(100 * time.Millisecond)
		time.Sleep(50 * time.Millisecond)

		var wg sync.WaitGroup
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(interval time.Duration) {
				defer wg.Done()
				m.reset(interval)
			}(time.Duration(50+i*10) * time.Millisecond)
		}

		wg.Wait()
		time.Sleep(100 * time.Millisecond)

		m.mu.Lock()
		running := m.running
		m.mu.Unlock()

		assert.True(t, running, "Meter should be running after concurrent resets")

		m.Stop()

		m.mu.Lock()
		assert.False(t, m.running, "Meter should be stopped")
		m.mu.Unlock()
	})
}

func TestMeterS_StopAndRestart(t *testing.T) {
	t.Run("stop multiple times", func(t *testing.T) {
		m := newMeter(defaultLogger)

		m.Run(100 * time.Millisecond)
		time.Sleep(50 * time.Millisecond)

		// Stop multiple times - should not panic
		m.Stop()
		m.Stop()
		m.Stop()

		m.mu.Lock()
		assert.False(t, m.running, "Meter should be stopped")
		m.mu.Unlock()
	})

	t.Run("run after stop", func(t *testing.T) {
		m := newMeter(defaultLogger)

		m.Run(100 * time.Millisecond)
		time.Sleep(150 * time.Millisecond)

		m.Stop()
		time.Sleep(50 * time.Millisecond)

		m.mu.Lock()
		assert.False(t, m.running, "Meter should be stopped")
		m.mu.Unlock()

		// Start again
		m.Run(100 * time.Millisecond)
		time.Sleep(150 * time.Millisecond)

		m.mu.Lock()
		running := m.running
		m.mu.Unlock()

		assert.True(t, running, "Meter should be running after restart")

		m.Stop()
	})
}

func TestMeterS_Reset_InternalState(t *testing.T) {
	t.Run("preserves numGC", func(t *testing.T) {
		m := newMeter(defaultLogger)

		_ = m.collectMetrics()

		m.mu.Lock()
		initialNumGC := m.numGC
		m.mu.Unlock()

		m.Run(100 * time.Millisecond)
		time.Sleep(50 * time.Millisecond)
		m.reset(200 * time.Millisecond)
		time.Sleep(50 * time.Millisecond)

		m.mu.Lock()
		currentNumGC := m.numGC
		m.mu.Unlock()

		assert.GreaterOrEqual(t, currentNumGC, initialNumGC, "numGC should be preserved or increased")

		m.Stop()
	})

	t.Run("creates new done channel", func(t *testing.T) {
		m := newMeter(defaultLogger)

		m.Run(100 * time.Millisecond)
		time.Sleep(50 * time.Millisecond)

		m.mu.Lock()
		firstDone := m.done
		m.mu.Unlock()

		m.reset(200 * time.Millisecond)
		time.Sleep(50 * time.Millisecond)

		m.mu.Lock()
		secondDone := m.done
		m.mu.Unlock()

		assert.NotEqual(t, firstDone, secondDone, "Reset should create a new done channel")

		m.Stop()
	})
}
