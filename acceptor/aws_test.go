// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

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
