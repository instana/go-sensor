package internal_test

import (
	"testing"
	"time"

	"github.com/instana/go-sensor/autoprofile/internal"
	"github.com/stretchr/testify/assert"
)

func TestTimer_Restart(t *testing.T) {
	var fired int
	timer := internal.NewTimer(0, 20*time.Millisecond, func() {
		fired++
	})

	time.Sleep(30 * time.Millisecond)
	timer.Stop()

	assert.Equal(t, 1, fired)

	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 1, fired)
}

func TestTimer_Sleep(t *testing.T) {
	var fired int
	timer := internal.NewTimer(20*time.Millisecond, 0, func() {
		fired++
	})

	time.Sleep(30 * time.Millisecond)
	timer.Stop()

	assert.Equal(t, 1, fired)
}

func TestTimer_Sleep_Stopped(t *testing.T) {
	var fired int
	timer := internal.NewTimer(20*time.Millisecond, 0, func() {
		fired++
	})

	timer.Stop()
	time.Sleep(30 * time.Millisecond)

	assert.Equal(t, 0, fired)
}
