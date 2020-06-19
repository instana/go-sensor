package instana

import (
	"context"
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

func newFargateSnapshot(taskMD aws.ECSTaskMetadata, containerMD aws.ECSContainerMetadata) fargateSnapshot {
	return fargateSnapshot{
		EntityID:  ecsEntityID(containerMD),
		Task:      taskMD,
		Container: containerMD,
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

func (a *fargateAgent) SendMetrics(data *MetricsS) error { return nil }

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
