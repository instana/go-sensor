// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instana

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"
	"github.com/instana/go-sensor/aws"
	"github.com/instana/go-sensor/docker"
)

type fargateSnapshot struct {
	Service   serverlessSnapshot
	Task      aws.ECSTaskMetadata
	Container aws.ECSContainerMetadata
}

func newFargateSnapshot(pid int, taskMD aws.ECSTaskMetadata, containerMD aws.ECSContainerMetadata) fargateSnapshot {
	return fargateSnapshot{
		Service: serverlessSnapshot{
			EntityID:  ecsEntityID(containerMD),
			Host:      containerMD.TaskARN,
			PID:       pid,
			StartedAt: processStartedAt,
			Container: containerSnapshot{
				ID:    containerMD.DockerID,
				Type:  "docker",
				Image: containerMD.Image,
			},
		},
		Task:      taskMD,
		Container: containerMD,
	}
}

func newECSTaskPluginPayload(snapshot fargateSnapshot) acceptor.PluginPayload {
	return acceptor.NewECSTaskPluginPayload(snapshot.Task.TaskARN, acceptor.ECSTaskData{
		TaskARN:               snapshot.Task.TaskARN,
		ClusterARN:            snapshot.Container.Cluster,
		AvailabilityZone:      snapshot.Task.AvailabilityZone,
		InstanaZone:           snapshot.Service.Zone,
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
		Tags:          snapshot.Service.Tags,
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

type metricsPayload struct {
	Plugins []acceptor.PluginPayload `json:"plugins"`
}

type fargateAgent struct {
	Endpoint string
	Key      string
	PID      int
	Zone     string
	Tags     map[string]interface{}

	snapshot         fargateSnapshot
	lastDockerStats  map[string]docker.ContainerStats
	lastProcessStats processStats

	mu        sync.Mutex
	spanQueue []Span

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
		Zone:     os.Getenv("INSTANA_ZONE"),
		Tags:     parseInstanaTags(os.Getenv("INSTANA_TAGS")),
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

				time.Sleep(expDelay(i + 1))
			}
			time.Sleep(snapshotCollectionInterval)
		}
	}()
	go agent.dockerStats.Run(context.Background(), time.Second)
	go agent.processStats.Run(context.Background(), time.Second)

	return agent
}

func (a *fargateAgent) Ready() bool { return a.snapshot.Service.EntityID != "" }

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
		Metrics metricsPayload `json:"metrics,omitempty"`
		Spans   []Span         `json:"spans,omitempty"`
	}{
		Metrics: metricsPayload{
			Plugins: []acceptor.PluginPayload{
				newECSTaskPluginPayload(a.snapshot),
				newProcessPluginPayload(a.snapshot.Service, a.lastProcessStats, processStats),
				acceptor.NewGoProcessPluginPayload(acceptor.GoProcessData{
					PID:      a.PID,
					Snapshot: a.runtimeSnapshot.Collect(),
					Metrics:  data,
				}),
			},
		},
	}

	for _, container := range a.snapshot.Task.Containers {
		instrumented := ecsEntityID(container) == a.snapshot.Service.EntityID
		payload.Metrics.Plugins = append(
			payload.Metrics.Plugins,
			newECSContainerPluginPayload(container, instrumented),
			newDockerContainerPluginPayload(
				container,
				a.lastDockerStats[container.DockerID],
				dockerStats[container.DockerID],
				instrumented,
			),
		)
	}

	a.mu.Lock()
	if len(a.spanQueue) > 0 {
		payload.Spans = make([]Span, len(a.spanQueue))
		copy(payload.Spans, a.spanQueue)
		a.spanQueue = a.spanQueue[:0]
	}
	a.mu.Unlock()

	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(payload); err != nil {
		return fmt.Errorf("failed to marshal metrics payload: %s", err)
	}

	req, err := http.NewRequest(http.MethodPost, a.Endpoint+"/bundle", buf)
	if err != nil {
		return fmt.Errorf("failed to prepare send metrics request: %s", err)
	}

	req.Header.Set("Content-Type", "application/json")

	return a.sendRequest(req)
}

func (a *fargateAgent) SendEvent(event *EventData) error { return nil }

func (a *fargateAgent) SendSpans(spans []Span) error {
	from := newServerlessAgentFromS(a.snapshot.Service.EntityID, "aws")
	for i := range spans {
		spans[i].From = from
	}

	// enqueue the spans to send them in a bundle with metrics instead of sending immediately
	a.mu.Lock()
	a.spanQueue = append(a.spanQueue, spans...)
	a.mu.Unlock()

	return nil
}

func (a *fargateAgent) SendProfiles(profiles []autoprofile.Profile) error { return nil }

func (a *fargateAgent) Flush(ctx context.Context) error {
	if len(a.spanQueue) == 0 {
		return nil
	}

	if !a.Ready() {
		return ErrAgentNotReady
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(a.spanQueue); err != nil {
		return fmt.Errorf("failed to marshal traces payload: %s", err)
	}
	a.spanQueue = a.spanQueue[:0]

	req, err := http.NewRequest(http.MethodPost, a.Endpoint+"/traces", buf)
	if err != nil {
		return fmt.Errorf("failed to prepare send traces request: %s", err)
	}

	req.Header.Set("Content-Type", "application/json")

	return a.sendRequest(req.WithContext(ctx))
}

func (a *fargateAgent) sendRequest(req *http.Request) error {
	req.Header.Set("X-Instana-Host", a.snapshot.Service.EntityID)
	req.Header.Set("X-Instana-Key", a.Key)
	req.Header.Set("X-Instana-Time", strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10))

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to the serverless agent: %s", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			a.logger.Debug("failed to read serverless agent response: ", err)
			return nil
		}

		a.logger.Info("serverless agent has responded with ", resp.Status, ": ", string(respBody))
		return nil
	}

	io.CopyN(io.Discard, resp.Body, 1<<20)

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

	snapshot := newFargateSnapshot(a.PID, taskMD, containerMD)
	snapshot.Service.Zone = a.Zone
	snapshot.Service.Tags = a.Tags

	a.logger.Debug("collected snapshot")

	return snapshot, true
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
			c.logger.Debug("failed to retrieve Docker container stats (timed out), skipping")
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

func ecsEntityID(md aws.ECSContainerMetadata) string {
	return md.TaskARN + "::" + md.Name
}
