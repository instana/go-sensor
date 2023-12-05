// (c) Copyright IBM Corp. 2022

package instana

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"

	f "github.com/looplab/fsm"
	"github.com/stretchr/testify/assert"
)

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
	InitSensor(DefaultOptions())
	defer ShutdownSensor()

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
