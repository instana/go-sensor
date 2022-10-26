// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

//go:build fargate && integration
// +build fargate,integration

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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var agent *serverlessAgent

func TestMain(m *testing.M) {
	teardownEnv := setupAWSFargateEnv()
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

func TestFargateAgent_SendMetrics(t *testing.T) {
	defer agent.Reset()

	require.Eventually(t, func() bool { return len(agent.Bundles) > 0 }, 2*time.Second, 500*time.Millisecond)

	collected := agent.Bundles[0]

	assert.Equal(t, "arn:aws:ecs:us-east-2:012345678910:task/9781c248-0edd-4cdb-9a93-f63cb662a5d3::nginx-curl", collected.Header.Get("X-Instana-Host"))
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

	t.Run("AWS ECS Task plugin payload", func(t *testing.T) {
		require.Len(t, pluginData["com.instana.plugin.aws.ecs.task"], 1)
		d := pluginData["com.instana.plugin.aws.ecs.task"][0]

		assert.NotEmpty(t, d.EntityID)
		assert.Equal(t, d.Data["taskArn"], d.EntityID)

		assert.Equal(t, "testzone", d.Data["instanaZone"])
		assert.Equal(t, map[string]interface{}{"key1": "value1", "key2": nil}, d.Data["tags"])
		assert.Equal(t, "default", d.Data["clusterArn"])
		assert.Equal(t, "nginx", d.Data["taskDefinition"])
		assert.Equal(t, "5", d.Data["taskDefinitionVersion"])
	})

	t.Run("AWS ECS Container plugin payload", func(t *testing.T) {
		require.Len(t, pluginData["com.instana.plugin.aws.ecs.container"], 2)

		containers := make(map[string]serverlessAgentPluginPayload)
		for _, container := range pluginData["com.instana.plugin.aws.ecs.container"] {
			containers[container.EntityID] = container
		}

		t.Run("instrumented", func(t *testing.T) {
			d := containers["arn:aws:ecs:us-east-2:012345678910:task/9781c248-0edd-4cdb-9a93-f63cb662a5d3::nginx-curl"]
			require.NotEmpty(t, d)

			assert.NotEmpty(t, d.EntityID)

			require.IsType(t, d.Data["taskArn"], "")
			require.IsType(t, d.Data["containerName"], "")
			assert.Equal(t, d.Data["taskArn"].(string)+"::"+d.Data["containerName"].(string), d.EntityID)

			if assert.NotEmpty(t, d.Data["taskArn"]) {
				require.NotEmpty(t, pluginData["com.instana.plugin.aws.ecs.task"])
				assert.Equal(t, pluginData["com.instana.plugin.aws.ecs.task"][0].EntityID, d.Data["taskArn"])
			}

			assert.Equal(t, true, d.Data["instrumented"])
			assert.Equal(t, "go", d.Data["runtime"])
			assert.Equal(t, "43481a6ce4842eec8fe72fc28500c6b52edcc0917f105b83379f88cac1ff3946", d.Data["dockerId"])
		})

		t.Run("non-instrumented", func(t *testing.T) {
			d := containers["arn:aws:ecs:us-east-2:012345678910:task/9781c248-0edd-4cdb-9a93-f63cb662a5d3::~internal~ecs~pause"]
			require.NotEmpty(t, d)

			assert.NotEmpty(t, d.EntityID)

			require.IsType(t, d.Data["taskArn"], "")
			require.IsType(t, d.Data["containerName"], "")
			assert.Equal(t, d.Data["taskArn"].(string)+"::"+d.Data["containerName"].(string), d.EntityID)

			if assert.NotEmpty(t, d.Data["taskArn"]) {
				require.NotEmpty(t, pluginData["com.instana.plugin.aws.ecs.task"])
				assert.Equal(t, pluginData["com.instana.plugin.aws.ecs.task"][0].EntityID, d.Data["taskArn"])
			}

			assert.Nil(t, d.Data["instrumented"])
			assert.Empty(t, d.Data["runtime"])
			assert.Equal(t, "731a0d6a3b4210e2448339bc7015aaa79bfe4fa256384f4102db86ef94cbbc4c", d.Data["dockerId"])
		})
	})

	t.Run("Docker plugin payload", func(t *testing.T) {
		require.Len(t, pluginData["com.instana.plugin.docker"], 2)

		containers := make(map[string]serverlessAgentPluginPayload)
		for _, container := range pluginData["com.instana.plugin.docker"] {
			containers[container.EntityID] = container
		}

		t.Run("instrumented", func(t *testing.T) {
			d := containers["arn:aws:ecs:us-east-2:012345678910:task/9781c248-0edd-4cdb-9a93-f63cb662a5d3::nginx-curl"]
			require.NotEmpty(t, d)

			assert.NotEmpty(t, d.EntityID)
			assert.Equal(t, "43481a6ce4842eec8fe72fc28500c6b52edcc0917f105b83379f88cac1ff3946", d.Data["Id"])
		})

		t.Run("non-instrumented", func(t *testing.T) {
			d := containers["arn:aws:ecs:us-east-2:012345678910:task/9781c248-0edd-4cdb-9a93-f63cb662a5d3::~internal~ecs~pause"]
			require.NotEmpty(t, d)

			assert.NotEmpty(t, d.EntityID)
			assert.Equal(t, "731a0d6a3b4210e2448339bc7015aaa79bfe4fa256384f4102db86ef94cbbc4c", d.Data["Id"])
		})
	})

	t.Run("Process plugin payload", func(t *testing.T) {
		require.Len(t, pluginData["com.instana.plugin.process"], 1)
		d := pluginData["com.instana.plugin.process"][0]

		assert.NotEmpty(t, d.EntityID)

		assert.Equal(t, "docker", d.Data["containerType"])
		assert.Equal(t, "43481a6ce4842eec8fe72fc28500c6b52edcc0917f105b83379f88cac1ff3946", d.Data["container"])
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

func TestFargateAgent_SendSpans(t *testing.T) {
	defer agent.Reset()

	sensor := instana.NewSensor("testing")

	sp := sensor.Tracer().StartSpan("entry")
	sp.SetTag("value", "42")
	sp.Finish()

	require.Eventually(t, func() bool {
		if len(agent.Bundles) == 0 {
			return false
		}

		for _, bundle := range agent.Bundles {
			var payload struct {
				Spans []json.RawMessage `json:"spans"`
			}

			json.Unmarshal(bundle.Body, &payload)
			if len(payload.Spans) > 0 {
				return true
			}
		}

		return false
	}, 4*time.Second, 500*time.Millisecond)

	var spans []map[string]json.RawMessage
	for _, bundle := range agent.Bundles {
		var payload struct {
			Spans []map[string]json.RawMessage `json:"spans"`
		}

		require.NoError(t, json.Unmarshal(bundle.Body, &payload), "%s", string(bundle.Body))
		spans = append(spans, payload.Spans...)
	}

	require.Len(t, spans, 1)
	assert.JSONEq(t, `{"hl": true, "cp": "aws", "e": "arn:aws:ecs:us-east-2:012345678910:task/9781c248-0edd-4cdb-9a93-f63cb662a5d3::nginx-curl"}`, string(spans[0]["f"]))
}

func setupAWSFargateEnv() func() {
	teardown := restoreEnvVarFunc("AWS_EXECUTION_ENV")
	os.Setenv("AWS_EXECUTION_ENV", "AWS_ECS_FARGATE")

	return teardown
}

func setupMetadataServer() func() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		http.ServeFile(w, req, "aws/testdata/container_metadata.json")
	})
	mux.HandleFunc("/task", func(w http.ResponseWriter, req *http.Request) {
		http.ServeFile(w, req, "aws/testdata/task_metadata.json")
	})
	mux.HandleFunc("/task/stats", func(w http.ResponseWriter, req *http.Request) {
		http.ServeFile(w, req, "aws/testdata/task_stats.json")
	})

	srv := httptest.NewServer(mux)

	teardown := restoreEnvVarFunc("ECS_CONTAINER_METADATA_URI")
	os.Setenv("ECS_CONTAINER_METADATA_URI", srv.URL)

	return func() {
		teardown()
		srv.Close()
	}
}
