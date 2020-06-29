// +build fargate_integration

package instana_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
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

	var err error
	agent, err = setupServerlessAgent()
	if err != nil {
		log.Fatalf("failed to initialize serverless agent: %s", err)
	}

	instana.InitSensor(&instana.Options{})

	os.Exit(m.Run())
}

func TestFargateAgent_SendMetrics(t *testing.T) {
	defer agent.Reset()

	require.Eventually(t, func() bool { return len(agent.Metrics) > 0 }, 2*time.Second, 500*time.Millisecond)

	collected := agent.Metrics[0]

	assert.Equal(t, "arn:aws:ecs:us-east-2:012345678910:task/9781c248-0edd-4cdb-9a93-f63cb662a5d3::nginx-curl", collected.Header.Get("X-Instana-Host"))
	assert.Equal(t, "testkey1", collected.Header.Get("X-Instana-Key"))
	assert.NotEmpty(t, collected.Header.Get("X-Instana-Time"))

	var payload struct {
		Plugins []struct {
			Name     string                 `json:"name"`
			EntityID string                 `json:"entityId"`
			Data     map[string]interface{} `json:"data"`
		} `json:"plugins"`
	}
	require.NoError(t, json.Unmarshal(collected.Body, &payload))

	pluginData := make(map[string]serverlessAgentPluginPayload)
	for _, plugin := range payload.Plugins {
		pluginData[plugin.Name] = serverlessAgentPluginPayload{plugin.EntityID, plugin.Data}
	}

	// AWS ECS Task plugin payload
	if assert.Contains(t, pluginData, "com.instana.plugin.aws.ecs.task") {
		d := pluginData["com.instana.plugin.aws.ecs.task"]

		assert.NotEmpty(t, d.EntityID)
		assert.Equal(t, d.Data["taskArn"], d.EntityID)

		assert.Equal(t, "default", d.Data["clusterArn"])
		assert.Equal(t, "nginx", d.Data["taskDefinition"])
		assert.Equal(t, "5", d.Data["taskDefinitionVersion"])
	}

	// AWS ECS Container plugin payload
	if assert.Contains(t, pluginData, "com.instana.plugin.aws.ecs.container") {
		d := pluginData["com.instana.plugin.aws.ecs.container"]

		assert.NotEmpty(t, d.EntityID)

		require.IsType(t, d.Data["taskArn"], "")
		require.IsType(t, d.Data["containerName"], "")
		assert.Equal(t, d.Data["taskArn"].(string)+"::"+d.Data["containerName"].(string), d.EntityID)

		if assert.NotEmpty(t, d.Data["taskArn"]) {
			assert.Equal(t, pluginData["com.instana.plugin.aws.ecs.task"].EntityID, d.Data["taskArn"])
		}

		assert.Equal(t, true, d.Data["instrumented"])
		assert.Equal(t, "go", d.Data["runtime"])
		assert.Equal(t, "43481a6ce4842eec8fe72fc28500c6b52edcc0917f105b83379f88cac1ff3946", d.Data["dockerId"])
	}

	// Docker plugin payload
	if assert.Contains(t, pluginData, "com.instana.plugin.docker") {
		d := pluginData["com.instana.plugin.docker"]

		assert.NotEmpty(t, d.EntityID)
		assert.Equal(t, pluginData["com.instana.plugin.aws.ecs.container"].EntityID, d.EntityID)

		assert.Equal(t, "43481a6ce4842eec8fe72fc28500c6b52edcc0917f105b83379f88cac1ff3946", d.Data["Id"])
	}

	// Process plugin payload
	if assert.Contains(t, pluginData, "com.instana.plugin.process") {
		d := pluginData["com.instana.plugin.process"]

		assert.NotEmpty(t, d.EntityID)

		assert.Equal(t, "docker", d.Data["containerType"])
		assert.Equal(t, "43481a6ce4842eec8fe72fc28500c6b52edcc0917f105b83379f88cac1ff3946", d.Data["container"])
	}

	// Go process plugin payload
	if assert.Contains(t, pluginData, "com.instana.plugin.golang") {
		d := pluginData["com.instana.plugin.golang"]

		assert.NotEmpty(t, d.EntityID)
		assert.Equal(t, pluginData["com.instana.plugin.process"].EntityID, d.EntityID)

		assert.NotEmpty(t, d.Data["metrics"])
	}
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

	srv := httptest.NewServer(mux)

	teardown := restoreEnvVarFunc("ECS_CONTAINER_METADATA_URI")
	os.Setenv("ECS_CONTAINER_METADATA_URI", srv.URL)

	return func() {
		teardown()
		srv.Close()
	}
}

type serverlessAgentPluginPayload struct {
	EntityID string
	Data     map[string]interface{}
}

type serverlessAgentRequest struct {
	Header http.Header
	Body   []byte
}

type serverlessAgent struct {
	Metrics []serverlessAgentRequest

	ln           net.Listener
	restoreEnvFn func()
}

func setupServerlessAgent() (*serverlessAgent, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize the serverless agent listener: %s", err)
	}

	srv := &serverlessAgent{
		ln:           ln,
		restoreEnvFn: restoreEnvVarFunc("INSTANA_ENDPOINT_URL"),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", srv.HandleMetrics)

	go http.Serve(ln, mux)

	os.Setenv("INSTANA_ENDPOINT_URL", "http://"+ln.Addr().String())

	return srv, nil
}

func (srv *serverlessAgent) HandleMetrics(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Printf("ERROR: failed to read serverless agent metrics request body: %s", err)
		body = nil
	}

	srv.Metrics = append(srv.Metrics, serverlessAgentRequest{
		Header: req.Header,
		Body:   body,
	})

	w.WriteHeader(http.StatusNoContent)
}

func (srv *serverlessAgent) Reset() {
	srv.Metrics = nil
}

func (srv *serverlessAgent) Teardown() {
	srv.restoreEnvFn()
	srv.ln.Close()
}

func restoreEnvVarFunc(key string) func() {
	if oldValue, ok := os.LookupEnv(key); ok {
		return func() { os.Setenv(key, oldValue) }
	}

	return func() { os.Unsetenv(key) }
}
