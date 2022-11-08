// (c) Copyright IBM Corp. 2022

package instana

import (
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
		agentComm: newAgentCommunicator(u.Hostname(), u.Port(), &fromS{}),
		fsm: f.NewFSM(
			"announced",
			f.Events{
				{Name: eTest, Src: []string{"announced"}, Dst: "ready"}},
			f.Callbacks{
				"ready": func(event *f.Event) {
					res <- true
				},
			}),
		retriesLeft: maximumRetries,
		expDelayFunc: func(retryNumber int) time.Duration {
			return 0
		},
		logger: defaultLogger,
	}

	r.testAgent(&f.Event{})

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
		agentComm: newAgentCommunicator(u.Hostname(), u.Port(), &fromS{}),
		fsm: f.NewFSM(
			"announced",
			f.Events{
				{Name: eInit, Src: []string{"announced"}, Dst: "init"}},
			f.Callbacks{
				"init": func(event *f.Event) {
					res <- true
				},
			}),
		retriesLeft: maximumRetries,
		expDelayFunc: func(retryNumber int) time.Duration {
			return 0
		},
		logger: defaultLogger,
	}

	r.testAgent(&f.Event{})

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
		agentComm: newAgentCommunicator(u.Hostname(), u.Port(), &fromS{}),
		fsm: f.NewFSM(
			"unannounced",
			f.Events{
				{Name: eAnnounce, Src: []string{"unannounced"}, Dst: "announced"}},
			f.Callbacks{
				"announced": func(event *f.Event) {
					res <- true
				},
			}),
		retriesLeft: maximumRetries,
		expDelayFunc: func(retryNumber int) time.Duration {
			return 0
		},
		logger: defaultLogger,
	}

	r.announceSensor(&f.Event{})

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
		agentComm: newAgentCommunicator(u.Hostname(), u.Port(), &fromS{}),
		fsm: f.NewFSM(
			"unannounced",
			f.Events{
				{Name: eInit, Src: []string{"unannounced"}, Dst: "init"}},
			f.Callbacks{
				"init": func(event *f.Event) {
					res <- true
				},
			}),
		retriesLeft: maximumRetries,
		expDelayFunc: func(retryNumber int) time.Duration {
			return 0
		},
		logger: defaultLogger,
	}

	r.announceSensor(&f.Event{})

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
		agentComm:                  newAgentCommunicator(u.Hostname(), u.Port(), &fromS{}),
		lookupAgentHostRetryPeriod: 0,
		fsm: f.NewFSM(
			"init",
			f.Events{
				{Name: eLookup, Src: []string{"init"}, Dst: "unannounced"}},
			f.Callbacks{
				"enter_unannounced": func(event *f.Event) {
					res <- true
				},
			}),
		retriesLeft: maximumRetries,
		expDelayFunc: func(retryNumber int) time.Duration {
			return 0
		},
		logger: defaultLogger,
	}

	r.lookupAgentHost(&f.Event{})

	assert.True(t, <-res)
	assert.Equal(t, maximumRetries, r.retriesLeft)
}
