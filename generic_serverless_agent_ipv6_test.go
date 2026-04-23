// SPDX-FileCopyrightText: 2024 IBM Corp.
//
// SPDX-License-Identifier: MIT

package instana

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestGenericServerlessAgent_Flush_IPv6Support verifies that the generic serverless agent
// can successfully flush spans to an IPv6 endpoint. This test creates an IPv6 test server
// listening on [::1]:0 and validates that the agent's Flush method can communicate with
// IPv6 addresses, ensuring proper header transmission and response handling.
func TestGenericServerlessAgent_Flush_IPv6Support(t *testing.T) {
	// Create a test server that listens on IPv6
	listener, err := net.Listen("tcp6", "[::1]:0") // IPv6 localhost
	if err != nil {
		t.Skipf("IPv6 not available on this system: %v", err)
	}
	defer listener.Close()

	requestReceived := make(chan bool, 1)
	var receivedHeaders http.Header

	// Create HTTP handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bundle" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// Verify it's a valid request
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		receivedHeaders = r.Header.Clone()
		requestReceived <- true
		w.WriteHeader(http.StatusOK)
	})

	// Start server with IPv6 listener
	server := &http.Server{Handler: handler}
	go server.Serve(listener)
	defer server.Close()

	// Get the IPv6 address with port
	addr := listener.Addr().(*net.TCPAddr)
	ipv6Endpoint := fmt.Sprintf("http://[%s]:%d", addr.IP, addr.Port)

	t.Logf("Test server listening on IPv6: %s", ipv6Endpoint)

	// Create agent with IPv6 endpoint
	agent := newGenericServerlessAgent(
		ipv6Endpoint,
		"test-key",
		&http.Client{Timeout: 5 * time.Second},
		defaultLogger,
	)

	// Send test spans (using the internal span structure)
	testSpan := Span{
		TraceID: 12345,
		SpanID:  67890,
		Name:    "test-operation",
	}

	err = agent.SendSpans([]Span{testSpan})
	if err != nil {
		t.Fatalf("failed to send spans: %v", err)
	}

	// Flush and verify
	err = agent.Flush(context.Background())
	if err != nil {
		t.Fatalf("flush failed with IPv6 endpoint: %v", err)
	}

	// Wait for request to be received
	select {
	case <-requestReceived:
		t.Logf("Successfully sent HTTP request to IPv6 endpoint")

		// Verify headers were set correctly
		if receivedHeaders.Get("X-Instana-Key") != "test-key" {
			t.Errorf("X-Instana-Key header not set correctly")
		}
		if receivedHeaders.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type header not set correctly")
		}

		t.Logf("IPv6 endpoint support verified: Flush() successfully communicates with IPv6 addresses")
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for request - IPv6 communication failed")
	}
}

// validateIPv4Endpoint checks if the endpoint has the correct IPv4 format
func validateIPv4Endpoint(t *testing.T, endpoint string) {
	if strings.Contains(endpoint, "[") || strings.Contains(endpoint, "]") {
		t.Errorf("IPv4 endpoint should not contain brackets: %s", endpoint)
	}
	if !strings.Contains(endpoint, "127.0.0.1") {
		t.Errorf("IPv4 endpoint should contain 127.0.0.1: %s", endpoint)
	}
}

// validateIPv6Endpoint checks if the endpoint has the correct IPv6 format
func validateIPv6Endpoint(t *testing.T, endpoint string) {
	if !strings.Contains(endpoint, "[") {
		t.Errorf("IPv6 endpoint should contain opening bracket: %s", endpoint)
	}
	if !strings.Contains(endpoint, "]") {
		t.Errorf("IPv6 endpoint should contain closing bracket: %s", endpoint)
	}
	if !strings.Contains(endpoint, "::1") {
		t.Errorf("IPv6 endpoint should contain ::1: %s", endpoint)
	}
}

// setupIPv4Server creates an IPv4 test server
func setupIPv4Server() (string, func()) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	return server.URL, server.Close
}

// setupIPv6Server creates an IPv6 test server
func setupIPv6Server(t *testing.T) (string, func()) {
	listener, err := net.Listen("tcp6", "[::1]:0")
	if err != nil {
		t.Skipf("IPv6 not available: %v", err)
	}

	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	}
	go server.Serve(listener)

	addr := listener.Addr().(*net.TCPAddr)
	endpoint := fmt.Sprintf("http://[%s]:%d", addr.IP, addr.Port)

	return endpoint, func() {
		server.Close()
		listener.Close()
	}
}

// testEndpointFlush tests flushing spans to an endpoint
func testEndpointFlush(t *testing.T, endpoint string) {
	agent := newGenericServerlessAgent(
		endpoint,
		"test-key",
		&http.Client{Timeout: 5 * time.Second},
		defaultLogger,
	)

	testSpan := Span{
		TraceID: 12345,
		SpanID:  67890,
		Name:    "test-operation",
	}

	err := agent.SendSpans([]Span{testSpan})
	if err != nil {
		t.Fatalf("failed to send spans: %v", err)
	}

	err = agent.Flush(context.Background())
	if err != nil {
		t.Fatalf("flush failed: %v", err)
	}
}

// TestGenericServerlessAgent_Flush_IPv4vsIPv6 is a comparative test that verifies the
// generic serverless agent works correctly with both IPv4 and IPv6 endpoints. This test
// ensures that the agent can flush spans to both address types without any issues,
// validating endpoint format (IPv4: http://127.0.0.1:port, IPv6: http://[::1]:port)
// and cross-compatibility for serverless environments.
func TestGenericServerlessAgent_Flush_IPv4vsIPv6(t *testing.T) {
	tests := []struct {
		name             string
		setupServer      func(*testing.T) (string, func())
		validateEndpoint func(*testing.T, string)
	}{
		{
			name:             "IPv4",
			setupServer:      func(t *testing.T) (string, func()) { return setupIPv4Server() },
			validateEndpoint: validateIPv4Endpoint,
		},
		{
			name:             "IPv6",
			setupServer:      setupIPv6Server,
			validateEndpoint: validateIPv6Endpoint,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoint, cleanup := tt.setupServer(t)
			defer cleanup()

			tt.validateEndpoint(t, endpoint)
			testEndpointFlush(t, endpoint)

			t.Logf("%s endpoint works correctly", tt.name)
		})
	}
}
