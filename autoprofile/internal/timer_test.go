// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package internal_test

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/instana/go-sensor/autoprofile/internal"
	"github.com/stretchr/testify/assert"
)

func TestTimer_Stop_Restart(t *testing.T) {
	var fired int64
	scheduleTime := 60 * time.Millisecond

	start := time.Now()
	timer := internal.NewTimer(0, scheduleTime, func() {
		atomic.AddInt64(&fired, 1)
	})

	time.Sleep(100 * time.Millisecond)
	timer.Stop()

	elapsed := time.Since(start)
	expectation := elapsed / scheduleTime

	assert.EqualValues(t, expectation, atomic.LoadInt64(&fired))

	time.Sleep(200 * time.Millisecond)
	assert.EqualValues(t, expectation, atomic.LoadInt64(&fired), "a stopped timer should not be restarted")
}

func TestTimer_Sleep_Stopped(t *testing.T) {
	timer := internal.NewTimer(20*time.Millisecond, 0, func() {
		t.Error("stopped timer has fired")
	})

	timer.Stop()
	time.Sleep(30 * time.Millisecond)
}
