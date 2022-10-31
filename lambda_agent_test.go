// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

//go:build lambda && integration
// +build lambda,integration

package instana_test

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"testing"
	"time"

	instana "github.com/instana/go-sensor"
	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var agent *serverlessAgent

func TestMain(m *testing.M) {
	teardownEnv := setupAWSLambdaEnv()
	defer teardownEnv()

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

func TestLambdaAgent_SendSpans(t *testing.T) {
	defer agent.Reset()

	tracer := instana.NewTracer()
	sensor := instana.NewSensorWithTracer(tracer)

	sp := sensor.Tracer().StartSpan("aws.lambda.entry", opentracing.Tags{
		"lambda.arn":     "aws::test-lambda::$LATEST",
		"lambda.name":    "test-lambda",
		"lambda.version": "$LATEST",
	})
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
	assert.JSONEq(t, `{"hl": true, "cp": "aws", "e": "aws::test-lambda::$LATEST"}`, string(spans[0]["f"]))
}

func setupAWSLambdaEnv() func() {
	teardown := restoreEnvVarFunc("AWS_EXECUTION_ENV")
	os.Setenv("AWS_EXECUTION_ENV", "AWS_Lambda_go1.x")

	return teardown
}
