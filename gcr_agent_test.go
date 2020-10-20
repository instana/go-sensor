// +build gcr_integration

package instana_test

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	instana "github.com/instana/go-sensor"
	"github.com/instana/testify/assert"
	"github.com/instana/testify/require"
)

var agent *serverlessAgent

func TestMain(m *testing.M) {
	teardownEnv := setupGCREnv()
	defer teardownEnv()

	teardownSrv := setupMetadataServer()
	defer teardownSrv()

	defer restoreEnvVarFunc("INSTANA_AGENT_KEY")
	os.Setenv("INSTANA_AGENT_KEY", "testkey1")

	defer restoreEnvVarFunc("INSTANA_ZONE")
	os.Setenv("INSTANA_ZONE", "testzone")

	defer restoreEnvVarFunc("INSTANA_TAGS")
	os.Setenv("INSTANA_TAGS", "key1=value1,key2")

	defer restoreEnvVarFunc("INSTANA_SECRETS")
	os.Setenv("INSTANA_SECRETS", "contains-ignore-case:key,password,secret,classified")

	defer restoreEnvVarFunc("CLASSIFIED_DATA")
	os.Setenv("CLASSIFIED_DATA", "classified")

	var err error
	agent, err = setupServerlessAgent()
	if err != nil {
		log.Fatalf("failed to initialize serverless agent: %s", err)
	}

	instana.InitSensor(instana.DefaultOptions())

	os.Exit(m.Run())
}

func TestGCRAgent_SendMetrics(t *testing.T) {
	defer agent.Reset()

	require.Eventually(t, func() bool { return len(agent.Bundles) > 0 }, 2*time.Second, 500*time.Millisecond)

	collected := agent.Bundles[0]

	assert.Equal(t, "gcp:cloud-run:revision:test-revision", collected.Header.Get("X-Instana-Host"))
	assert.Equal(t, "testkey1", collected.Header.Get("X-Instana-Key"))
	assert.NotEmpty(t, collected.Header.Get("X-Instana-Time"))

	var payload struct {
		Metrics struct {
			Plugins []struct {
				Name     string                 `json:"name"`
				EntityID string                 `json:"entityId"`
				Data     map[string]interface{} `json:"data"`
			} `json:"plugins"`
		} `json:"metrics"`
	}
	require.NoError(t, json.Unmarshal(collected.Body, &payload))

	pluginData := make(map[string][]serverlessAgentPluginPayload)
	for _, plugin := range payload.Metrics.Plugins {
		pluginData[plugin.Name] = append(pluginData[plugin.Name], serverlessAgentPluginPayload{plugin.EntityID, plugin.Data})
	}

	t.Run("Process plugin payload", func(t *testing.T) {
		require.Len(t, pluginData["com.instana.plugin.process"], 1)
		d := pluginData["com.instana.plugin.process"][0]

		assert.NotEmpty(t, d.EntityID)

		assert.Equal(t, "gcpCloudRunInstance", d.Data["containerType"])
		assert.Equal(t, "id1", d.Data["container"])
		assert.Equal(t, "gcp:cloud-run:revision:test-revision", d.Data["com.instana.plugin.host.name"])

		if assert.IsType(t, map[string]interface{}{}, d.Data["env"]) {
			env := d.Data["env"].(map[string]interface{})

			assert.Equal(t, os.Getenv("INSTANA_ZONE"), env["INSTANA_ZONE"])
			assert.Equal(t, os.Getenv("INSTANA_TAGS"), env["INSTANA_TAGS"])
			assert.Equal(t, os.Getenv("INSTANA_AGENT_KEY"), env["INSTANA_AGENT_KEY"])

			assert.Equal(t, "<redacted>", env["INSTANA_SECRETS"])
			assert.Equal(t, "<redacted>", env["CLASSIFIED_DATA"])
		}
	})

	t.Run("Go process plugin payload", func(t *testing.T) {
		require.Len(t, pluginData["com.instana.plugin.golang"], 1)
		d := pluginData["com.instana.plugin.golang"][0]

		assert.NotEmpty(t, d.EntityID)

		require.NotEmpty(t, pluginData["com.instana.plugin.process"])
		assert.Equal(t, pluginData["com.instana.plugin.process"][0].EntityID, d.EntityID)

		assert.NotEmpty(t, d.Data["metrics"])
	})
}

func setupGCREnv() func() {
	var teardownFns []func()

	teardownFns = append(teardownFns, restoreEnvVarFunc("K_SERVICE"))
	os.Setenv("K_SERVICE", "test-service")

	teardownFns = append(teardownFns, restoreEnvVarFunc("K_CONFIGURATION"))
	os.Setenv("K_CONFIGURATION", "test-configuration")

	teardownFns = append(teardownFns, restoreEnvVarFunc("K_REVISION"))
	os.Setenv("K_REVISION", "test-revision")

	return func() {
		for _, fn := range teardownFns {
			fn()
		}
	}
}

func setupMetadataServer() func() {
	mux := http.NewServeMux()
	mux.HandleFunc("/computeMetadata/v1", func(w http.ResponseWriter, req *http.Request) {
		http.ServeFile(w, req, "gcloud/testdata/computeMetadata.json")
	})

	srv := httptest.NewServer(mux)

	teardown := restoreEnvVarFunc("GOOGLE_CLOUD_RUN_METADATA_ENDPOINT")
	os.Setenv("GOOGLE_CLOUD_RUN_METADATA_ENDPOINT", srv.URL)

	return func() {
		teardown()
		srv.Close()
	}
}
