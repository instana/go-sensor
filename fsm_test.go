// (c) Copyright IBM Corp. 2022

package instana

import (
	"errors"
	"github.com/instana/testify/assert"
	f "github.com/looplab/fsm"
	"testing"
	"time"
)

func Test_fsmS_testAgent(t *testing.T) {
	// init channels for agent mock
	rCh := make(chan string, 2)
	errCh := make(chan error, 2)

	res := make(chan bool, 1)

	r := &fsmS{
		agent: &mockFsmAgent{
			headRequestResponse: rCh,
			headRequestErr:      errCh,
		},
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

	// simulate errors and successful requests
	rCh <- ""
	errCh <- errors.New("some error")

	rCh <- "Hello"
	errCh <- nil

	r.testAgent(&f.Event{})

	assert.True(t, <-res)
	assert.Empty(t, rCh)
	assert.Empty(t, errCh)
	assert.Equal(t, maximumRetries, r.retriesLeft)
}

func Test_fsmS_testAgent_Error(t *testing.T) {
	// init channels for agent mock
	rCh := make(chan string, 3)
	errCh := make(chan error, 3)

	res := make(chan bool, 1)

	r := &fsmS{
		agent: &mockFsmAgent{
			headRequestResponse: rCh,
			headRequestErr:      errCh,
		},
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

	// simulate errors
	rCh <- ""
	errCh <- errors.New("error #1")
	rCh <- ""
	errCh <- errors.New("error #2")
	rCh <- ""
	errCh <- errors.New("error #3")

	r.testAgent(&f.Event{})

	assert.True(t, <-res)
	assert.Empty(t, rCh)
	assert.Empty(t, errCh)
	assert.Equal(t, 0, r.retriesLeft)
}

func Test_fsmS_announceSensor(t *testing.T) {
	// init channels for agent mock
	rCh := make(chan string, 2)
	errCh := make(chan error, 2)

	res := make(chan bool, 1)

	r := &fsmS{
		agent: &mockFsmAgent{
			announceRequestResponse: rCh,
			announceRequestErr:      errCh,
		},
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

	// simulate errors and successful requests
	rCh <- ""
	errCh <- errors.New("some error")

	rCh <- "Hello"
	errCh <- nil

	r.announceSensor(&f.Event{})

	assert.True(t, <-res)
	assert.Empty(t, rCh)
	assert.Empty(t, errCh)
	assert.Equal(t, maximumRetries, r.retriesLeft)
}

func Test_fsmS_announceSensor_Error(t *testing.T) {
	// init channels for agent mock
	rCh := make(chan string, 3)
	errCh := make(chan error, 3)

	res := make(chan bool, 1)

	r := &fsmS{
		agent: &mockFsmAgent{
			announceRequestResponse: rCh,
			announceRequestErr:      errCh,
		},
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

	// simulate errors
	rCh <- ""
	errCh <- errors.New("error #1")
	rCh <- ""
	errCh <- errors.New("error #2")
	rCh <- ""
	errCh <- errors.New("error #3")

	r.announceSensor(&f.Event{})

	assert.True(t, <-res)
	assert.Empty(t, rCh)
	assert.Empty(t, errCh)
	assert.Equal(t, 0, r.retriesLeft)
}

func Test_fsmS_lookupAgentHost(t *testing.T) {
	// init channels for agent mock
	rCh := make(chan string, 2)
	errCh := make(chan error, 2)

	res := make(chan bool, 1)

	r := &fsmS{
		agent: &mockFsmAgent{
			requestHeaderResponse: rCh,
			requestHeaderErr:      errCh,
		},
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

	// simulate errors and successful requests
	rCh <- ""
	errCh <- errors.New("some error")

	rCh <- agentHeader
	errCh <- nil

	r.lookupAgentHost(&f.Event{})

	assert.True(t, <-res)
	assert.Empty(t, rCh)
	assert.Empty(t, errCh)
	assert.Equal(t, maximumRetries, r.retriesLeft)
}

type mockFsmAgent struct {
	host string

	requestHeaderResponse chan string
	requestHeaderErr      chan error

	announceRequestResponse chan string
	announceRequestErr      chan error

	headRequestResponse chan string
	headRequestErr      chan error
}

func (a *mockFsmAgent) getHost() string {
	return a.host
}

func (a *mockFsmAgent) setHost(host string) {
	a.host = host
}

func (a *mockFsmAgent) requestHeader(url string, method string, header string) (string, error) {
	return <-a.requestHeaderResponse, <-a.requestHeaderErr
}

func (a *mockFsmAgent) makeHostURL(host string, prefix string) string {
	return "http://" + host + ":5555" + prefix
}

func (a *mockFsmAgent) applyHostAgentSettings(resp agentResponse) {
	return
}

func (a *mockFsmAgent) announceRequest(url string, method string, data interface{}, ret *agentResponse) (string, error) {
	return <-a.announceRequestResponse, <-a.announceRequestErr
}

func (a *mockFsmAgent) makeURL(prefix string) string {
	return a.makeHostURL(a.getHost(), prefix)
}

func (a *mockFsmAgent) head(url string) (string, error) {
	return <-a.headRequestResponse, <-a.headRequestErr
}
