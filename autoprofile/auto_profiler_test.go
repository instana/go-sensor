package autoprofile

import (
	"testing"
	"time"
)

func TestTimerPeriod(t *testing.T) {
	fired := 0
	timer := createTimer(0, 10*time.Millisecond, func() {
		fired++
	})

	time.Sleep(20 * time.Millisecond)

	timer.Stop()

	time.Sleep(30 * time.Millisecond)

	if fired > 2 {
		t.Errorf("interval fired too many times: %v", fired)
	}
}

func TestTimerDelay(t *testing.T) {
	fired := 0
	timer := createTimer(10*time.Millisecond, 0, func() {
		fired++
	})

	time.Sleep(20 * time.Millisecond)

	timer.Stop()

	if fired != 1 {
		t.Errorf("delay should fire once: %v", fired)
	}
}

func TestTimerDelayStop(t *testing.T) {
	fired := 0
	timer := createTimer(10*time.Millisecond, 0, func() {
		fired++
	})

	timer.Stop()

	time.Sleep(20 * time.Millisecond)

	if fired == 1 {
		t.Errorf("delay should not fire")
	}
}
