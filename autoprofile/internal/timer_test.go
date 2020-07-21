package internal_test

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/instana/go-sensor/autoprofile/internal"
	"github.com/instana/testify/assert"
)

func TestTimer_Restart(t *testing.T) {
	var fired int64
	timer := internal.NewTimer(0, 20*time.Millisecond, func() {
		atomic.AddInt64(&fired, 1)
	})

	time.Sleep(30 * time.Millisecond)
	timer.Stop()

	assert.EqualValues(t, 1, atomic.LoadInt64(&fired))

	time.Sleep(50 * time.Millisecond)
	assert.EqualValues(t, 1, atomic.LoadInt64(&fired))
}

func TestTimer_Sleep(t *testing.T) {
	var fired int64
	timer := internal.NewTimer(0, 20*time.Millisecond, func() {
		atomic.AddInt64(&fired, 1)
	})

	time.Sleep(30 * time.Millisecond)
	timer.Stop()

	assert.EqualValues(t, 1, atomic.LoadInt64(&fired))
}

func TestTimer_Sleep_Stopped(t *testing.T) {
	timer := internal.NewTimer(20*time.Millisecond, 0, func() {
		t.Error("stopped timer has fired")
	})

	timer.Stop()
	time.Sleep(30 * time.Millisecond)
}
