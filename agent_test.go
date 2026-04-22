// (c) Copyright IBM Corp. 2022

package instana

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"

	"github.com/stretchr/testify/assert"
)

func Test_agentS_SendSpans(t *testing.T) {
	tests := []struct {
		name  string
		spans []Span
	}{
		{
			name: "big span",
			spans: []Span{
				{
					Data: HTTPSpanData{
						Tags: HTTPSpanTags{
							URL: strings.Repeat("1", maxContentLength),
						},
					},
				},
			},
		},
		{
			name: "multiple big span",
			spans: []Span{
				{Data: HTTPSpanData{Tags: HTTPSpanTags{URL: strings.Repeat("1", maxContentLength)}}},
				{Data: HTTPSpanData{Tags: HTTPSpanTags{URL: strings.Repeat("1", maxContentLength)}}},
				{Data: HTTPSpanData{Tags: HTTPSpanTags{URL: strings.Repeat("1", maxContentLength)}}},
				{Data: HTTPSpanData{Tags: HTTPSpanTags{URL: strings.Repeat("1", maxContentLength)}}},
				{Data: HTTPSpanData{Tags: HTTPSpanTags{URL: strings.Repeat("1", maxContentLength)}}},
				{Data: HTTPSpanData{Tags: HTTPSpanTags{URL: strings.Repeat("1", maxContentLength)}}},
			},
		},
		{
			name: "not really a big span",
			spans: []Span{
				{
					Data: HTTPSpanData{
						Tags: HTTPSpanTags{
							URL: strings.Repeat("1", maxContentLength/2),
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ad := &agentCommunicator{
				host: "", from: &fromS{},
				client: &httpClientMock{
					resp: &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(bytes.NewReader([]byte("{}"))),
					},
				},
			}
			agent := &agentS{agentComm: ad, logger: defaultLogger}
			err := agent.SendSpans(tt.spans)

			assert.NoError(t, err)
		})
	}
}

type httpClientMock struct {
	resp   *http.Response
	err    error
	doFunc func(req *http.Request) (*http.Response, error)
}

func (h httpClientMock) Do(req *http.Request) (*http.Response, error) {
	if h.doFunc != nil {
		return h.doFunc(req)
	}
	return h.resp, h.err
}

func Test_agentResponse_getExtraHTTPHeaders(t *testing.T) {

	tests := []struct {
		name         string
		originalJSON string
		want         []string
	}{
		{
			name:         "old config",
			originalJSON: `{"pid":37808,"agentUuid":"88:66:5a:ff:fe:05:a5:f0","extraHeaders":["expected-value"],"secrets":{"matcher":"contains-ignore-case","list":["key","pass","secret"]}}`,
			want:         []string{"expected-value"},
		},
		{
			name:         "new config",
			originalJSON: `{"pid":38381,"agentUuid":"88:66:5a:ff:fe:05:a5:f0","tracing":{"extra-http-headers":["expected-value"]},"extraHeaders":["non-expected-value"],"secrets":{"matcher":"contains-ignore-case","list":["key","pass","secret"]}}`,
			want:         []string{"expected-value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &agentResponse{}
			json.Unmarshal([]byte(tt.originalJSON), r)
			assert.Equalf(t, tt.want, r.getExtraHTTPHeaders(), "getExtraHTTPHeaders()")
		})
	}
}

func Test_agentApplyHostSettings(t *testing.T) {
	fsm := &fsmS{
		agentComm: &agentCommunicator{
			host: "",
			from: &fromS{},
		},
		logger: defaultLogger,
	}

	response := agentResponse{
		Pid:    37892,
		HostID: "myhost",
		Tracing: struct {
			ExtraHTTPHeaders []string          `json:"extra-http-headers"`
			Disable          []map[string]bool `json:"disable"`
		}{
			ExtraHTTPHeaders: []string{"my-unwanted-custom-headers"},
		},
	}

	opts := &Options{
		Service: "test_service",
		Tracer: TracerOptions{
			CollectableHTTPHeaders: []string{"x-custom-header-1", "x-custom-header-2"},
		},
		AgentClient: alwaysReadyClient{},
	}

	sensor = newSensor(opts)
	defer func() {
		sensor = nil
	}()

	fsm.applyHostAgentSettings(response)

	assert.NotContains(t, sensor.options.Tracer.CollectableHTTPHeaders, "my-unwanted-custom-headers")
}

type alwaysReadyClient struct{}

func (alwaysReadyClient) Ready() bool                                       { return true }
func (alwaysReadyClient) SendMetrics(data acceptor.Metrics) error           { return nil }
func (alwaysReadyClient) SendEvent(event *EventData) error                  { return nil }
func (alwaysReadyClient) SendSpans(spans []Span) error                      { return nil }
func (alwaysReadyClient) SendProfiles(profiles []autoprofile.Profile) error { return nil }
func (alwaysReadyClient) Flush(context.Context) error                       { return nil }

func Test_agentS_SendEvent(t *testing.T) {
	tests := []struct {
		name             string
		event            *EventData
		mockResp         *http.Response
		mockErr          error
		expectedError    bool
		expectedErrorMsg string
	}{
		{
			name: "successful event send with warning severity",
			event: &EventData{
				Title:    "Test Event",
				Text:     "This is a test event",
				Duration: 1000,
				Severity: int(SeverityWarning),
				Plugin:   ServicePlugin,
				ID:       "test-service",
				Host:     ServiceHost,
			},
			mockResp: &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader([]byte("{}"))),
			},
			mockErr:       nil,
			expectedError: false,
		},
		{
			name: "successful event send with critical severity",
			event: &EventData{
				Title:    "Critical Event",
				Text:     "Critical issue detected",
				Duration: 5000,
				Severity: int(SeverityCritical),
				Plugin:   ServicePlugin,
				ID:       "critical-service",
				Host:     ServiceHost,
			},
			mockResp: &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader([]byte("{}"))),
			},
			mockErr:       nil,
			expectedError: false,
		},
		{
			name: "successful event send with change severity",
			event: &EventData{
				Title:    "Change Event",
				Text:     "Configuration changed",
				Duration: 2000,
				Severity: int(SeverityChange),
				Plugin:   ServicePlugin,
				ID:       "change-service",
				Host:     ServiceHost,
			},
			mockResp: &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader([]byte("{}"))),
			},
			mockErr:       nil,
			expectedError: false,
		},
		{
			name: "event send with non-200 status code",
			event: &EventData{
				Title:    "Server Error Event",
				Text:     "Server returned error",
				Duration: 1000,
				Severity: int(SeverityWarning),
			},
			mockResp: &http.Response{
				StatusCode: 500,
				Status:     "Internal Server Error",
				Body:       io.NopCloser(bytes.NewReader([]byte("{}"))),
			},
			mockErr:       nil,
			expectedError: false,
		},
		{
			name: "event with empty title",
			event: &EventData{
				Title:    "",
				Text:     "Event with no title",
				Duration: 1000,
				Severity: int(SeverityWarning),
			},
			mockResp: &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader([]byte("{}"))),
			},
			mockErr:       nil,
			expectedError: false,
		},
		{
			name: "event with minimal data",
			event: &EventData{
				Title: "Minimal Event",
			},
			mockResp: &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader([]byte("{}"))),
			},
			mockErr:       nil,
			expectedError: false,
		},
		{
			name: "event with payload too large",
			event: &EventData{
				Title:    "Large Event",
				Text:     strings.Repeat("x", maxContentLength),
				Duration: 1000,
				Severity: int(SeverityWarning),
			},
			mockResp:      nil,
			mockErr:       nil,
			expectedError: true,
		},
		{
			name:             "nil event returns error",
			event:            nil,
			mockResp:         nil,
			mockErr:          nil,
			expectedError:    true,
			expectedErrorMsg: "event cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock HTTP client
			mockClient := &httpClientMock{
				resp: tt.mockResp,
				err:  tt.mockErr,
			}

			// Create agent communicator with mock client
			agentComm := &agentCommunicator{
				host:   "localhost",
				port:   "42699",
				from:   &fromS{EntityID: "12345"},
				client: mockClient,
				l:      defaultLogger,
			}

			// Create agent with mock communicator
			agent := &agentS{
				agentComm: agentComm,
				logger:    defaultLogger,
			}

			// Execute SendEvent
			err := agent.SendEvent(tt.event)

			// Verify results
			if tt.expectedError {
				assert.Error(t, err)
				if tt.expectedErrorMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrorMsg)
				}
			} else {
				assert.NoError(t, err)
			}

		})
	}
}

func Test_agentS_SendEvent_ConcurrentCalls(t *testing.T) {
	// Test that concurrent SendEvent calls work correctly
	// Create a mock client that returns a new response body for each request
	// to avoid data races when multiple goroutines read from the same body
	mockClient := &httpClientMock{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader([]byte("{}"))),
			}, nil
		},
	}

	agentComm := &agentCommunicator{
		host:   "localhost",
		port:   "42699",
		from:   &fromS{EntityID: "12345"},
		client: mockClient,
		l:      defaultLogger,
	}

	agent := &agentS{
		agentComm: agentComm,
		logger:    defaultLogger,
	}

	// Send multiple events concurrently
	const numEvents = 10
	errChan := make(chan error, numEvents)

	for i := 0; i < numEvents; i++ {
		go func(index int) {
			event := &EventData{
				Title:    "Concurrent Event " + strconv.Itoa(index),
				Text:     "Testing concurrent sends",
				Duration: 1000,
				Severity: int(SeverityWarning),
			}
			errChan <- agent.SendEvent(event)
		}(i)
	}

	// Collect all results
	for i := 0; i < numEvents; i++ {
		err := <-errChan
		assert.NoError(t, err)
	}
}

func TestAgent_SendSpans_IPv6Support(t *testing.T) {
	// Create a test server that listens on IPv6
	listener, err := net.Listen("tcp6", "[::1]:0") // IPv6 localhost
	if err != nil {
		t.Skipf("IPv6 not available on this system: %v", err)
	}
	defer listener.Close()

	requestReceived := make(chan bool, 1)
	var receivedPath string

	// Create HTTP handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path

		// Verify it's a valid traces request
		if !strings.HasPrefix(r.URL.Path, agentTracesURL) {
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		requestReceived <- true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
	})

	// Start server with IPv6 listener
	server := &http.Server{Handler: handler}
	go server.Serve(listener)
	defer server.Close()

	// Get the IPv6 address with port
	addr := listener.Addr().(*net.TCPAddr)
	ipv6Host := fmt.Sprintf("[%s]", addr.IP)
	ipv6Port := strconv.Itoa(addr.Port)

	t.Logf("Test server listening on IPv6: %s:%s", ipv6Host, ipv6Port)

	// Create agent with IPv6 endpoint
	agentComm := &agentCommunicator{
		host: ipv6Host,
		port: ipv6Port,
		from: &fromS{EntityID: "12345"},
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		l: defaultLogger,
	}

	agent := &agentS{
		agentComm: agentComm,
		logger:    defaultLogger,
	}

	// Send test spans
	testSpans := []Span{
		{
			TraceID: 12345,
			SpanID:  67890,
			Name:    "test-operation",
		},
	}

	err = agent.SendSpans(testSpans)
	if err != nil {
		t.Fatalf("failed to send spans: %v", err)
	}

	// Wait for request to be received
	select {
	case <-requestReceived:
		t.Logf("Successfully sent HTTP request to IPv6 endpoint")

		// Verify correct endpoint was called
		if !strings.HasPrefix(receivedPath, agentTracesURL) {
			t.Errorf("Expected path to start with %s, got %s", agentTracesURL, receivedPath)
		}

		t.Logf("IPv6 endpoint support verified: SendSpans() successfully communicates with IPv6 addresses")
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for request - IPv6 communication failed")
	}
}

func TestAgent_SendMetrics_IPv6Support(t *testing.T) {
	// Create a test server that listens on IPv6
	listener, err := net.Listen("tcp6", "[::1]:0")
	if err != nil {
		t.Skipf("IPv6 not available on this system: %v", err)
	}
	defer listener.Close()

	requestReceived := make(chan bool, 1)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, agentDataURL) {
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		requestReceived <- true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
	})

	server := &http.Server{Handler: handler}
	go server.Serve(listener)
	defer server.Close()

	addr := listener.Addr().(*net.TCPAddr)
	ipv6Host := fmt.Sprintf("[%s]", addr.IP)
	ipv6Port := strconv.Itoa(addr.Port)

	agentComm := &agentCommunicator{
		host: ipv6Host,
		port: ipv6Port,
		from: &fromS{EntityID: "12345"},
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		l: defaultLogger,
	}

	agent := &agentS{
		agentComm: agentComm,
		snapshot: &SnapshotCollector{
			CollectionInterval: snapshotCollectionInterval,
			ServiceName:        "test-service",
		},
		logger: defaultLogger,
	}

	// Send test metrics
	testMetrics := acceptor.Metrics{
		CgoCall:   100,
		Goroutine: 200,
	}

	err = agent.SendMetrics(testMetrics)
	if err != nil {
		t.Fatalf("failed to send metrics: %v", err)
	}

	select {
	case <-requestReceived:
		t.Logf("IPv6 endpoint support verified: SendMetrics() successfully communicates with IPv6 addresses")
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for request - IPv6 communication failed")
	}
}

func TestAgent_SendProfiles_IPv6Support(t *testing.T) {
	listener, err := net.Listen("tcp6", "[::1]:0")
	if err != nil {
		t.Skipf("IPv6 not available on this system: %v", err)
	}
	defer listener.Close()

	requestReceived := make(chan bool, 1)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, agentProfilesURL) {
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		requestReceived <- true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
	})

	server := &http.Server{Handler: handler}
	go server.Serve(listener)
	defer server.Close()

	addr := listener.Addr().(*net.TCPAddr)
	ipv6Host := fmt.Sprintf("[%s]", addr.IP)
	ipv6Port := strconv.Itoa(addr.Port)

	agentComm := &agentCommunicator{
		host: ipv6Host,
		port: ipv6Port,
		from: &fromS{EntityID: "12345"},
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		l: defaultLogger,
	}

	agent := &agentS{
		agentComm: agentComm,
		logger:    defaultLogger,
	}

	// Send test profiles
	testProfiles := []autoprofile.Profile{
		{
			Type:     "cpu",
			Duration: 1000,
		},
	}

	err = agent.SendProfiles(testProfiles)
	if err != nil {
		t.Fatalf("failed to send profiles: %v", err)
	}

	select {
	case <-requestReceived:
		t.Logf("IPv6 endpoint support verified: SendProfiles() successfully communicates with IPv6 addresses")
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for request - IPv6 communication failed")
	}
}

func TestAgent_IPv4vsIPv6(t *testing.T) {
	tests := []struct {
		name        string
		setupServer func() (string, string, func())
	}{
		{
			name: "IPv4",
			setupServer: func() (string, string, func()) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("{}"))
				}))

				// Extract host and port from server URL
				// server.URL format is "http://127.0.0.1:port"
				host := "127.0.0.1"
				port := server.URL[strings.LastIndex(server.URL, ":")+1:]

				return host, port, server.Close
			},
		},
		{
			name: "IPv6",
			setupServer: func() (string, string, func()) {
				listener, err := net.Listen("tcp6", "[::1]:0")
				if err != nil {
					t.Skipf("IPv6 not available: %v", err)
				}

				server := &http.Server{
					Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusOK)
						w.Write([]byte("{}"))
					}),
				}
				go server.Serve(listener)

				addr := listener.Addr().(*net.TCPAddr)
				host := fmt.Sprintf("[%s]", addr.IP)
				port := strconv.Itoa(addr.Port)

				return host, port, func() {
					server.Close()
					listener.Close()
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, port, cleanup := tt.setupServer()
			defer cleanup()

			agentComm := &agentCommunicator{
				host: host,
				port: port,
				from: &fromS{EntityID: "12345"},
				client: &http.Client{
					Timeout: 5 * time.Second,
				},
				l: defaultLogger,
			}

			agent := &agentS{
				agentComm: agentComm,
				snapshot: &SnapshotCollector{
					CollectionInterval: snapshotCollectionInterval,
					ServiceName:        "test-service",
				},
				logger: defaultLogger,
			}

			// Test SendSpans
			testSpans := []Span{
				{
					TraceID: 12345,
					SpanID:  67890,
					Name:    "test-operation",
				},
			}

			err := agent.SendSpans(testSpans)
			if err != nil {
				t.Fatalf("SendSpans failed for %s: %v", tt.name, err)
			}

			// Test SendMetrics
			testMetrics := acceptor.Metrics{
				CgoCall: 100,
			}

			err = agent.SendMetrics(testMetrics)
			if err != nil {
				t.Fatalf("SendMetrics failed for %s: %v", tt.name, err)
			}

			t.Logf("%s endpoint works correctly", tt.name)
		})
	}
}
