package autoprofile

import "time"

// timer periodically executes provided job after a delay until it's stopped. Any panic
// occurred inside the job is recovered and logged
type timer struct {
	delayTimer         *time.Timer
	delayTimerDone     chan bool
	intervalTicker     *time.Ticker
	intervalTickerDone chan bool
	stopped            bool
}

func newTimer(delay time.Duration, interval time.Duration, job func()) *timer {
	t := &timer{
		stopped: false,
	}

	t.delayTimerDone = make(chan bool)
	t.delayTimer = time.NewTimer(delay)
	go func() {
		defer recoverAndLog()

		select {
		case <-t.delayTimer.C:
			if interval > 0 {
				t.intervalTickerDone = make(chan bool)
				t.intervalTicker = time.NewTicker(interval)
				go func() {
					defer recoverAndLog()

					for {
						select {
						case <-t.intervalTicker.C:
							job()
						case <-t.intervalTickerDone:
							return
						}
					}
				}()
			}

			if delay > 0 {
				job()
			}
		case <-t.delayTimerDone:
			return
		}
	}()

	return t
}

// Stop stops the job execution
func (t *timer) Stop() {
	if t.stopped {
		return
	}

	t.stopped = true

	t.delayTimer.Stop()
	close(t.delayTimerDone)

	if t.intervalTicker != nil {
		t.intervalTicker.Stop()
		close(t.intervalTickerDone)
	}
}

func recoverAndLog() {
	if err := recover(); err != nil {
		log.error("recovered from panic in agent:", err)
	}
}
