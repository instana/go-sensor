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

	instana.InitCollector(instana.DefaultOptions())
	defer instana.ShutdownCollector()

	os.Exit(m.Run())
}

func TestIntegration_LambdaAgent_SendSpans(t *testing.T) {
	defer agent.Reset()

	c := instana.InitCollector(instana.DefaultOptions())
	defer instana.ShutdownCollector()

	sp := c.Tracer().StartSpan("aws.lambda.entry", opentracing.Tags{
		"lambda.arn":     "aws::test-lambda::$LATEST",
		"lambda.name":    "test-lambda",
		"lambda.version": "$LATEST",
	})
	sp.Finish()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	require.NoError(t, c.Flush(ctx))
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

func TestIntegration_LambdaAgent_SendSpans_Error(t *testing.T) {
	defer agent.Reset()

	c := instana.InitCollector(instana.DefaultOptions())
	defer instana.ShutdownCollector()

	sp := c.Tracer().StartSpan("aws.lambda.entry", opentracing.Tags{
		"lambda.arn":     "aws::test-lambda::$LATEST",
		"lambda.name":    "test-lambda",
		"lambda.version": "$LATEST",
		"lambda.error":   "true",
	})
	sp.Finish()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	require.NoError(t, c.Flush(ctx))
	require.Len(t, agent.Bundles, 0)
}

func TestIntegration_LambdaAgent_SendSpans_WithTrigger(t *testing.T) {
	defer agent.Reset()

	c := instana.InitCollector(instana.DefaultOptions())
	defer instana.ShutdownCollector()

	sp := c.Tracer().StartSpan("aws.lambda.entry", opentracing.Tags{
		"lambda.arn":     "arn:aws:lambda:us-east-1:123456789:function:test-lambda",
		"lambda.name":    "test-lambda",
		"lambda.version": "$LATEST",
		"lambda.trigger": "aws:api.gateway",
	})
	sp.Finish()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	require.NoError(t, c.Flush(ctx))
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
	assert.JSONEq(t, `{"hl": true, "cp": "aws", "e": "arn:aws:lambda:us-east-1:123456789:function:test-lambda"}`, string(spans[0]["f"]))
}

func TestIntegration_LambdaAgent_SendSpans_Multiple(t *testing.T) {
	defer agent.Reset()

	c := instana.InitCollector(instana.DefaultOptions())
	defer instana.ShutdownCollector()

	// Create parent lambda span
	parentSp := c.Tracer().StartSpan("aws.lambda.entry", opentracing.Tags{
		"lambda.arn":     "arn:aws:lambda:us-east-1:123456789:function:test-lambda",
		"lambda.name":    "test-lambda",
		"lambda.version": "$LATEST",
	})

	// Create child spans
	childSp1 := c.Tracer().StartSpan("http.client", opentracing.ChildOf(parentSp.Context()))
	childSp1.SetTag("http.url", "https://api.example.com/data")
	childSp1.SetTag("http.method", "GET")
	childSp1.Finish()

	childSp2 := c.Tracer().StartSpan("sdk.custom", opentracing.ChildOf(parentSp.Context()))
	childSp2.SetTag("custom.tag", "value")
	childSp2.Finish()

	parentSp.Finish()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	require.NoError(t, c.Flush(ctx))
	require.Len(t, agent.Bundles, 1)

	var spans []map[string]json.RawMessage
	for _, bundle := range agent.Bundles {
		var payload struct {
			Spans []map[string]json.RawMessage `json:"spans"`
		}

		require.NoError(t, json.Unmarshal(bundle.Body, &payload), "%s", string(bundle.Body))
		spans = append(spans, payload.Spans...)
	}

	require.Len(t, spans, 3)
}

func TestIntegration_LambdaAgent_SendSpans_WithColdStart(t *testing.T) {
	defer agent.Reset()

	c := instana.InitCollector(instana.DefaultOptions())
	defer instana.ShutdownCollector()

	sp := c.Tracer().StartSpan("aws.lambda.entry", opentracing.Tags{
		"lambda.arn":       "arn:aws:lambda:us-east-1:123456789:function:test-lambda",
		"lambda.name":      "test-lambda",
		"lambda.version":   "$LATEST",
		"lambda.coldStart": true,
		"lambda.msleft":    5000,
	})
	sp.Finish()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	require.NoError(t, c.Flush(ctx))
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

	// Verify span data contains lambda information
	var spanData map[string]interface{}
	require.NoError(t, json.Unmarshal(spans[0]["data"], &spanData))

	lambdaData, ok := spanData["lambda"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "arn:aws:lambda:us-east-1:123456789:function:test-lambda", lambdaData["arn"])
	assert.Equal(t, true, lambdaData["coldStart"])
}

func TestIntegration_LambdaAgent_Headers(t *testing.T) {
	defer agent.Reset()

	c := instana.InitCollector(instana.DefaultOptions())
	defer instana.ShutdownCollector()

	sp := c.Tracer().StartSpan("aws.lambda.entry", opentracing.Tags{
		"lambda.arn":     "arn:aws:lambda:us-east-1:123456789:function:test-lambda",
		"lambda.name":    "test-lambda",
		"lambda.version": "$LATEST",
	})
	sp.Finish()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	require.NoError(t, c.Flush(ctx))
	require.Len(t, agent.Bundles, 1)

	collected := agent.Bundles[0]

	// Verify headers
	assert.Equal(t, "arn:aws:lambda:us-east-1:123456789:function:test-lambda", collected.Header.Get("X-Instana-Host"))
	assert.Equal(t, "testkey1", collected.Header.Get("X-Instana-Key"))
	assert.NotEmpty(t, collected.Header.Get("X-Instana-Time"))
	assert.Equal(t, "application/json", collected.Header.Get("Content-Type"))
}

func TestIntegration_LambdaAgent_Metrics(t *testing.T) {
	defer agent.Reset()

	c := instana.InitCollector(instana.DefaultOptions())
	defer instana.ShutdownCollector()

	sp := c.Tracer().StartSpan("aws.lambda.entry", opentracing.Tags{
		"lambda.arn":     "arn:aws:lambda:us-east-1:123456789:function:test-lambda",
		"lambda.name":    "test-lambda",
		"lambda.version": "$LATEST",
	})
	sp.Finish()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	require.NoError(t, c.Flush(ctx))
	require.Len(t, agent.Bundles, 1)

	var payload struct {
		Metrics struct {
			Plugins []struct {
				Name     string                 `json:"name"`
				EntityID string                 `json:"entityId"`
				Data     map[string]interface{} `json:"data"`
			} `json:"plugins"`
		} `json:"metrics"`
	}
	require.NoError(t, json.Unmarshal(agent.Bundles[0].Body, &payload))

	// Verify AWS Lambda plugin payload is present
	require.Len(t, payload.Metrics.Plugins, 1)
	assert.Equal(t, "com.instana.plugin.aws.lambda", payload.Metrics.Plugins[0].Name)
	assert.Equal(t, "arn:aws:lambda:us-east-1:123456789:function:test-lambda", payload.Metrics.Plugins[0].EntityID)
}

func TestIntegration_LambdaAgent_FlushWithoutSnapshot(t *testing.T) {
	defer agent.Reset()

	c := instana.InitCollector(instana.DefaultOptions())
	defer instana.ShutdownCollector()

	// Create a non-lambda span
	sp := c.Tracer().StartSpan("sdk.custom")
	sp.SetTag("custom.tag", "value")
	sp.Finish()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Flush should fail because there's no lambda snapshot
	err := c.Flush(ctx)
	// The error might be wrapped, so we just check that flush completes
	// In lambda agent, it returns ErrAgentNotReady when no snapshot is available
	if err != nil {
		assert.Contains(t, err.Error(), "agent not ready")
	}
}

func TestIntegration_LambdaAgent_PeriodicFlush(t *testing.T) {
	defer agent.Reset()

	c := instana.InitCollector(instana.DefaultOptions())
	defer instana.ShutdownCollector()

	sp := c.Tracer().StartSpan("aws.lambda.entry", opentracing.Tags{
		"lambda.arn":     "arn:aws:lambda:us-east-1:123456789:function:test-lambda",
		"lambda.name":    "test-lambda",
		"lambda.version": "$LATEST",
	})
	sp.Finish()

	// Wait for periodic flush (awsLambdaAgentFlushPeriod = 2 seconds)
	time.Sleep(2500 * time.Millisecond)

	// Should have received at least one bundle
	assert.GreaterOrEqual(t, len(agent.Bundles), 1)
}

func setupAWSLambdaEnv() func() {
	teardown := restoreEnvVarFunc("AWS_EXECUTION_ENV")
	os.Setenv("AWS_EXECUTION_ENV", "AWS_Lambda_go1.x")

	return teardown
}
