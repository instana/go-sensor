// (c) Copyright IBM Corp. 2024

//go:build generic_serverless && integration
// +build generic_serverless,integration

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
	"github.com/stretchr/testify/require"
)

var agent *serverlessAgent

func TestMain(m *testing.M) {
	teardownInstanaEnv := setupInstanaEnv()
	defer teardownInstanaEnv()

	var err error
	agent, err = setupServerlessAgent()
	if err != nil {
		log.Fatalf("failed to initialize serverless agent: %s", err)
	}

	os.Exit(m.Run())
}

func TestIntegration_LocalServerlessAgent_SendSpans(t *testing.T) {
	defer agent.Reset()

	c := instana.InitCollector(instana.DefaultOptions())
	defer instana.ShutdownCollector()

	sp := c.Tracer().StartSpan("generic_serverless")
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
}

func TestIntegration_LocalServerlessAgent_SendSpans_Error(t *testing.T) {
	defer agent.Reset()

	c := instana.InitCollector(instana.DefaultOptions())
	defer instana.ShutdownCollector()

	sp := c.Tracer().StartSpan("http")
	sp.SetTag("returnError", "true")
	sp.Finish()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	require.NoError(t, c.Flush(ctx))
	require.Len(t, agent.Bundles, 0)
}

func TestIntegration_LocalServerlessAgent_SendSpans_Multiple(t *testing.T) {
	defer agent.Reset()

	c := instana.InitCollector(instana.DefaultOptions())
	defer instana.ShutdownCollector()

	// Create parent span
	parentSp := c.Tracer().StartSpan("entry")
	parentSp.SetTag("span.kind", "entry")

	// Create child spans
	childSp1 := c.Tracer().StartSpan("http.client", opentracing.ChildOf(parentSp.Context()))
	childSp1.SetTag("http.url", "https://api.example.com/data")
	childSp1.SetTag("http.method", "GET")
	childSp1.Finish()

	childSp2 := c.Tracer().StartSpan("database", opentracing.ChildOf(parentSp.Context()))
	childSp2.SetTag("db.type", "postgresql")
	childSp2.SetTag("db.statement", "SELECT * FROM users")
	childSp2.Finish()

	childSp3 := c.Tracer().StartSpan("sdk.custom", opentracing.ChildOf(parentSp.Context()))
	childSp3.SetTag("custom.tag", "value")
	childSp3.Finish()

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

	require.Len(t, spans, 4)
}

func TestIntegration_LocalServerlessAgent_Headers(t *testing.T) {
	defer agent.Reset()

	c := instana.InitCollector(instana.DefaultOptions())
	defer instana.ShutdownCollector()

	sp := c.Tracer().StartSpan("test-span")
	sp.Finish()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	require.NoError(t, c.Flush(ctx))
	require.Len(t, agent.Bundles, 1)

	collected := agent.Bundles[0]

	// Verify headers
	require.NotEmpty(t, collected.Header.Get("X-Instana-Host"))
	require.Equal(t, "testkey1", collected.Header.Get("X-Instana-Key"))
	require.Equal(t, "application/json", collected.Header.Get("Content-Type"))
}

func TestIntegration_LocalServerlessAgent_SpanFrom(t *testing.T) {
	defer agent.Reset()

	c := instana.InitCollector(instana.DefaultOptions())
	defer instana.ShutdownCollector()

	sp := c.Tracer().StartSpan("test-span")
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

	// Verify span has correct From field
	var fromData map[string]interface{}
	require.NoError(t, json.Unmarshal(spans[0]["f"], &fromData))

	require.NotEmpty(t, fromData["e"])                     // entity ID
	require.Equal(t, true, fromData["hl"])                 // hostless
	require.Equal(t, "generic_serverless", fromData["cp"]) // cloud provider
}

func TestIntegration_LocalServerlessAgent_PeriodicFlush(t *testing.T) {
	defer agent.Reset()

	c := instana.InitCollector(instana.DefaultOptions())
	defer instana.ShutdownCollector()

	sp := c.Tracer().StartSpan("test-span")
	sp.Finish()

	// Wait for periodic flush (flushPeriodForGenericInSec = 2 seconds)
	time.Sleep(2500 * time.Millisecond)

	// Should have received at least one bundle
	require.GreaterOrEqual(t, len(agent.Bundles), 1)
}

func TestIntegration_LocalServerlessAgent_ConcurrentSpans(t *testing.T) {
	defer agent.Reset()

	c := instana.InitCollector(instana.DefaultOptions())
	defer instana.ShutdownCollector()

	// Create multiple spans concurrently
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			sp := c.Tracer().StartSpan("concurrent-span")
			sp.SetTag("span.id", id)
			sp.Finish()
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

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

	require.Len(t, spans, 10)
}

func TestIntegration_LocalServerlessAgent_SpanWithTags(t *testing.T) {
	defer agent.Reset()

	c := instana.InitCollector(instana.DefaultOptions())
	defer instana.ShutdownCollector()

	sp := c.Tracer().StartSpan("http.server")
	sp.SetTag("http.url", "https://example.com/api/users")
	sp.SetTag("http.method", "POST")
	sp.SetTag("http.status", 200)
	sp.SetTag("http.path", "/api/users")
	sp.SetTag("http.host", "example.com")
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

	// Verify span has data
	require.NotEmpty(t, spans[0]["data"])
}

func TestIntegration_LocalServerlessAgent_SpanWithError(t *testing.T) {
	defer agent.Reset()

	c := instana.InitCollector(instana.DefaultOptions())
	defer instana.ShutdownCollector()

	sp := c.Tracer().StartSpan("error-span")
	sp.SetTag("error", true)
	sp.SetTag("error.message", "Something went wrong")
	sp.SetTag("error.type", "RuntimeError")
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

	// Verify span has error count
	var ec int
	require.NoError(t, json.Unmarshal(spans[0]["ec"], &ec))
	require.Equal(t, 1, ec)
}

func TestIntegration_LocalServerlessAgent_EmptyFlush(t *testing.T) {
	defer agent.Reset()

	c := instana.InitCollector(instana.DefaultOptions())
	defer instana.ShutdownCollector()

	// Flush without creating any spans
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	require.NoError(t, c.Flush(ctx))
	// Should not send any bundles when there are no spans
	require.Len(t, agent.Bundles, 0)
}

func TestIntegration_LocalServerlessAgent_MultipleFlushes(t *testing.T) {
	defer agent.Reset()

	c := instana.InitCollector(instana.DefaultOptions())
	defer instana.ShutdownCollector()

	// First flush
	sp1 := c.Tracer().StartSpan("span-1")
	sp1.Finish()

	ctx1, cancel1 := context.WithTimeout(context.Background(), time.Second)
	defer cancel1()

	require.NoError(t, c.Flush(ctx1))
	require.Len(t, agent.Bundles, 1)

	// Second flush
	sp2 := c.Tracer().StartSpan("span-2")
	sp2.Finish()

	ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second)
	defer cancel2()

	require.NoError(t, c.Flush(ctx2))
	require.Len(t, agent.Bundles, 2)

	// Verify each bundle has one span
	for _, bundle := range agent.Bundles {
		var payload struct {
			Spans []map[string]json.RawMessage `json:"spans"`
		}

		require.NoError(t, json.Unmarshal(bundle.Body, &payload), "%s", string(bundle.Body))
		require.Len(t, payload.Spans, 1)
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
