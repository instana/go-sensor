// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

//go:build gcr && integration
// +build gcr,integration

package instana_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	instana "github.com/instana/go-sensor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	instana.InitCollector(instana.DefaultOptions())
	defer instana.ShutdownCollector()

	os.Exit(m.Run())
}

func TestIntegration_GCRAgent_SendMetrics(t *testing.T) {
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

	t.Run("GCR service revision instance plugin payload", func(t *testing.T) {
		require.Len(t, pluginData["com.instana.plugin.gcp.run.revision.instance"], 1)
		d := pluginData["com.instana.plugin.gcp.run.revision.instance"][0]

		assert.Equal(t, "id1", d.EntityID)

		assert.Equal(t, "go", d.Data["runtime"])
		assert.Equal(t, "us-central1", d.Data["region"])
		assert.Equal(t, "test-service", d.Data["service"])
		assert.Equal(t, "test-configuration", d.Data["configuration"])
		assert.Equal(t, "test-revision", d.Data["revision"])
		assert.Equal(t, "id1", d.Data["instanceId"])
		assert.Equal(t, "8081", d.Data["port"])
		assert.EqualValues(t, 1234567890, d.Data["numericProjectId"])
		assert.Equal(t, "test-project", d.Data["projectId"])
	})

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

func TestIntegration_GCRAgent_SendSpans(t *testing.T) {
	defer agent.Reset()

	c := instana.InitCollector(instana.DefaultOptions())
	defer instana.ShutdownCollector()

	sp := c.Tracer().StartSpan("entry")
	sp.SetTag("value", "42")
	sp.Finish()

	require.Eventually(t, func() bool {
		agent.mu.Lock()
		defer agent.mu.Unlock()
		if len(agent.Bundles) == 0 {
			fmt.Println("false1")
			return false
		}

		for i, bundle := range agent.Bundles {
			fmt.Println("bundle.Body", i, string(bundle.Body))
			var payload struct {
				Spans []json.RawMessage `json:"spans"`
			}

			json.Unmarshal(bundle.Body, &payload)
			if len(payload.Spans) > 0 {
				fmt.Println("true1")
				return true
			}
		}
		// fmt.Println("false2", agent.Bundles)
		fmt.Println("false2")
		return false
	}, 20*time.Second, 500*time.Millisecond)

	var spans []map[string]json.RawMessage
	for _, bundle := range agent.Bundles {
		var payload struct {
			Spans []map[string]json.RawMessage `json:"spans"`
		}

		require.NoError(t, json.Unmarshal(bundle.Body, &payload), "%s", string(bundle.Body))
		spans = append(spans, payload.Spans...)
	}

	require.Len(t, spans, 1)
	assert.JSONEq(t, `{"hl": true, "cp": "gcp", "e": "id1"}`, string(spans[0]["f"]))
}

func TestIntegration_GCRAgent_FlushSpans(t *testing.T) {
	defer agent.Reset()

	c := instana.InitCollector(instana.DefaultOptions())
	defer instana.ShutdownCollector()

	sp := c.Tracer().StartSpan("entry")
	sp.SetTag("value", "42")
	sp.Finish()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	require.NoError(t, c.Flush(ctx))
}

func setupGCREnv() func() {
	var teardownFns []func()

	teardownFns = append(teardownFns, restoreEnvVarFunc("K_SERVICE"))
	os.Setenv("K_SERVICE", "test-service")

	teardownFns = append(teardownFns, restoreEnvVarFunc("K_CONFIGURATION"))
	os.Setenv("K_CONFIGURATION", "test-configuration")

	teardownFns = append(teardownFns, restoreEnvVarFunc("K_REVISION"))
	os.Setenv("K_REVISION", "test-revision")

	teardownFns = append(teardownFns, restoreEnvVarFunc("PORT"))
	os.Setenv("PORT", "8081")

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
