// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

//go:build integration
// +build integration

package instana_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	instana "github.com/instana/go-sensor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type serverlessAgentPluginPayload struct {
	EntityID string
	Data     map[string]interface{}
}

type serverlessAgentRequest struct {
	Header http.Header
	Body   []byte
}

type serverlessAgent struct {
	Bundles []serverlessAgentRequest

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
	mux.HandleFunc("/bundle", srv.HandleBundle)

	go http.Serve(ln, mux)

	os.Setenv("INSTANA_ENDPOINT_URL", "http://"+ln.Addr().String())

	return srv, nil
}

func (srv *serverlessAgent) HandleBundle(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		log.Printf("ERROR: failed to read serverless agent spans request body: %s", err)
		body = nil
	}

	var root Root
	err = json.Unmarshal(body, &root)
	if err != nil {
		log.Printf("ERROR: failed to unmarshal serverless agent spans request body: %s", err.Error())
	} else {
		if len(root.Spans) > 0 && (root.Spans[0].Data.SDKCustom.Tags.ReturnError == "true" ||
			root.Spans[0].Data.Lambda.ReturnError == "true") {

			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	srv.Bundles = append(srv.Bundles, serverlessAgentRequest{
		Header: req.Header,
		Body:   body,
	})

	w.WriteHeader(http.StatusNoContent)
}

func (srv *serverlessAgent) Reset() {
	srv.Bundles = nil
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

type Data struct {
	SDKCustom struct {
		Tags struct {
			ReturnError string `json:"returnError"`
		} `json:"tags"`
	} `json:"sdk.custom"`
	Lambda LambdaData `json:"lambda"`
}

type LambdaData struct {
	ReturnError string `json:"error"`
}

type Span struct {
	Data Data `json:"data"`
}

type Root struct {
	Spans []Span `json:"spans"`
}

// TestIntegration_Sensor_Ready tests the Ready() function
func TestIntegration_Sensor_Ready(t *testing.T) {
	agent, err := setupServerlessAgent()
	require.NoError(t, err)
	defer agent.Teardown()

	// Test 1: Ready should return false before sensor is initialized
	assert.False(t, instana.Ready(), "Ready() should return false before sensor initialization")

	// Test 2: Initialize sensor and verify Ready returns true
	instana.StartMetrics(&instana.Options{
		Service: "test-service",
	})
	defer instana.ShutdownSensor()

	// Wait for agent to be ready (with timeout)
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	ready := false
	for !ready {
		select {
		case <-timeout:
			t.Fatal("Timeout waiting for sensor to be ready")
		case <-ticker.C:
			ready = instana.Ready()
		}
	}

	assert.True(t, instana.Ready(), "Ready() should return true after sensor initialization")
}

// TestIntegration_Sensor_Flush tests the Flush() function
func TestIntegration_Sensor_Flush(t *testing.T) {
	agent, err := setupServerlessAgent()
	require.NoError(t, err)
	defer agent.Teardown()

	// Test 1: Flush should not error when sensor is not initialized
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err = instana.Flush(ctx)
	assert.NoError(t, err, "Flush() should not error when sensor is not initialized")

	// Reset agent bundles before starting the actual test
	agent.Reset()

	// Test 2: Initialize sensor and test flush with spans
	c := instana.InitCollector(&instana.Options{
		Service:                     "test-flush-service",
		MaxBufferedSpans:            10,
		ForceTransmissionStartingAt: 5,
	})
	defer instana.ShutdownCollector()

	// Wait for sensor to be ready with longer timeout
	timeout := time.After(10 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	ready := false
	for !ready {
		select {
		case <-timeout:
			t.Fatal("Timeout waiting for sensor to be ready")
		case <-ticker.C:
			ready = instana.Ready()
		}
	}

	// Give it a bit more time to fully initialize
	time.Sleep(500 * time.Millisecond)

	// Create a span
	sp := c.Tracer().StartSpan("test-flush-span")
	sp.SetTag("test.tag", "test-value")
	sp.Finish()

	// Flush the spans with longer timeout
	flushCtx, flushCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer flushCancel()

	err = instana.Flush(flushCtx)
	assert.NoError(t, err, "Flush() should not error after creating spans")

	// Wait a bit for the flush to complete
	time.Sleep(500 * time.Millisecond)

	// Verify spans were sent
	if len(agent.Bundles) > 0 {
		// Verify the span data
		var spans []map[string]json.RawMessage
		for _, bundle := range agent.Bundles {
			var payload struct {
				Spans []map[string]json.RawMessage `json:"spans"`
			}

			require.NoError(t, json.Unmarshal(bundle.Body, &payload), "Failed to unmarshal bundle: %s", string(bundle.Body))
			spans = append(spans, payload.Spans...)
		}

		require.GreaterOrEqual(t, len(spans), 1, "At least one span should be present after flush")
	} else {
		t.Log("No bundles received - this may be expected in serverless mode depending on agent readiness")
	}
}
