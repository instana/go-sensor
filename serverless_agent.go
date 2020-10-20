package instana

import (
	"context"
	"sync"
	"time"

	"github.com/instana/go-sensor/process"
)

type processStats struct {
	Tick   int
	CPU    process.CPUStats
	Memory process.MemStats
	Limits process.ResourceLimits
}

type processStatsCollector struct {
	logger LeveledLogger
	mu     sync.RWMutex
	stats  processStats
}

func (c *processStatsCollector) Run(ctx context.Context, collectionInterval time.Duration) {
	timer := time.NewTicker(collectionInterval)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			fetchCtx, cancel := context.WithTimeout(ctx, collectionInterval)
			c.fetchStats(fetchCtx)
			cancel()
		case <-ctx.Done():
			return
		}
	}
}

func (c *processStatsCollector) Collect() processStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.stats
}

func (c *processStatsCollector) fetchStats(ctx context.Context) {
	stats := c.Collect()

	var wg sync.WaitGroup
	wg.Add(3)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	go func() {
		defer wg.Done()

		st, tick, err := process.Stats().CPU()
		if err != nil {
			c.logger.Debug("failed to read process CPU stats, skipping: ", err)
			return
		}

		stats.CPU, stats.Tick = st, tick
	}()

	go func() {
		defer wg.Done()

		st, err := process.Stats().Memory()
		if err != nil {
			c.logger.Debug("failed to read process memory stats, skipping: ", err)
			return
		}

		stats.Memory = st
	}()

	go func() {
		defer wg.Done()

		st, err := process.Stats().Limits()
		if err != nil {
			c.logger.Debug("failed to read process open files stats, skipping: ", err)
			return
		}

		stats.Limits = st
	}()

	select {
	case <-done:
		break
	case <-ctx.Done():
		c.logger.Debug("failed to obtain process stats (timed out)")
		return // context has been cancelled, skip this update
	}

	c.mu.Lock()
	c.stats = stats
	defer c.mu.Unlock()
}
