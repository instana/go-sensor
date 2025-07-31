// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package internal

import (
	"sync"
	"time"

	"github.com/instana/go-sensor/autoprofile/internal/logger"
)

// Timer periodically executes provided job after a delay until it's stopped. Any panic
// occurred inside the job is recovered and logged
type Timer struct {
	mu         sync.Mutex
	delayTimer *time.Timer
	done       chan bool
	stopped    bool
	ticker     *time.Ticker
}

type Timer2 struct {
	stopChan chan bool
	stopOnce sync.Once
}

func NewTimer2(delay, interval time.Duration, job func()) *Timer2 {
	t := &Timer2{
		stopChan: make(chan bool),
	}

	go func() {
		select {
		case <-time.After(delay):
			job()
		case <-t.stopChan:
			return
		}

		if interval == 0 {
			return
		}

		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				job()
			case <-t.stopChan:
				return
			}
		}
	}()

	return t
}

func (t *Timer2) Stop() {
	t.stopOnce.Do(func() {
		close(t.stopChan)
	})
}

func recoverAndLog() {
	if err := recover(); err != nil {
		logger.Error("recovered from panic in agent:", err)
	}
}

// NewTimer initializes a new Timer
func NewTimer(delay, interval time.Duration, job func()) *Timer {
	t := &Timer{
		done: make(chan bool),
	}

	defer recoverAndLog()

	t.delayTimer = time.AfterFunc(delay, func() {
		defer recoverAndLog()

		if interval > 0 {
			go t.runTicker(interval, job)
		}

		if delay > 0 {
			job()
		}
	})

	return t
}

func (t *Timer) runTicker(interval time.Duration, job func()) {
	func() {
		t.mu.Lock()
		defer t.mu.Unlock()

		if t.stopped {
			return
		}

		t.ticker = time.NewTicker(interval)
	}()

	defer recoverAndLog()

	for {
		select {
		case <-t.ticker.C:
			job()
		case <-t.done:
			return
		}
	}
}

// Stop stops the job execution
func (t *Timer) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.stopped {
		return
	}

	t.stopped = true
	t.delayTimer.Stop()

	close(t.done)

	if t.ticker != nil {
		t.ticker.Stop()
	}
}
