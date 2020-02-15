package autoprofile

import (
	"sync"
	"time"
)

type ProfileRecorder struct {
	FlushInterval int64

	profiler *AutoProfiler
	started  *Flag

	flushTimer *Timer

	queue              []interface{}
	queueLock          *sync.Mutex
	lastFlushTimestamp int64
	backoffSeconds     int64
}

func newProfileRecorder(profiler *AutoProfiler) *ProfileRecorder {
	mq := &ProfileRecorder{
		FlushInterval: 5,

		profiler: profiler,
		started:  &Flag{},

		flushTimer: nil,

		queue:              make([]interface{}, 0),
		queueLock:          &sync.Mutex{},
		lastFlushTimestamp: 0,
		backoffSeconds:     0,
	}

	return mq
}

func (pr *ProfileRecorder) start() {
	if !pr.started.SetIfUnset() {
		return
	}

	pr.flushTimer = createTimer(0, time.Duration(pr.FlushInterval)*time.Second, func() {
		pr.flush()
	})
}

func (pr *ProfileRecorder) stop() {
	if !pr.started.UnsetIfSet() {
		return
	}

	if pr.flushTimer != nil {
		pr.flushTimer.Stop()
	}
}

func (pr *ProfileRecorder) size() int {
	pr.queueLock.Lock()
	defer pr.queueLock.Unlock()

	return len(pr.queue)
}

func (pr *ProfileRecorder) record(record map[string]interface{}) {
	pr.queueLock.Lock()
	pr.queue = append(pr.queue, record)
	if len(pr.queue) > pr.profiler.MaxBufferedProfiles {
		pr.queue = pr.queue[1:len(pr.queue)]
	}
	pr.queueLock.Unlock()

	log.debug("Added record to the queue", record)
}

func (pr *ProfileRecorder) flush() {
	if pr.size() == 0 {
		return
	}

	now := time.Now().Unix()

	// flush only if backoff time is elapsed
	if pr.lastFlushTimestamp+pr.backoffSeconds > now {
		return
	}

	pr.queueLock.Lock()
	outgoing := pr.queue
	pr.queue = make([]interface{}, 0)
	pr.queueLock.Unlock()

	pr.lastFlushTimestamp = now

	if err := pr.profiler.SendProfiles(outgoing); err == nil {
		// reset backoff
		pr.backoffSeconds = 0
	} else {
		// prepend outgoing records back to the queue
		pr.queueLock.Lock()
		pr.queue = append(outgoing, pr.queue...)
		pr.queueLock.Unlock()

		// increase backoff up to 1 minute
		log.error("Failed sending profiles, backing off next sending")
		if pr.backoffSeconds == 0 {
			pr.backoffSeconds = 10
		} else if pr.backoffSeconds*2 < 60 {
			pr.backoffSeconds *= 2
		}

		log.error(err)
	}
}
