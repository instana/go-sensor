package instana

import (
	"context"
	"os"
	"os/user"
	"strconv"
	"sync"
	"time"

	"github.com/instana/go-sensor/autoprofile"
	"github.com/instana/go-sensor/aws"
)

type fargateSnapshot struct {
	EntityID  string
	Task      aws.ECSTaskMetadata
	Container aws.ECSContainerMetadata
}

type awsContainerLimits struct {
	CPU    int `json:"cpu"`
	Memory int `json:"memory"`
}

func newFargateSnapshot(taskMD aws.ECSTaskMetadata, containerMD aws.ECSContainerMetadata) fargateSnapshot {
	return fargateSnapshot{
		EntityID:  ecsEntityID(containerMD),
		Task:      taskMD,
		Container: containerMD,
	}
}

type acceptorPluginPayload struct {
	Name     string      `json:"name"`
	EntityID string      `json:"entityId"`
	Data     interface{} `json:"data"`
}

type ecsTaskPluginData struct {
	TaskARN               string             `json:"taskArn"`
	ClusterARN            string             `json:"clusterArn"`
	TaskDefinition        string             `json:"taskDefinition"`
	TaskDefinitionVersion string             `json:"taskDefinitionVersion"`
	DesiredStatus         string             `json:"desiredStatus"`
	KnownStatus           string             `json:"knownStatus"`
	Limits                awsContainerLimits `json:"limits"`
	PullStartedAt         time.Time          `json:"pullStartedAt"`
	PullStoppedAt         time.Time          `json:"pullStoppedAt"`
}

func newECSTaskPluginPayload(snapshot fargateSnapshot) acceptorPluginPayload {
	const pluginName = "com.instana.plugin.aws.ecs.task"

	return acceptorPluginPayload{
		Name:     pluginName,
		EntityID: snapshot.Task.TaskARN,
		Data: ecsTaskPluginData{
			TaskARN:               snapshot.Task.TaskARN,
			ClusterARN:            snapshot.Container.Cluster,
			TaskDefinition:        snapshot.Task.Family,
			TaskDefinitionVersion: snapshot.Task.Revision,
			DesiredStatus:         snapshot.Task.DesiredStatus,
			KnownStatus:           snapshot.Task.KnownStatus,
			Limits: awsContainerLimits{
				CPU:    snapshot.Container.Limits.CPU,
				Memory: snapshot.Container.Limits.Memory,
			},
			PullStartedAt: snapshot.Task.PullStartedAt,
			PullStoppedAt: snapshot.Task.PullStoppedAt,
		},
	}
}

type ecsContainerPluginData struct {
	Runtime               string             `json:"runtime"`
	Instrumented          bool               `json:"instrumented,omitempty"`
	DockerID              string             `json:"dockerId"`
	DockerName            string             `json:"dockerName"`
	ContainerName         string             `json:"containerName"`
	Image                 string             `json:"image"`
	ImageID               string             `json:"imageId"`
	TaskARN               string             `json:"taskArn"`
	TaskDefinition        string             `json:"taskDefinition"`
	TaskDefinitionVersion string             `json:"taskDefinitionVersion"`
	ClusterARN            string             `json:"clusterArn"`
	DesiredStatus         string             `json:"desiredStatus"`
	KnownStatus           string             `json:"knownStatus"`
	Limits                awsContainerLimits `json:"limits"`
	CreatedAt             time.Time          `json:"createdAt"`
	StartedAt             time.Time          `json:"startedAt"`
	Type                  string             `json:"type"`
}

func newECSContainerPluginPayload(container aws.ECSContainerMetadata) acceptorPluginPayload {
	const pluginName = "com.instana.plugin.aws.ecs.container"

	return acceptorPluginPayload{
		Name:     pluginName,
		EntityID: ecsEntityID(container),
		Data: ecsContainerPluginData{
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
			Limits: awsContainerLimits{
				CPU:    container.Limits.CPU,
				Memory: container.Limits.Memory,
			},
			CreatedAt: container.CreatedAt,
			StartedAt: container.StartedAt,
			Type:      container.Type,
		},
	}
}

type dockerContainerPluginData struct {
	ID               string              `json:"Id"`
	Command          string              `json:"Command"`
	CreatedAt        time.Time           `json:"Created"`
	StartedAt        time.Time           `json:"Started"`
	Image            string              `json:"Image"`
	Labels           aws.ContainerLabels `json:"Labels,omitempty"`
	Ports            string              `json:"Ports,omitempty"`
	PortBindings     string              `json:"PortBindings,omitempty"`
	Names            []string            `json:"Names,omitempty"`
	NetworkMode      string              `json:"NetworkMode,omitempty"`
	StorageDriver    string              `json:"StorageDriver,omitempty"`
	DockerVersion    string              `json:"docker_version,omitempty"`
	DockerAPIVersion string              `json:"docker_api_version,omitempty"`
	Memory           int                 `json:"Memory"`
}

func newDockerContainerPluginPayload(container aws.ECSContainerMetadata) acceptorPluginPayload {
	const pluginName = "com.instana.plugin.docker"

	var networkMode string
	if len(container.Networks) > 0 {
		networkMode = container.Networks[0].Mode
	}

	return acceptorPluginPayload{
		Name:     pluginName,
		EntityID: ecsEntityID(container),
		Data: dockerContainerPluginData{
			ID:          container.DockerID,
			Command:     os.Args[0],
			CreatedAt:   container.CreatedAt,
			StartedAt:   container.StartedAt,
			Image:       container.Image,
			Labels:      container.ContainerLabels,
			Names:       []string{container.DockerName},
			NetworkMode: networkMode,
			Memory:      container.Limits.Memory,
		},
	}
}

type processPluginData struct {
	PID           int               `json:"pid"`
	Exec          string            `json:"exec"`
	Args          []string          `json:"args,omitempty"`
	Env           map[string]string `json:"env,omitempty"`
	User          string            `json:"user,omitempty"`
	Group         string            `json:"group,omitempty"`
	ContainerID   string            `json:"container,omitempty"`
	ContainerPid  int               `json:"containerPid,string,omitempty"`
	ContainerType string            `json:"containerType,omitempty"`
	Start         int64             `json:"start"`
	HostName      string            `json:"com.instana.plugin.host.name"`
	HostPID       int               `json:"com.instana.plugin.host.pid,string"`
}

func newProcessPluginPayload(snapshot fargateSnapshot) acceptorPluginPayload {
	const pluginName = "com.instana.plugin.process"

	pid := os.Getpid()

	var currUser, currGroup string
	if u, err := user.Current(); err == nil {
		currUser = u.Username

		if g, err := user.LookupGroupId(u.Gid); err == nil {
			currGroup = g.Name
		}
	}

	return acceptorPluginPayload{
		Name:     pluginName,
		EntityID: strconv.Itoa(pid),
		Data: processPluginData{
			PID:           pid,
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
		},
	}
}

func newGolangPluginPayload(data interface{}) acceptorPluginPayload {
	const pluginName = "com.instana.plugin.golang"

	pid := os.Getpid()

	return acceptorPluginPayload{
		Name:     pluginName,
		EntityID: strconv.Itoa(pid),
		Data:     data,
	}
}

type fargateAgent struct {
	Endpoint string
	Key      string

	snapshot fargateSnapshot

	ecs    *aws.ECSMetadataProvider
	logger LeveledLogger
}

func newFargateAgent(acceptorEndpoint, agentKey string, mdProvider *aws.ECSMetadataProvider, logger LeveledLogger) *fargateAgent {
	if logger == nil {
		logger = defaultLogger
	}

	logger.Debug("initializing aws fargate agent")

	agent := &fargateAgent{
		Endpoint: acceptorEndpoint,
		Key:      agentKey,
		ecs:      mdProvider,
		logger:   logger,
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

func (a *fargateAgent) SendMetrics(data *MetricsS) error                  { return nil }
func (a *fargateAgent) SendEvent(event *EventData) error                  { return nil }
func (a *fargateAgent) SendSpans(spans []Span) error                      { return nil }
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

	return newFargateSnapshot(taskMD, containerMD), true
}

func ecsEntityID(md aws.ECSContainerMetadata) string {
	return md.TaskARN + "::" + md.Name
}
