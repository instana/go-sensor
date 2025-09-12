// (c) Copyright IBM Corp. 2022

package instana

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/instana/go-sensor/secrets"
	f "github.com/looplab/fsm"
	"github.com/stretchr/testify/assert"
)

type testLogger struct {
	infoMsg string
	errMsg  string
}

func (tl *testLogger) Debug(v ...interface{}) {}
func (tl *testLogger) Info(v ...interface{}) {
	tl.infoMsg = fmt.Sprint(v...)
}
func (tl *testLogger) Warn(v ...interface{}) {}
func (tl *testLogger) Error(v ...interface{}) {
	tl.errMsg = fmt.Sprint(v...)
}

func getTestServer(fn func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	handler := http.NewServeMux()
	handler.HandleFunc("/", fn)
	return httptest.NewServer(handler)
}

// Case: Current state is ANNOUNCED, trying to be READY
func Test_fsmS_testAgent(t *testing.T) {
	// Forces the mocked agent to fail with HTTP 400 in the first call to lead fsm to retry once
	var serverGaveErrorOnFirstCall bool

	server := getTestServer(func(w http.ResponseWriter, r *http.Request) {
		if serverGaveErrorOnFirstCall {
			w.WriteHeader(http.StatusOK)
		} else {
			serverGaveErrorOnFirstCall = true
			w.WriteHeader(http.StatusBadRequest)
		}
	})

	defer server.Close()

	serverURL := server.URL
	u, err := url.Parse(serverURL)
	assert.NoError(t, err, "failed to parse the URL")

	res := make(chan bool, 1)

	r := &fsmS{
		agentComm: newAgentCommunicator(u.Hostname(), u.Port(), &fromS{EntityID: "12345"}, defaultLogger),
		fsm: f.NewFSM(
			"announced",
			f.Events{
				{Name: eTest, Src: []string{"announced"}, Dst: "ready"},
			},
			f.Callbacks{
				"ready": func(_ context.Context, event *f.Event) {
					res <- true
				},
			}),
		retriesLeft: maximumRetries,
		expDelayFunc: func(retryNumber int) time.Duration {
			return 0
		},
		logger: defaultLogger,
	}

	r.testAgent(context.Background(), &f.Event{})

	assert.True(t, <-res)
	// after a successful request, retriesLeft is reset to maximumRetries
	assert.Equal(t, maximumRetries, r.retriesLeft)
}

func Test_fsmS_testAgent_Error(t *testing.T) {
	// Forces the mocked agent to fail with HTTP 400 to lead fsm to retry
	server := getTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})

	defer server.Close()

	serverURL := server.URL
	u, err := url.Parse(serverURL)
	assert.NoError(t, err, "failed to parse the URL")

	res := make(chan bool, 1)

	r := &fsmS{
		agentComm: newAgentCommunicator(u.Hostname(), u.Port(), &fromS{}, defaultLogger),
		fsm: f.NewFSM(
			"announced",
			f.Events{
				{Name: eInit, Src: []string{"announced"}, Dst: "init"}},
			f.Callbacks{
				"init": func(_ context.Context, event *f.Event) {
					res <- true
				},
			}),
		retriesLeft: maximumRetries,
		expDelayFunc: func(retryNumber int) time.Duration {
			return 0
		},
		logger: defaultLogger,
	}

	r.testAgent(context.Background(), &f.Event{})

	assert.True(t, <-res)
	assert.Equal(t, 0, r.retriesLeft)
}

func Test_fsmS_announceSensor(t *testing.T) {
	// initializes the global sensor as it is needed when the announcement is successful
	InitCollector(DefaultOptions())
	defer ShutdownCollector()

	// Forces the mocked agent to fail with HTTP 400 in the first call to lead fsm to retry once
	var serverGaveErrorOnFirstCall bool

	server := getTestServer(func(w http.ResponseWriter, r *http.Request) {
		if serverGaveErrorOnFirstCall {
			pid := strconv.FormatInt(int64(os.Getpid()), 10)
			_, _ = io.WriteString(w, `{"pid":`+pid+`}`)
		} else {
			serverGaveErrorOnFirstCall = true
			w.WriteHeader(http.StatusBadRequest)
		}
	})
	defer server.Close()

	serverURL := server.URL
	u, err := url.Parse(serverURL)
	assert.NoError(t, err, "failed  to parse the URL")

	res := make(chan bool, 1)

	r := &fsmS{
		agentComm: newAgentCommunicator(u.Hostname(), u.Port(), &fromS{}, defaultLogger),
		fsm: f.NewFSM(
			"unannounced",
			f.Events{
				{Name: eAnnounce, Src: []string{"unannounced"}, Dst: "announced"}},
			f.Callbacks{
				"announced": func(_ context.Context, event *f.Event) {
					res <- true
				},
			}),
		retriesLeft: maximumRetries,
		expDelayFunc: func(retryNumber int) time.Duration {
			return 0
		},
		logger: defaultLogger,
	}

	r.announceSensor(context.Background(), &f.Event{})

	assert.True(t, <-res)
	assert.Equal(t, maximumRetries, r.retriesLeft)
}

func Test_fsmS_announceSensor_Error(t *testing.T) {
	server := getTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})
	defer server.Close()

	serverURL := server.URL
	u, err := url.Parse(serverURL)
	assert.NoError(t, err, "failed to parse the URL")

	res := make(chan bool, 1)

	r := &fsmS{
		agentComm: newAgentCommunicator(u.Hostname(), u.Port(), &fromS{}, defaultLogger),
		fsm: f.NewFSM(
			"unannounced",
			f.Events{
				{Name: eInit, Src: []string{"unannounced"}, Dst: "init"}},
			f.Callbacks{
				"init": func(_ context.Context, event *f.Event) {
					res <- true
				},
			}),
		retriesLeft: maximumRetries,
		expDelayFunc: func(retryNumber int) time.Duration {
			return 0
		},
		logger: defaultLogger,
	}

	r.announceSensor(context.Background(), &f.Event{})

	assert.True(t, <-res)
	assert.Equal(t, 0, r.retriesLeft)
}

func Test_fsmS_lookupAgentHost(t *testing.T) {
	// Forces the mocked agent to fail with HTTP 400 in the first call to lead fsm to retry once
	var serverGaveErrorOnFirstCall bool

	server := getTestServer(func(w http.ResponseWriter, r *http.Request) {
		if serverGaveErrorOnFirstCall {
			w.Header().Add("Server", agentHeader)
			w.WriteHeader(http.StatusOK)
		} else {
			serverGaveErrorOnFirstCall = true
			w.WriteHeader(http.StatusBadRequest)
		}
	})
	defer server.Close()

	serverURL := server.URL
	u, err := url.Parse(serverURL)
	assert.NoError(t, err, "failed to parse the URL")

	res := make(chan bool, 1)

	r := &fsmS{
		agentComm:                  newAgentCommunicator(u.Hostname(), u.Port(), &fromS{}, defaultLogger),
		lookupAgentHostRetryPeriod: 0,
		fsm: f.NewFSM(
			"init",
			f.Events{
				{Name: eLookup, Src: []string{"init"}, Dst: "unannounced"}},
			f.Callbacks{
				"enter_unannounced": func(_ context.Context, event *f.Event) {
					res <- true
				},
			}),
		retriesLeft: maximumRetries,
		expDelayFunc: func(retryNumber int) time.Duration {
			return 0
		},
		logger: defaultLogger,
	}

	r.lookupAgentHost(context.Background(), &f.Event{})

	assert.True(t, <-res)
	assert.Equal(t, maximumRetries, r.retriesLeft)
}

// Case:
// 1. Connection between the application and the agent is established.
// 2. Connection with Agent is lost.
// 3. Former Agent host is no longer available.
// 4. A valid Agent hostname is available in the INSTANA_AGENT_HOST env var.
// 5. A connection with the Agent is reestablished via env var.
func Test_fsmS_agentConnectionReestablished(t *testing.T) {
	agentResponseJSON := `{
		"pid": 37808,
		"agentUuid": "88:66:5a:ff:fe:05:a5:f0",
		"extraHeaders": ["expected-value"],
		"secrets": {
			"matcher": "contains-ignore-case",
			"list": ["key","pass","secret"]
		}
	}`

	sensor = &sensorS{
		options: DefaultOptions(),
	}
	defer func() {
		sensor = nil
	}()

	server := getTestServer(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path

		// announce phase (enter_unannounced)
		if p == "/com.instana.plugin.golang.discovery" {
			_, _ = io.WriteString(w, agentResponseJSON)
			return
		}

		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	serverURL := server.URL
	u, err := url.Parse(serverURL)
	assert.NoError(t, err, "failed to parse the URL")

	err = os.Setenv("INSTANA_AGENT_HOST", u.Hostname())
	assert.NoError(t, err, "failed to set the env variable")

	defer func() {
		err = os.Unsetenv("INSTANA_AGENT_HOST")
		assert.NoError(t, err, "failed to set the env variable")
	}()

	assert.NoError(t, err)

	res := make(chan bool)

	r := &fsmS{
		agentComm:                  newAgentCommunicator(u.Hostname(), u.Port(), &fromS{EntityID: "12345"}, defaultLogger),
		lookupAgentHostRetryPeriod: 0,
		retriesLeft:                maximumRetries,
		expDelayFunc: func(retryNumber int) time.Duration {
			return 0
		},
		logger: defaultLogger,
	}

	r.fsm = f.NewFSM(
		"none",
		f.Events{
			{Name: eInit, Src: []string{"none", "unannounced", "announced", "ready"}, Dst: "init"},
			{Name: eLookup, Src: []string{"init"}, Dst: "unannounced"},
			{Name: eAnnounce, Src: []string{"unannounced"}, Dst: "announced"},
			{Name: eTest, Src: []string{"announced"}, Dst: "ready"}},
		f.Callbacks{
			"init": func(ctx context.Context, e *f.Event) {
				r.lookupAgentHost(ctx, e)
			},
			"enter_unannounced": func(ctx context.Context, e *f.Event) {
				r.announceSensor(ctx, e)
			},
			"enter_announced": func(ctx context.Context, e *f.Event) {
				r.testAgent(ctx, e)
			},
			"ready": func(ctx context.Context, e *f.Event) {
				r.ready(ctx, e)
				res <- true
			},
		})

	// We fail the test if the channel does not resolve after 5 seconds
	go func() {
		time.AfterFunc(time.Second*5, func() {
			res <- false
		})
	}()

	err = r.fsm.Event(context.Background(), eInit)
	assert.NoError(t, err, "failed to initiate the state transition")

	assert.True(t, <-res)

	// Simulate Agent connection lost
	r.agentComm.host = "invalid_host"
	r.reset()

	assert.True(t, <-res)
	assert.Equal(t, os.Getenv("INSTANA_AGENT_HOST"), r.agentComm.host, "Configured host to be updated with env var value")
}

func Test_fsmS_applyHostAgentSettings_agent_Override(t *testing.T) {

	opts := DefaultOptions()
	opts.Tracer.tracerDefaultSecrets = true
	opts.Tracer.Secrets = secrets.NewContainsMatcher("test")

	sensor = &sensorS{
		options: opts,
	}
	defer func() {
		sensor = nil
	}()

	r := &fsmS{
		agentComm:   newAgentCommunicator("123", "456", &fromS{}, defaultLogger),
		fsm:         f.NewFSM("", []f.EventDesc{}, map[string]f.Callback{}),
		retriesLeft: maximumRetries,
		expDelayFunc: func(retryNumber int) time.Duration {
			return 0
		},
		logger: defaultLogger,
	}

	resp := agentResponse{
		Pid:    1234,
		HostID: "45664w32",
		Secrets: struct {
			Matcher string   "json:\"matcher\""
			List    []string "json:\"list\""
		}{
			Matcher: "contains-ignore-case",
			List:    []string{"key123", "pass123", "secret123"},
		},
		ExtraHTTPHeaders: []string{"abc", "def"},
	}

	r.applyHostAgentSettings(resp)

	assert.Equal(t, false, sensor.options.Tracer.Secrets.Match("test"))
	assert.Equal(t, true, sensor.options.Tracer.Secrets.Match("key123"))

	assert.Equal(t, []string{"abc", "def"}, sensor.options.Tracer.CollectableHTTPHeaders)

}

func Test_fsmS_applyHostAgentSettings_agent_NotOverride(t *testing.T) {

	opts := DefaultOptions()
	opts.Tracer.tracerDefaultSecrets = false
	opts.Tracer.Secrets = secrets.NewContainsMatcher("test")
	opts.Tracer.CollectableHTTPHeaders = []string{"testHeader"}

	sensor = &sensorS{
		options: opts,
	}
	defer func() {
		sensor = nil
	}()

	tLogger := &testLogger{}
	r := &fsmS{
		agentComm:   newAgentCommunicator("123", "456", &fromS{}, defaultLogger),
		fsm:         f.NewFSM("", []f.EventDesc{}, map[string]f.Callback{}),
		retriesLeft: maximumRetries,
		expDelayFunc: func(retryNumber int) time.Duration {
			return 0
		},
		logger: tLogger,
	}

	resp := agentResponse{
		Pid:    1234,
		HostID: "45664w32",
		Secrets: struct {
			Matcher string   "json:\"matcher\""
			List    []string "json:\"list\""
		}{
			Matcher: "contains-ignore-case",
			List:    []string{"key123", "pass123", "secret123"},
		},
		ExtraHTTPHeaders: []string{"abc", "def"},
	}

	r.applyHostAgentSettings(resp)

	assert.Equal(t, true, sensor.options.Tracer.Secrets.Match("test"))
	assert.Equal(t, false, sensor.options.Tracer.Secrets.Match("key123"))
	assert.Equal(t, "identified custom defined secrets matcher. Ignoring host agent default secrets configuration.", tLogger.infoMsg)

	assert.Equal(t, []string{"testHeader"}, sensor.options.Tracer.CollectableHTTPHeaders)

}

func Test_fsmS_applyHostAgentSettings_agent_secrets_empty_Error(t *testing.T) {

	opts := DefaultOptions()
	opts.Tracer.tracerDefaultSecrets = true
	opts.Tracer.CollectableHTTPHeaders = []string{"testHeader"}

	sensor = &sensorS{
		options: opts,
	}
	defer func() {
		sensor = nil
	}()

	tLogger := &testLogger{}
	r := &fsmS{
		agentComm:   newAgentCommunicator("123", "456", &fromS{}, defaultLogger),
		fsm:         f.NewFSM("", []f.EventDesc{}, map[string]f.Callback{}),
		retriesLeft: maximumRetries,
		expDelayFunc: func(retryNumber int) time.Duration {
			return 0
		},
		logger: tLogger,
	}

	resp := agentResponse{
		Pid:    1234,
		HostID: "45664w32",
		Secrets: struct {
			Matcher string   "json:\"matcher\""
			List    []string "json:\"list\""
		}{
			Matcher: "",
			List:    []string{},
		},
		ExtraHTTPHeaders: []string{"abc", "def"},
	}

	r.applyHostAgentSettings(resp)

	assert.Equal(t, false, sensor.options.Tracer.Secrets.Match("test"))
	assert.Equal(t, true, sensor.options.Tracer.Secrets.Match("key123"))
	assert.Equal(t, "invalid host agent secret matcher config: secrets-matcher:  secrets-list: []", tLogger.errMsg)

	assert.Equal(t, []string{"testHeader"}, sensor.options.Tracer.CollectableHTTPHeaders)

}

func Test_fsmS_applyHostAgentSettings_agent_secrets_not_valid_Error(t *testing.T) {

	opts := DefaultOptions()
	opts.Tracer.tracerDefaultSecrets = true
	opts.Tracer.CollectableHTTPHeaders = []string{"testHeader"}

	sensor = &sensorS{
		options: opts,
	}
	defer func() {
		sensor = nil
	}()

	tLogger := &testLogger{}
	r := &fsmS{
		agentComm:   newAgentCommunicator("123", "456", &fromS{}, defaultLogger),
		fsm:         f.NewFSM("", []f.EventDesc{}, map[string]f.Callback{}),
		retriesLeft: maximumRetries,
		expDelayFunc: func(retryNumber int) time.Duration {
			return 0
		},
		logger: tLogger,
	}

	resp := agentResponse{
		Pid:    1234,
		HostID: "45664w32",
		Secrets: struct {
			Matcher string   "json:\"matcher\""
			List    []string "json:\"list\""
		}{
			Matcher: "test_error_matcher",
			List:    []string{"test"},
		},
		ExtraHTTPHeaders: []string{"abc", "def"},
	}

	r.applyHostAgentSettings(resp)

	assert.Equal(t, false, sensor.options.Tracer.Secrets.Match("test"))
	assert.Equal(t, true, sensor.options.Tracer.Secrets.Match("key123"))
	assert.Equal(t, "failed to apply secrets matcher configuration: unknown secrets matcher type \"test_error_matcher\"", tLogger.errMsg)

	assert.Equal(t, []string{"testHeader"}, sensor.options.Tracer.CollectableHTTPHeaders)

}

func TestApplyDisableTracingConfig(t *testing.T) {
	tests := []struct {
		name            string
		envVarSet       bool
		initialDisable  map[string]bool
		agentConfig     []map[string]bool
		expectedDisable map[string]bool
	}{
		{
			name:            "Empty agent config",
			agentConfig:     []map[string]bool{},
			initialDisable:  map[string]bool{},
			expectedDisable: map[string]bool{},
		},
		{
			name: "in-code config exists",
			agentConfig: []map[string]bool{
				{"logging": true},
			},
			initialDisable: map[string]bool{
				"logging": true,
			},
			expectedDisable: map[string]bool{"logging": true},
		},
		{
			name: "apply agent config",
			agentConfig: []map[string]bool{
				{"logging": true},
			},
			initialDisable:  map[string]bool{},
			expectedDisable: map[string]bool{"logging": true},
		},
		{
			name:            "env variable set",
			envVarSet:       true,
			agentConfig:     []map[string]bool{},
			initialDisable:  map[string]bool{},
			expectedDisable: map[string]bool{"logging": true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.envVarSet {
				envVarValue := "logging"
				os.Setenv("INSTANA_TRACING_DISABLE", envVarValue)
				defer os.Unsetenv("INSTANA_TRACING_DISABLE")
			}

			InitCollector(&Options{
				Tracer: TracerOptions{
					Disable: tt.initialDisable,
				},
			})
			defer ShutdownCollector()

			resp := agentResponse{}
			resp.Tracing.Disable = tt.agentConfig

			testLogger := &testLogger{}

			fsm := &fsmS{
				logger: testLogger,
			}
			fsm.applyDisableTracingConfig(resp)

			// Check if the maps have the same size
			assert.Equal(t, len(tt.expectedDisable), len(sensor.options.Tracer.Disable),
				"Expected map size %d, got %d", len(tt.expectedDisable), len(sensor.options.Tracer.Disable))

			// Check if all expected keys are present with correct values
			for k, v := range tt.expectedDisable {
				actualValue, exists := sensor.options.Tracer.Disable[k]
				assert.True(t, exists, "Expected key %s not found in result", k)
				assert.Equal(t, v, actualValue, "Expected %s to be %v, got %v", k, v, actualValue)
			}

			// Check if there are no unexpected keys
			for k := range sensor.options.Tracer.Disable {
				_, exists := tt.expectedDisable[k]
				assert.True(t, exists, "Unexpected key in result: %s", k)
			}
		})
	}
}
