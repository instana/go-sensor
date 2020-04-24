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

func NewTimer(delay, interval time.Duration, job func()) *Timer {
	t := &Timer{
		done: make(chan bool),
	}

	t.delayTimer = time.AfterFunc(delay, func() {
		defer recoverAndLog()

		if interval > 0 {
			t.runTicker(interval, job)
		}

		if delay > 0 {
			job()
		}
	})

	return t
}

func (t *Timer) runTicker(interval time.Duration, job func()) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.stopped {
		return
	}

	t.ticker = time.NewTicker(interval)
	go func() {
		defer recoverAndLog()

		for {
			select {
			case <-t.ticker.C:
				job()
			case <-t.done:
				return
			}
		}
	}()
}

// Stop stops the job execution
func (t *Timer) Stop() {
	if t.stopped {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	t.stopped = true
	t.delayTimer.Stop()

	close(t.done)

	if t.ticker != nil {
		t.ticker.Stop()
	}
}

func recoverAndLog() {
	if err := recover(); err != nil {
		logger.Error("recovered from panic in agent:", err)
	}
}
