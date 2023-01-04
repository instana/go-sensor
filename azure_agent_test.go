// (c) Copyright IBM Corp. 2022

//go:build azure && integration
// +build azure,integration

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
	teardownAzureEnv := setupAzureFunctionEnv()
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

func TestAzureAgent_SendSpans(t *testing.T) {
	defer agent.Reset()

	tracer := instana.NewTracer()
	sensor := instana.NewSensorWithTracer(tracer)
	defer instana.ShutdownSensor()

	sp := sensor.Tracer().StartSpan("azf")
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
	assert.JSONEq(t, `{"hl": true, "cp": "azure", "e": "/subscriptions/testgh05-3f0d-4bf9-8f53-209408003632/resourceGroups/test-resourcegroup/providers/Microsoft.Web/sites/test-funcname"}`, string(spans[0]["f"]))
}

func TestAzureAgent_SpanDetails(t *testing.T) {
	defer agent.Reset()

	tracer := instana.NewTracer()
	sensor := instana.NewSensorWithTracer(tracer)
	defer instana.ShutdownSensor()

	sp := sensor.Tracer().StartSpan("azf")
	sp.SetTag("azf.triggername", "HTTP")
	sp.SetTag("azf.functionname", "testfunction")
	sp.SetTag("azf.name", "testapp")
	sp.SetTag("azf.methodname", "testmethod")
	sp.SetTag("azf.runtime", "custom")
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
	assert.JSONEq(t, `{"hl": true, "cp": "azure", "e": "/subscriptions/testgh05-3f0d-4bf9-8f53-209408003632/resourceGroups/test-resourcegroup/providers/Microsoft.Web/sites/test-funcname"}`, string(spans[0]["f"]))
	assert.JSONEq(t, ` {
        "azf": {
          "name": "testapp",
          "methodname" : "testmethod",
          "functionname": "testfunction",
          "triggername": "HTTP",
          "runtime": "custom"
        } }`, string(spans[0]["data"]))
}

func setupAzureFunctionEnv() func() {
	var teardownFuncs []func()

	teardownFuncs = append(teardownFuncs, restoreEnvVarFunc("FUNCTIONS_WORKER_RUNTIME"))
	os.Setenv("FUNCTIONS_WORKER_RUNTIME", "custom")

	teardownFuncs = append(teardownFuncs, restoreEnvVarFunc("WEBSITE_OWNER_NAME"))
	os.Setenv("WEBSITE_OWNER_NAME", "testgh05-3f0d-4bf9-8f53-209408003632+testresourcegroup-GermanyWestCentralwebspace")

	teardownFuncs = append(teardownFuncs, restoreEnvVarFunc("WEBSITE_RESOURCE_GROUP"))
	os.Setenv("WEBSITE_RESOURCE_GROUP", "test-resourcegroup")

	teardownFuncs = append(teardownFuncs, restoreEnvVarFunc("APPSETTING_WEBSITE_SITE_NAME"))
	os.Setenv("APPSETTING_WEBSITE_SITE_NAME", "test-funcname")

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
