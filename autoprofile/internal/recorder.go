// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package internal

import (
	"sync"
	"time"

	"github.com/instana/go-sensor/autoprofile/internal/logger"
)

// DefaultMaxBufferedProfiles is the default number of profiles to keep in recorder buffer
const DefaultMaxBufferedProfiles = 100

// SendProfilesFunc is a callback to emit collected profiles from recorder
type SendProfilesFunc func([]AgentProfile) error

// NoopSendProfiles is the default function to be called by Recorded to send collected profiles
func NoopSendProfiles([]AgentProfile) error {
	logger.Warn(
		"autoprofile.SendProfiles callback is not set, ",
		"make sure that you have it configured using autoprofile.SetSendProfilesFunc() in your code",
	)

	return nil
}

// Recorder collects and stores recorded profiles
type Recorder struct {
	FlushInterval       int64
	MaxBufferedProfiles int
	SendProfiles        SendProfilesFunc

	started            Flag
	flushTimer         *Timer
	queue              []AgentProfile
	queueLock          *sync.Mutex
	lastFlushTimestamp int64
	backoffSeconds     int64
}

// NewRecorder initializes and returns a new profile recorder
func NewRecorder() *Recorder {
	mq := &Recorder{
		FlushInterval:       5,
		SendProfiles:        NoopSendProfiles,
		MaxBufferedProfiles: DefaultMaxBufferedProfiles,

		queue:     make([]AgentProfile, 0),
		queueLock: &sync.Mutex{},
	}

	return mq
}

// Start initiates profile recorder flush loop
func (pr *Recorder) Start() {
	if !pr.started.SetIfUnset() {
		return
	}

	pr.flushTimer = NewTimer(0, time.Duration(pr.FlushInterval)*time.Second, func() {
		pr.Flush()
	})
}

// Stop terminates profile recorder flush loop
func (pr *Recorder) Stop() {
	if !pr.started.UnsetIfSet() {
		return
	}

	if pr.flushTimer != nil {
		pr.flushTimer.Stop()
	}
}

// Size returns the number of profiles enqueued for submission
func (pr *Recorder) Size() int {
	pr.queueLock.Lock()
	defer pr.queueLock.Unlock()

	return len(pr.queue)
}

// Record stores collected AgentProfile and enqueues it for submission
func (pr *Recorder) Record(record AgentProfile) {
	if pr.MaxBufferedProfiles < 1 {
		return
	}

	pr.queueLock.Lock()
	pr.queue = append(pr.queue, record)
	if len(pr.queue) > pr.MaxBufferedProfiles {
		pr.queue = pr.queue[1:len(pr.queue)]
	}
	pr.queueLock.Unlock()

	logger.Debug("Added record to the queue", record)
}

// Flush forces the recorder to submit collected profiles
func (pr *Recorder) Flush() {
	if pr.Size() == 0 {
		return
	}

	now := time.Now().Unix()

	// flush only if backoff time is elapsed
	if pr.lastFlushTimestamp+pr.backoffSeconds > now {
		return
	}

	pr.queueLock.Lock()
	outgoing := pr.queue
	pr.queue = make([]AgentProfile, 0)
	pr.queueLock.Unlock()

	pr.lastFlushTimestamp = now

	if err := pr.SendProfiles(outgoing); err == nil {
		// reset backoff
		pr.backoffSeconds = 0
	} else {
		// prepend outgoing records back to the queue
		pr.queueLock.Lock()
		pr.queue = append(outgoing, pr.queue...)
		pr.queueLock.Unlock()

		// increase backoff up to 1 minute
		logger.Error("Failed sending profiles, backing off next sending")
		if pr.backoffSeconds == 0 {
			pr.backoffSeconds = 10
		} else if pr.backoffSeconds*2 < 60 {
			pr.backoffSeconds *= 2
		}

		logger.Error(err)
	}
}
