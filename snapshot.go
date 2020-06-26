package instana

import (
	"runtime"
	"sync"
	"time"
)

// SnapshotCollector returns a snapshot of Go runtime
type SnapshotCollector struct {
	ServiceName        string
	CollectionInterval time.Duration

	mu                 sync.RWMutex
	lastCollectionTime time.Time
}

// Collect returns a snaphot of current runtime state. Any call this
// method made before the next interval elapses will return nil
func (sc *SnapshotCollector) Collect() *SnapshotS {
	sc.mu.RLock()
	lastSnapshotCollectionTime := sc.lastCollectionTime
	sc.mu.RUnlock()

	if time.Since(lastSnapshotCollectionTime) < sc.CollectionInterval {
		return nil
	}

	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.lastCollectionTime = time.Now()

	return &SnapshotS{
		Name:     sc.ServiceName,
		Version:  runtime.Version(),
		Root:     runtime.GOROOT(),
		MaxProcs: runtime.GOMAXPROCS(0),
		Compiler: runtime.Compiler,
		NumCPU:   runtime.NumCPU(),
	}
}
