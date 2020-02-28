package autoprofile_test

import (
	"testing"
	"time"

	"github.com/instana/go-sensor/autoprofile"
	"github.com/stretchr/testify/assert"
)

func TestTimer_Restart(t *testing.T) {
	var fired int
	timer := autoprofile.NewTimer(0, 10*time.Millisecond, func() {
		fired++
	})

	time.Sleep(15 * time.Millisecond)
	timer.Stop()

	assert.Equal(t, 1, fired)

	time.Sleep(30 * time.Millisecond)
	assert.Equal(t, 1, fired)
}

func TestTimer_Sleep(t *testing.T) {
	var fired int
	timer := autoprofile.NewTimer(10*time.Millisecond, 0, func() {
		fired++
	})

	time.Sleep(15 * time.Millisecond)
	timer.Stop()

	assert.Equal(t, 1, fired)
}

func TestTimer_Sleep_Stopped(t *testing.T) {
	var fired int
	timer := autoprofile.NewTimer(10*time.Millisecond, 0, func() {
		fired++
	})

	timer.Stop()
	time.Sleep(15 * time.Millisecond)

	assert.Equal(t, 0, fired)
}
