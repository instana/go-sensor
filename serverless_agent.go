// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instana

import (
	"context"
	"errors"
	"os"
	"os/user"
	"strconv"
	"sync"
	"time"

	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/process"
)

// ErrAgentNotReady is an error returned for an attempt to communicate with an agent before the client
// announcement process is done
var ErrAgentNotReady = errors.New("agent not ready")

type containerSnapshot struct {
	ID    string
	Type  string
	Image string
}

type serverlessSnapshot struct {
	EntityID  string
	Host      string
	PID       int
	Container containerSnapshot
	StartedAt time.Time
	Zone      string
	Tags      map[string]interface{}
}

func newProcessPluginPayload(snapshot serverlessSnapshot, prevStats, currentStats processStats) acceptor.PluginPayload {
	var currUser, currGroup string
	if u, err := user.Current(); err == nil {
		currUser = u.Username

		if g, err := user.LookupGroupId(u.Gid); err == nil {
			currGroup = g.Name
		}
	}

	env := getProcessEnv()
	for k := range env {
		if k == "INSTANA_AGENT_KEY" {
			continue
		}

		s := getSensorR()

		if s.options.Tracer.Secrets.Match(k) {
			env[k] = "<redacted>"
		}
	}

	return acceptor.NewProcessPluginPayload(strconv.Itoa(snapshot.PID), acceptor.ProcessData{
		PID:           snapshot.PID,
		Exec:          os.Args[0],
		Args:          os.Args[1:],
		Env:           env,
		User:          currUser,
		Group:         currGroup,
		ContainerID:   snapshot.Container.ID,
		ContainerType: snapshot.Container.Type,
		Start:         snapshot.StartedAt.UnixNano() / int64(time.Millisecond),
		HostName:      snapshot.Host,
		HostPID:       snapshot.PID,
		CPU:           acceptor.NewProcessCPUStatsDelta(prevStats.CPU, currentStats.CPU, currentStats.Tick-prevStats.Tick),
		Memory:        acceptor.NewProcessMemoryStatsUpdate(prevStats.Memory, currentStats.Memory),
		OpenFiles:     acceptor.NewProcessOpenFilesStatsUpdate(prevStats.Limits, currentStats.Limits),
	})
}

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
