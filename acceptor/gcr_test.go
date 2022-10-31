// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package acceptor_test

import (
	"testing"

	"github.com/instana/go-sensor/acceptor"
	"github.com/stretchr/testify/assert"
)

func TestNewGCRServiceRevisionInstancePluginPayload(t *testing.T) {
	data := acceptor.GCRServiceRevisionInstanceData{
		Region:           "test-region",
		Service:          "test-server",
		Revision:         "test-revision",
		InstanceID:       "test-instance",
		NumericProjectID: 42,
	}

	assert.Equal(t, acceptor.PluginPayload{
		Name:     "com.instana.plugin.gcp.run.revision.instance",
		EntityID: "id1",
		Data:     data,
	}, acceptor.NewGCRServiceRevisionInstancePluginPayload("id1", data))
}
