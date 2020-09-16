package instana

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"strconv"
	"sync"
	"time"

	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"
	"github.com/instana/go-sensor/aws"
	"github.com/instana/go-sensor/docker"
	"github.com/instana/go-sensor/process"
)

type fargateSnapshot struct {
	EntityID  string
	PID       int
	Task      aws.ECSTaskMetadata
	Container aws.ECSContainerMetadata
}

func newFargateSnapshot(pid int, taskMD aws.ECSTaskMetadata, containerMD aws.ECSContainerMetadata) fargateSnapshot {
	return fargateSnapshot{
		PID:       pid,
		EntityID:  ecsEntityID(containerMD),
		Task:      taskMD,
		Container: containerMD,
	}
}

func newECSTaskPluginPayload(snapshot fargateSnapshot) acceptor.PluginPayload {
	return acceptor.NewECSTaskPluginPayload(snapshot.Task.TaskARN, acceptor.ECSTaskData{
		TaskARN:               snapshot.Task.TaskARN,
		ClusterARN:            snapshot.Container.Cluster,
		AvailabilityZone:      snapshot.Task.AvailabilityZone,
		InstanaZone:           os.Getenv("INSTANA_ZONE"),
		TaskDefinition:        snapshot.Task.Family,
		TaskDefinitionVersion: snapshot.Task.Revision,
		DesiredStatus:         snapshot.Task.DesiredStatus,
		KnownStatus:           snapshot.Task.KnownStatus,
		Limits: acceptor.AWSContainerLimits{
			CPU:    snapshot.Container.Limits.CPU,
			Memory: snapshot.Container.Limits.Memory,
		},
		PullStartedAt: snapshot.Task.PullStartedAt,
		PullStoppedAt: snapshot.Task.PullStoppedAt,
	})
}

func newECSContainerPluginPayload(container aws.ECSContainerMetadata, instrumented bool) acceptor.PluginPayload {
	data := acceptor.ECSContainerData{
		Instrumented:          instrumented,
		DockerID:              container.DockerID,
		DockerName:            container.DockerName,
		ContainerName:         container.Name,
		Image:                 container.Image,
		ImageID:               container.ImageID,
		TaskARN:               container.TaskARN,
		TaskDefinition:        container.TaskDefinition,
		TaskDefinitionVersion: container.TaskDefinitionVersion,
		ClusterARN:            container.Cluster,
		DesiredStatus:         container.DesiredStatus,
		KnownStatus:           container.KnownStatus,
		Limits: acceptor.AWSContainerLimits{
			CPU:    container.Limits.CPU,
			Memory: container.Limits.Memory,
		},
		CreatedAt: container.CreatedAt,
		StartedAt: container.StartedAt,
		Type:      container.Type,
	}

	// we only know the runtime for sure for the instrumented container
	if instrumented {
		data.Runtime = "go"
	}

	return acceptor.NewECSContainerPluginPayload(ecsEntityID(container), data)
}

func newDockerContainerPluginPayload(
	container aws.ECSContainerMetadata,
	prevStats, currentStats docker.ContainerStats,
	instrumented bool,
) acceptor.PluginPayload {

	var networkMode string
	if len(container.Networks) > 0 {
		networkMode = container.Networks[0].Mode
	}

	data := acceptor.DockerData{
		ID:          container.DockerID,
		CreatedAt:   container.CreatedAt,
		StartedAt:   container.StartedAt,
		Image:       container.Image,
		Labels:      container.ContainerLabels,
		Names:       []string{container.DockerName},
		NetworkMode: networkMode,
		Memory:      acceptor.NewDockerMemoryStatsUpdate(prevStats.Memory, currentStats.Memory),
		CPU:         acceptor.NewDockerCPUStatsDelta(prevStats.CPU, currentStats.CPU),
		Network:     acceptor.NewDockerNetworkAggregatedStatsDelta(prevStats.Networks, currentStats.Networks),
		BlockIO:     acceptor.NewDockerBlockIOStatsDelta(prevStats.BlockIO, currentStats.BlockIO),
	}

	// we only know the command for the instrumented container
	if instrumented {
		data.Command = os.Args[0]
	}

	return acceptor.NewDockerPluginPayload(ecsEntityID(container), data)
}

func newProcessPluginPayload(snapshot fargateSnapshot, prevStats, currentStats processStats) acceptor.PluginPayload {
	var currUser, currGroup string
	if u, err := user.Current(); err == nil {
		currUser = u.Username

		if g, err := user.LookupGroupId(u.Gid); err == nil {
			currGroup = g.Name
		}
	}

	return acceptor.NewProcessPluginPayload(strconv.Itoa(snapshot.PID), acceptor.ProcessData{
		PID:           snapshot.PID,
		Exec:          os.Args[0],
		Args:          os.Args[1:],
		Env:           getProcessEnv(),
		User:          currUser,
		Group:         currGroup,
		ContainerID:   snapshot.Container.DockerID,
		ContainerType: "docker",
		Start:         snapshot.Container.StartedAt.UnixNano() / int64(time.Millisecond),
		HostName:      snapshot.Container.TaskARN,
		HostPID:       snapshot.PID,
		CPU:           acceptor.NewProcessCPUStatsDelta(prevStats.CPU, currentStats.CPU, currentStats.Tick-prevStats.Tick),
		Memory:        acceptor.NewProcessMemoryStatsUpdate(prevStats.Memory, currentStats.Memory),
		OpenFiles:     acceptor.NewProcessOpenFilesStatsUpdate(prevStats.Limits, currentStats.Limits),
	})
}

type fargateAgent struct {
	Endpoint string
	Key      string
	PID      int

	snapshot         fargateSnapshot
	lastDockerStats  map[string]docker.ContainerStats
	lastProcessStats processStats

	runtimeSnapshot *SnapshotCollector
	dockerStats     *ecsDockerStatsCollector
	processStats    *processStatsCollector
	client          *http.Client
	ecs             *aws.ECSMetadataProvider
	logger          LeveledLogger
}

func newFargateAgent(
	serviceName, acceptorEndpoint, agentKey string,
	client *http.Client,
	mdProvider *aws.ECSMetadataProvider,
	logger LeveledLogger,
) *fargateAgent {

	if logger == nil {
		logger = defaultLogger
	}

	if client == nil {
		client = http.DefaultClient
	}

	logger.Debug("initializing aws fargate agent")

	agent := &fargateAgent{
		Endpoint: acceptorEndpoint,
		Key:      agentKey,
		PID:      os.Getpid(),
		runtimeSnapshot: &SnapshotCollector{
			CollectionInterval: snapshotCollectionInterval,
			ServiceName:        serviceName,
		},
		dockerStats: &ecsDockerStatsCollector{
			ecs:    mdProvider,
			logger: logger,
		},
		processStats: &processStatsCollector{
			logger: logger,
		},
		client: client,
		ecs:    mdProvider,
		logger: logger,
	}

	go func() {
		for {
			// ECS task metadata publishes the full data (e.g. container.StartedAt)
			// only after a while, so we need to keep trying to gather the full data
			for i := 0; i < maximumRetries; i++ {
				snapshot, ok := agent.collectSnapshot(context.Background())
				if ok {
					agent.snapshot = snapshot
					break
				}

				time.Sleep(retryPeriod)
			}
			time.Sleep(snapshotCollectionInterval)
		}
	}()
	go agent.dockerStats.Run(context.Background(), time.Second)
	go agent.processStats.Run(context.Background(), time.Second)

	return agent
}

func (a *fargateAgent) Ready() bool { return a.snapshot.EntityID != "" }

func (a *fargateAgent) SendMetrics(data acceptor.Metrics) (err error) {
	dockerStats := a.dockerStats.Collect()
	processStats := a.processStats.Collect()
	defer func() {
		if err == nil {
			// only update the last sent stats if they were transmitted successfully
			// since they are updated on the backend incrementally using received
			// deltas
			a.lastDockerStats = dockerStats
			a.lastProcessStats = processStats
		}
	}()

	payload := struct {
		Plugins []acceptor.PluginPayload `json:"plugins"`
	}{
		Plugins: []acceptor.PluginPayload{
			newECSTaskPluginPayload(a.snapshot),
			newProcessPluginPayload(a.snapshot, a.lastProcessStats, processStats),
			acceptor.NewGoProcessPluginPayload(acceptor.GoProcessData{
				PID:      a.PID,
				Snapshot: a.runtimeSnapshot.Collect(),
				Metrics:  data,
			}),
		},
	}

	for _, container := range a.snapshot.Task.Containers {
		instrumented := ecsEntityID(container) == a.snapshot.EntityID
		payload.Plugins = append(
			payload.Plugins,
			newECSContainerPluginPayload(container, instrumented),
			newDockerContainerPluginPayload(
				container,
				a.lastDockerStats[container.DockerID],
				dockerStats[container.DockerID],
				instrumented,
			),
		)
	}

	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(payload); err != nil {
		return fmt.Errorf("failed to marshal metrics payload: %s", err)
	}

	req, err := http.NewRequest(http.MethodPost, a.Endpoint+"/metrics", buf)
	if err != nil {
		return fmt.Errorf("failed to prepare send metrics request: %s", err)
	}

	req.Header.Set("Content-Type", "application/json")

	return a.sendRequest(req)
}

func (a *fargateAgent) SendEvent(event *EventData) error { return nil }

func (a *fargateAgent) SendSpans(spans []Span) error {
	from := newServerlessAgentFromS(a.snapshot.EntityID, "aws")

	agentSpans := make([]agentSpan, 0, len(spans))
	for _, sp := range spans {
		agentSpans = append(agentSpans, agentSpan{sp, from})
	}

	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(agentSpans); err != nil {
		return fmt.Errorf("failed to marshal spans payload: %s", err)
	}

	req, err := http.NewRequest(http.MethodPost, a.Endpoint+"/traces", buf)
	if err != nil {
		return fmt.Errorf("failed to prepare send spans request: %s", err)
	}

	req.Header.Set("Content-Type", "application/json")

	return a.sendRequest(req)
}

func (a *fargateAgent) SendProfiles(profiles []autoprofile.Profile) error { return nil }

func (a *fargateAgent) sendRequest(req *http.Request) error {
	req.Header.Set("X-Instana-Host", a.snapshot.EntityID)
	req.Header.Set("X-Instana-Key", a.Key)
	req.Header.Set("X-Instana-Time", strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10))

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to the serverless agent: %s", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			a.logger.Debug("failed to read serverless agent response: ", err)
			return nil
		}

		a.logger.Info("serverless agent has responded with ", resp.Status, ": ", string(respBody))
		return nil
	}

	io.CopyN(ioutil.Discard, resp.Body, 1<<20)

	return nil
}

func (a *fargateAgent) collectSnapshot(ctx context.Context) (fargateSnapshot, bool) {
	var wg sync.WaitGroup

	// fetch task metadata
	wg.Add(1)
	var taskMD aws.ECSTaskMetadata
	go func() {
		defer wg.Done()

		var err error
		taskMD, err = a.ecs.TaskMetadata(ctx)
		if err != nil {
			a.logger.Warn("failed to get task metadata: ", err)
		}
	}()

	// fetch container metadata
	wg.Add(1)
	var containerMD aws.ECSContainerMetadata
	go func() {
		defer wg.Done()

		var err error
		containerMD, err = a.ecs.ContainerMetadata(ctx)
		if err != nil {
			a.logger.Warn("failed to get container metadata: ", err)
		}
	}()

	wg.Wait()

	// ensure that all metadata has been gathered
	if taskMD.TaskARN == "" || containerMD.StartedAt.IsZero() {
		a.logger.Error("snapshot collection failed (the metadata might not be ready yet)")
		return fargateSnapshot{}, false
	}

	a.logger.Debug("collected snapshot")

	return newFargateSnapshot(a.PID, taskMD, containerMD), true
}

type ecsDockerStatsCollector struct {
	ecs interface {
		TaskStats(context.Context) (map[string]docker.ContainerStats, error)
	}
	logger LeveledLogger

	mu    sync.RWMutex
	stats map[string]docker.ContainerStats
}

func (c *ecsDockerStatsCollector) Run(ctx context.Context, collectionInterval time.Duration) {
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

func (c *ecsDockerStatsCollector) Collect() map[string]docker.ContainerStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.stats
}

func (c *ecsDockerStatsCollector) fetchStats(ctx context.Context) {
	stats, err := c.ecs.TaskStats(ctx)
	if err != nil {
		if ctx.Err() != nil {
			// request either timed out or had been cancelled, keep the old value
			c.logger.Debug("failed to retireve Docker container stats (timed out), skipping")
			return
		}

		// request failed, reset recorded stats
		c.logger.Warn("failed to retrieve Docker container stats: ", err)
		stats = nil
	}

	c.mu.Lock()
	c.stats = stats
	defer c.mu.Unlock()
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

func ecsEntityID(md aws.ECSContainerMetadata) string {
	return md.TaskARN + "::" + md.Name
}
