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

func newECSContainerPluginPayload(container aws.ECSContainerMetadata) acceptor.PluginPayload {
	return acceptor.NewECSContainerPluginPayload(ecsEntityID(container), acceptor.ECSContainerData{
		Runtime:               "go",
		Instrumented:          true,
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
	})
}

func newDockerContainerPluginPayload(container aws.ECSContainerMetadata) acceptor.PluginPayload {
	var networkMode string
	if len(container.Networks) > 0 {
		networkMode = container.Networks[0].Mode
	}

	return acceptor.NewDockerPluginPayload(ecsEntityID(container), acceptor.DockerData{
		ID:          container.DockerID,
		Command:     os.Args[0],
		CreatedAt:   container.CreatedAt,
		StartedAt:   container.StartedAt,
		Image:       container.Image,
		Labels:      container.ContainerLabels,
		Names:       []string{container.DockerName},
		NetworkMode: networkMode,
		Memory:      container.Limits.Memory,
	})
}

func newProcessPluginPayload(snapshot fargateSnapshot) acceptor.PluginPayload {
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
	})
}

type fargateAgent struct {
	Endpoint string
	Key      string
	PID      int

	snapshot fargateSnapshot

	runtimeSnapshot *SnapshotCollector
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

	return agent
}

func (a *fargateAgent) Ready() bool { return a.snapshot.EntityID != "" }

func (a *fargateAgent) SendMetrics(data acceptor.Metrics) error {
	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(struct {
		Plugins []acceptor.PluginPayload `json:"plugins"`
	}{
		Plugins: []acceptor.PluginPayload{
			newECSTaskPluginPayload(a.snapshot),
			newECSContainerPluginPayload(a.snapshot.Container),
			newDockerContainerPluginPayload(a.snapshot.Container),
			newProcessPluginPayload(a.snapshot),
			acceptor.NewGoProcessPluginPayload(acceptor.GoProcessData{
				PID:      a.PID,
				Snapshot: a.runtimeSnapshot.Collect(),
				Metrics:  data,
			}),
		},
	},
	); err != nil {
		return fmt.Errorf("failed to marshal metrics payload: %s", err)
	}

	req, err := http.NewRequest(http.MethodPost, a.Endpoint+"/metrics", buf)
	if err != nil {
		return fmt.Errorf("failed to prepare send metrics request: %s", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Instana-Host", a.snapshot.EntityID)
	req.Header.Set("X-Instana-Key", a.Key)
	req.Header.Set("X-Instana-Time", strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10))

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send metrics: %s", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			a.logger.Debug("failed to read server response: ", err)
			return nil
		}

		a.logger.Info("acceptor has responded with ", resp.Status, ": ", string(respBody))
		return nil
	}

	io.CopyN(ioutil.Discard, resp.Body, 1<<20)

	return nil
}

func (a *fargateAgent) SendEvent(event *EventData) error { return nil }

func (a *fargateAgent) SendSpans(spans []Span) error {
	return nil
}

func (a *fargateAgent) SendProfiles(profiles []autoprofile.Profile) error { return nil }

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

func ecsEntityID(md aws.ECSContainerMetadata) string {
	return md.TaskARN + "::" + md.Name
}
