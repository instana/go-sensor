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

	surl := server.URL
	u, err := url.Parse(surl)

	assert.NoError(t, err)

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

	surl := server.URL
	u, err := url.Parse(surl)

	assert.NoError(t, err)

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
			io.WriteString(w, `{"pid":`+pid+`}`)
		} else {
			serverGaveErrorOnFirstCall = true
			w.WriteHeader(http.StatusBadRequest)
		}
	})
	defer server.Close()

	surl := server.URL
	u, err := url.Parse(surl)

	assert.NoError(t, err)

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

	surl := server.URL
	u, err := url.Parse(surl)

	assert.NoError(t, err)

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

	surl := server.URL
	u, err := url.Parse(surl)

	assert.NoError(t, err)

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
			io.WriteString(w, agentResponseJSON)
			return
		}

		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	surl := server.URL
	u, err := url.Parse(surl)

	os.Setenv("INSTANA_AGENT_HOST", u.Hostname())
	defer func() {
		os.Unsetenv("INSTANA_AGENT_HOST")
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

	r.fsm.Event(context.Background(), eInit)

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

func Test_exponentialDelay(t *testing.T) {
	testcases := map[string]struct {
		Input            int
		ExpectedOutputMs time.Duration
	}{
		"retry1": {
			Input:            1,
			ExpectedOutputMs: 10000,
		},
		"retry2": {
			Input:            2,
			ExpectedOutputMs: 20000,
		},
		"retry3": {
			Input:            3,
			ExpectedOutputMs: 40000,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			output := expDelay(testcase.Input)
			assert.Equal(t, output, testcase.ExpectedOutputMs*time.Millisecond, "The actual delay varies from the expected delay")
		})
	}

}

func Test_fsmS_cpusetFileContent(t *testing.T) {
	r := &fsmS{
		agentComm:   newAgentCommunicator("123", "456", &fromS{}, defaultLogger),
		fsm:         f.NewFSM("", []f.EventDesc{}, map[string]f.Callback{}),
		retriesLeft: maximumRetries,
		expDelayFunc: func(retryNumber int) time.Duration {
			return 0
		},
		logger: defaultLogger,
	}

	data := r.cpusetFileContent(3)
	assert.Equal(t, data, "")
}

func Test_fsmS_lookupDefaultGateway(t *testing.T) {
	server := getTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})

	rawURL := server.URL

	serverURL, err := url.Parse(rawURL)
	assert.NoError(t, err, "failed to parse the URL")

	agc := newAgentCommunicator(serverURL.Host, serverURL.Port(), &fromS{}, defaultLogger)

	newFSM(agc, defaultLogger)

	//r := &fsmS{
	//	agentComm:   newAgentCommunicator(serverURL.Host, serverURL.Port(), &fromS{}, defaultLogger),
	//	fsm:         f.NewFSM(eInit, []f.EventDesc{}, map[string]f.Callback{}),
	//	retriesLeft: maximumRetries,
	//	expDelayFunc: func(retryNumber int) time.Duration {
	//		return 0
	//	},
	//	logger: defaultLogger,
	//}

}
