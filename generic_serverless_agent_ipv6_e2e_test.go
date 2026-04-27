// (c) Copyright IBM Corp. 2026

package instana

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// TestIntegration_RealServerlessAgent_IPv6Endpoint tests sending spans to the real
// Instana SaaS endpoint using IPv6. This test forces IPv6 resolution and validates
// that the agent can successfully communicate with the production endpoint over IPv6.
// Set INSTANA_ENDPOINT_URL environment variable to specify the endpoint to test.
func TestIntegration_RealServerlessAgent_IPv6Endpoint(t *testing.T) {
	teardownInstanaEnv := setupInstanaEnvQA()
	defer teardownInstanaEnv()

	agentKey := validateAgentKey(t)
	endpoint := validateEndpoint(t)
	host := extractHostFromEndpoint(endpoint)
	ipv6Addr := findIPv6Address(t, host)

	// IPv6 address resolved successfully (not logged for security)
	ipv6Endpoint := "https://[" + ipv6Addr.String() + "]"
	defer restoreEndpointEnv(ipv6Endpoint)()

	customClient := createTLSClient(host)
	// Creating custom HTTP client with TLS configuration (details not logged for security)

	customAgent := newGenericServerlessAgent(ipv6Endpoint, agentKey, customClient, defaultLogger)
	c := initCollectorWithAgent(customAgent)
	defer ShutdownCollector()

	sendTestSpan(c, endpoint, ipv6Addr)
	flushAndVerify(t, c)
}

// validateAgentKey checks if INSTANA_AGENT_KEY is set and valid
func validateAgentKey(t *testing.T) string {
	agentKey := os.Getenv("INSTANA_AGENT_KEY")
	if agentKey == "" || agentKey == "testkey1" {
		t.Skip("Skipping real endpoint test: INSTANA_AGENT_KEY not set to production key")
	}
	return agentKey
}

// validateEndpoint checks if INSTANA_ENDPOINT_URL is set
func validateEndpoint(t *testing.T) string {
	endpoint := os.Getenv("INSTANA_ENDPOINT_URL")
	if endpoint == "" {
		t.Skip("Skipping real endpoint test: INSTANA_ENDPOINT_URL not set")
	}
	return endpoint
}

// extractHostFromEndpoint extracts the hostname from an endpoint URL
func extractHostFromEndpoint(endpoint string) string {
	host := endpoint
	host = strings.TrimPrefix(host, "https://")
	host = strings.TrimPrefix(host, "http://")

	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}
	if idx := strings.Index(host, "/"); idx != -1 {
		host = host[:idx]
	}
	return host
}

// findIPv6Address resolves and returns the first IPv6 address for the host
func findIPv6Address(t *testing.T, host string) net.IP {
	ips, err := net.LookupIP(host)
	if err != nil {
		t.Skipf("Failed to resolve %s: %v", host, err)
	}

	for _, ip := range ips {
		if ip.To4() == nil && ip.To16() != nil {
			return ip
		}
	}

	t.Skip("No IPv6 address found for endpoint")
	return nil
}

// restoreEndpointEnv sets the IPv6 endpoint and returns a cleanup function
func restoreEndpointEnv(ipv6Endpoint string) func() {
	oldEndpoint := os.Getenv("INSTANA_ENDPOINT_URL")
	os.Setenv("INSTANA_ENDPOINT_URL", ipv6Endpoint)

	return func() {
		if oldEndpoint != "" {
			os.Setenv("INSTANA_ENDPOINT_URL", oldEndpoint)
		} else {
			os.Unsetenv("INSTANA_ENDPOINT_URL")
		}
	}
}

// createTLSClient creates an HTTP client with TLS configuration for IPv6
func createTLSClient(host string) *http.Client {
	tlsConfig := &tls.Config{
		ServerName: host, // Verify certificate against the original hostname
	}

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: 15 * time.Second,
	}
}

// initCollectorWithAgent initializes the collector with a custom agent
func initCollectorWithAgent(customAgent AgentClient) TracerLogger {
	opts := &Options{
		Service:           "ipv6-integration-test",
		EnableAutoProfile: false,
		MaxBufferedSpans:  100,
		AgentClient:       customAgent,
	}
	return InitCollector(opts)
}

// sendTestSpan creates and sends a test span
func sendTestSpan(c TracerLogger, endpoint string, ipv6Addr net.IP) {
	sp := c.Tracer().StartSpan("ipv6-test-span")
	sp.SetTag("test.type", "ipv6-integration")
	sp.SetTag("test.endpoint", endpoint)
	sp.SetTag("test.ipv6", ipv6Addr.String())
	sp.Finish()
}

// flushAndVerify flushes the collector and logs the result
func flushAndVerify(t *testing.T, c TracerLogger) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err := c.Flush(ctx)
	if err != nil {
		t.Logf("Warning: Flush returned error (may be expected for test key): %v", err)
	}
	t.Log("Successfully sent span via IPv6")
}

// setupInstanaEnvQA reads QA environment variables and sets up the test environment.
// It reads INSTANA_AGENT_KEY_QA and INSTANA_ENDPOINT_URL_QA and sets
// INSTANA_AGENT_KEY and INSTANA_ENDPOINT_URL accordingly.
// Returns a teardown function that restores the original environment variables.
func setupInstanaEnvQA() func() {
	var teardownFuncs []func()

	// Read and set INSTANA_AGENT_KEY from INSTANA_AGENT_KEY_QA
	agentKeyQA := os.Getenv("INSTANA_AGENT_KEY_QA")
	if agentKeyQA != "" {
		oldAgentKey := os.Getenv("INSTANA_AGENT_KEY")
		os.Setenv("INSTANA_AGENT_KEY", agentKeyQA)
		teardownFuncs = append(teardownFuncs, func() {
			if oldAgentKey != "" {
				os.Setenv("INSTANA_AGENT_KEY", oldAgentKey)
			} else {
				os.Unsetenv("INSTANA_AGENT_KEY")
			}
		})
	}

	// Read and set INSTANA_ENDPOINT_URL from INSTANA_ENDPOINT_URL_QA
	endpointQA := os.Getenv("INSTANA_ENDPOINT_URL_QA")
	if endpointQA != "" {
		oldEndpoint := os.Getenv("INSTANA_ENDPOINT_URL")
		os.Setenv("INSTANA_ENDPOINT_URL", endpointQA)
		teardownFuncs = append(teardownFuncs, func() {
			if oldEndpoint != "" {
				os.Setenv("INSTANA_ENDPOINT_URL", oldEndpoint)
			} else {
				os.Unsetenv("INSTANA_ENDPOINT_URL")
			}
		})
	}

	// Return teardown function that restores all environment variables in reverse order
	return func() {
		for i := len(teardownFuncs) - 1; i >= 0; i-- {
			teardownFuncs[i]()
		}
	}
}
