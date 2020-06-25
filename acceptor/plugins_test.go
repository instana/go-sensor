package acceptor_test

import (
	"testing"

	"github.com/instana/go-sensor/acceptor"
	"github.com/stretchr/testify/assert"
)

func TestNewECSTaskPluginPayload(t *testing.T) {
	data := acceptor.ECSTaskData{
		TaskARN: "arn::task",
	}

	assert.Equal(t, acceptor.PluginPayload{
		Name:     "com.instana.plugin.aws.ecs.task",
		EntityID: "id1",
		Data:     data,
	}, acceptor.NewECSTaskPluginPayload("id1", data))
}

func TestNewECSContainerPluginPayload(t *testing.T) {
	data := acceptor.ECSContainerData{
		DockerID: "docker1",
	}

	assert.Equal(t, acceptor.PluginPayload{
		Name:     "com.instana.plugin.aws.ecs.container",
		EntityID: "id1",
		Data:     data,
	}, acceptor.NewECSContainerPluginPayload("id1", data))
}

func TestNewDockerPluginPayload(t *testing.T) {
	data := acceptor.DockerData{
		ID: "docker1",
	}

	assert.Equal(t, acceptor.PluginPayload{
		Name:     "com.instana.plugin.docker",
		EntityID: "id1",
		Data:     data,
	}, acceptor.NewDockerPluginPayload("id1", data))
}

func TestNewProcessPluginPayload(t *testing.T) {
	data := acceptor.ProcessData{
		PID: 42,
	}

	assert.Equal(t, acceptor.PluginPayload{
		Name:     "com.instana.plugin.process",
		EntityID: "id1",
		Data:     data,
	}, acceptor.NewProcessPluginPayload("id1", data))
}
