// (c) Copyright IBM Corp. 2022
// (c) Copyright Instana Inc. 2022

package acceptor_test

import (
	"testing"

	"github.com/instana/go-sensor/acceptor"
	"github.com/stretchr/testify/assert"
)

func TestAzurePluginPayload(t *testing.T) {
	assert.Equal(t,
		acceptor.PluginPayload{
			Name:     "com.instana.plugin.azure.functionapp",
			EntityID: "test-entity-id",
		},
		acceptor.NewAzurePluginPayload("test-entity-id"))
}
