// (c) Copyright IBM Corp. 2024

//go:build azureContainerApps && integration
// +build azureContainerApps,integration

package instana_test

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"testing"
	"time"

	instana "github.com/instana/go-sensor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var agent *serverlessAgent

func TestMain(m *testing.M) {
	teardownAzureEnv := setupAzureContainerAppsEnv()
	defer teardownAzureEnv()

	teardownInstanaEnv := setupInstanaEnv()
	defer teardownInstanaEnv()

	var err error
	agent, err = setupServerlessAgent()
	if err != nil {
		log.Fatalf("failed to initialize serverless agent: %s", err)
	}

	os.Exit(m.Run())
}

func TestAzureContainerApps_SendSpans(t *testing.T) {
	defer agent.Reset()

	tracer := instana.NewTracer()
	sensor := instana.NewSensorWithTracer(tracer)
	defer instana.ShutdownSensor()

	sp := sensor.Tracer().StartSpan("aca")
	sp.Finish()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	require.NoError(t, tracer.Flush(ctx))
	require.Len(t, agent.Bundles, 1)

	var spans []map[string]json.RawMessage
	for _, bundle := range agent.Bundles {
		var payload struct {
			Spans []map[string]json.RawMessage `json:"spans"`
		}

		require.NoError(t, json.Unmarshal(bundle.Body, &payload), "%s", string(bundle.Body))
		spans = append(spans, payload.Spans...)
	}

	require.Len(t, spans, 1)
	assert.JSONEq(t, `{"hl": true, "cp": "azure", "e": "/subscriptions/testgh05-3f0d-4bf9-8f53-209408003632/resourceGroups/testresourcegroup/providers/Microsoft.App/containerapps/azureapp"}`, string(spans[0]["f"]))
}

func TestAzureAgent_SendSpans_Error(t *testing.T) {
	defer agent.Reset()

	tracer := instana.NewTracer()
	sensor := instana.NewSensorWithTracer(tracer)
	defer instana.ShutdownSensor()

	sp := sensor.Tracer().StartSpan("azf")
	sp.SetTag("returnError", "true")
	sp.Finish()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	require.NoError(t, tracer.Flush(ctx))
	require.Len(t, agent.Bundles, 0)
}

func setupAzureContainerAppsEnv() func() {
	var teardownFuncs []func()

	teardownFuncs = append(teardownFuncs, restoreEnvVarFunc("AZURE_SUBSCRIPTION_ID"))
	os.Setenv("AZURE_SUBSCRIPTION_ID", "testgh05-3f0d-4bf9-8f53-209408003632")

	teardownFuncs = append(teardownFuncs, restoreEnvVarFunc("AZURE_RESOURCE_GROUP"))
	os.Setenv("AZURE_RESOURCE_GROUP", "testresourcegroup")

	teardownFuncs = append(teardownFuncs, restoreEnvVarFunc("CONTAINER_APP_NAME"))
	os.Setenv("CONTAINER_APP_NAME", "azureapp")

	teardownFuncs = append(teardownFuncs, restoreEnvVarFunc("CONTAINER_APP_HOSTNAME"))
	os.Setenv("CONTAINER_APP_HOSTNAME", "azure_app_host")

	return func() {
		for _, f := range teardownFuncs {
			f()
		}
	}
}

func setupInstanaEnv() func() {
	var teardownFuncs []func()

	teardownFuncs = append(teardownFuncs, restoreEnvVarFunc("INSTANA_AGENT_KEY"))
	os.Setenv("INSTANA_AGENT_KEY", "testkey1")

	teardownFuncs = append(teardownFuncs, restoreEnvVarFunc("INSTANA_ZONE"))
	os.Setenv("INSTANA_ZONE", "testzone")

	teardownFuncs = append(teardownFuncs, restoreEnvVarFunc("INSTANA_TAGS"))
	os.Setenv("INSTANA_TAGS", "key1=value1,key2")

	teardownFuncs = append(teardownFuncs, restoreEnvVarFunc("INSTANA_SECRETS"))
	os.Setenv("INSTANA_SECRETS", "contains-ignore-case:key,password,secret,classified")

	teardownFuncs = append(teardownFuncs, restoreEnvVarFunc("CLASSIFIED_DATA"))
	os.Setenv("CLASSIFIED_DATA", "classified")

	return func() {
		for _, f := range teardownFuncs {
			f()
		}
	}
}
