// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2016

package instana

import (
	"sync"
	"testing"
	"time"
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
