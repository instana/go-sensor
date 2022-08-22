// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2016

package instana

import (
	"fmt"
	"io/ioutil"
	"math"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	f "github.com/looplab/fsm"
)

const (
	eInit     = "init"
	eLookup   = "lookup"
	eAnnounce = "announce"
	eTest     = "test"

	retryPeriod                = 30 * 1000 * time.Millisecond
	exponentialRetryPeriodBase = 10 * 1000 * time.Millisecond
	maximumRetries             = 3
)

type fsmAgent interface {
	getHost() string
	setHost(host string)
	makeURL(prefix string) string
	makeHostURL(host string, prefix string) string
	applyHostAgentSettings(resp agentResponse)
	requestHeader(url string, method string, header string) (string, error)
	announceRequest(url string, method string, data interface{}, ret *agentResponse) (string, error)
	head(url string) (string, error)
}

type fsmS struct {
	name                       string
	agent                      fsmAgent
	fsm                        *f.FSM
	timer                      *time.Timer
	retriesLeft                int
	expDelayFunc               func(retryNumber int) time.Duration
	lookupAgentHostRetryPeriod time.Duration
	logger                     LeveledLogger
}

func newFSM(agent fsmAgent, logger LeveledLogger) *fsmS {
	logger.Warn("Stan is on the scene. Starting Instana instrumentation.")
	logger.Debug("initializing fsm")

	ret := &fsmS{
		agent:                      agent,
		retriesLeft:                maximumRetries,
		expDelayFunc:               expDelay,
		logger:                     logger,
		lookupAgentHostRetryPeriod: retryPeriod,
	}

	ret.fsm = f.NewFSM(
		"none",
		f.Events{
			{Name: eInit, Src: []string{"none", "unannounced", "announced", "ready"}, Dst: "init"},
			{Name: eLookup, Src: []string{"init"}, Dst: "unannounced"},
			{Name: eAnnounce, Src: []string{"unannounced"}, Dst: "announced"},
			{Name: eTest, Src: []string{"announced"}, Dst: "ready"}},
		f.Callbacks{
			"init":              ret.lookupAgentHost,
			"enter_unannounced": ret.announceSensor,
			"enter_announced":   ret.testAgent,
		})
	ret.fsm.Event(eInit)

	return ret
}

func (r *fsmS) scheduleRetry(e *f.Event, cb func(e *f.Event)) {
	r.timer = time.NewTimer(r.lookupAgentHostRetryPeriod)
	go func() {
		<-r.timer.C
		cb(e)
	}()
}

func (r *fsmS) scheduleRetryWithExponentialDelay(e *f.Event, cb func(e *f.Event), retryNumber int) {
	time.Sleep(r.expDelayFunc(retryNumber))
	cb(e)
}

func (r *fsmS) lookupAgentHost(e *f.Event) {
	go r.checkHost(e, r.agent.getHost())
}

func (r *fsmS) checkHost(e *f.Event, host string) {
	r.logger.Debug("checking host ", host)
	header, err := r.agent.requestHeader(r.agent.makeHostURL(host, "/"), "GET", "Server")

	found := err == nil && header == agentHeader

	// Agent host is found through the checkHost method, that attempts to read "Instana Agent" from the response header.
	if found {
		r.lookupSuccess(host)
		return
	}

	if _, fileNotFoundErr := os.Stat("/proc/net/route"); fileNotFoundErr == nil {
		gateway, err := getDefaultGateway("/proc/net/route")
		if err != nil {
			// This will be always the "failed to open /proc/net/route: no such file or directory" error.
			// As this info is not relevant to the customer, we can remove it from the message.
			r.logger.Error("Couldn't open the /proc/net/route file in order to retrieve the default gateway. Scheduling retry.")
			r.scheduleRetry(e, r.lookupAgentHost)

			return
		}

		if gateway == "" {
			r.logger.Error("Couldn't parse the default gateway address from /proc/net/route. Scheduling retry.")
			r.scheduleRetry(e, r.lookupAgentHost)

			return
		}

		if found {
			r.lookupSuccess(gateway)
			return
		}

		r.logger.Error("Cannot connect to the agent through localhost or default gateway. Scheduling retry.")
		r.scheduleRetry(e, r.lookupAgentHost)
	} else {
		r.logger.Error("Cannot connect to the agent. Scheduling retry.")
		r.logger.Debug("Connecting through the default gateway has not been attempted because proc/net/route does not exist.")
		r.scheduleRetry(e, r.lookupAgentHost)
	}
}

func (r *fsmS) lookupSuccess(host string) {
	r.logger.Debug("agent lookup success ", host)

	r.agent.setHost(host)
	r.retriesLeft = maximumRetries
	r.fsm.Event(eLookup)
}

func (r *fsmS) announceSensor(e *f.Event) {
	r.logger.Debug("announcing sensor to the agent")

	go func() {
		defer func() {
			if err := recover(); err != nil {
				r.logger.Debug("Announce recovered:", err)
			}
		}()

		d := r.getDiscoveryS()

		var resp agentResponse
		_, err := r.agent.announceRequest(r.agent.makeURL(agentDiscoveryURL), "PUT", d, &resp)

		if err != nil {
			r.retriesLeft--
			if r.retriesLeft == 0 {
				r.logger.Error("Couldn't announce the sensor after reaching the maximum amount of attempts.")
				r.fsm.Event(eInit)
				return
			} else {
				r.logger.Debug("Cannot announce sensor. Scheduling retry.")
			}

			retryNumber := maximumRetries - r.retriesLeft + 1
			r.scheduleRetryWithExponentialDelay(e, r.announceSensor, retryNumber)

			return
		}

		r.logger.Info("Host agent available. We're in business. Announced pid:", resp.Pid)
		r.agent.applyHostAgentSettings(resp)

		r.retriesLeft = maximumRetries
		r.fsm.Event(eAnnounce)

	}()
}

func (r *fsmS) getDiscoveryS() *discoveryS {
	pid := os.Getpid()
	cpuSetFileContent := ""

	if runtime.GOOS == "linux" {
		cpuSetFileContent = r.cpuSetFileContent(pid)
	}

	d := &discoveryS{
		PID:               pid,
		CPUSetFileContent: cpuSetFileContent,
		Name:              os.Args[0],
		Args:              os.Args[1:],
	}

	if name, args, ok := getProcCommandLine(); ok {
		r.logger.Debug("got cmdline from /proc: ", name)
		d.Name, d.Args = name, args
	} else {
		r.logger.Debug("no /proc, using OS reported cmdline")
	}

	if _, err := os.Stat("/proc"); err == nil {
		if addr, err := net.ResolveTCPAddr("tcp", r.agent.getHost()+":42699"); err == nil {
			if tcpConn, err := net.DialTCP("tcp", nil, addr); err == nil {
				defer tcpConn.Close()

				file, err := tcpConn.File()

				if err != nil {
					r.logger.Error(err)
				} else {
					d.Fd = fmt.Sprintf("%v", file.Fd())

					link := fmt.Sprintf("/proc/%d/fd/%d", os.Getpid(), file.Fd())
					if _, err := os.Stat(link); err == nil {
						d.Inode, _ = os.Readlink(link)
					}
				}
			}
		}
	}

	return d
}

func (r *fsmS) testAgent(e *f.Event) {
	r.logger.Debug("testing communication with the agent")
	go func() {
		_, err := r.agent.head(r.agent.makeURL(agentDataURL))
		b := err == nil

		if b {
			r.retriesLeft = maximumRetries
			r.fsm.Event(eTest)
		} else {
			r.logger.Debug("Agent is not yet ready. Scheduling retry.")
			r.retriesLeft--
			if r.retriesLeft > 0 {
				retryNumber := maximumRetries - r.retriesLeft + 1
				r.scheduleRetryWithExponentialDelay(e, r.testAgent, retryNumber)
			} else {
				r.fsm.Event(eInit)
			}
		}
	}()
}

func (r *fsmS) reset() {
	r.retriesLeft = maximumRetries
	r.fsm.Event(eInit)
}

func (r *fsmS) cpuSetFileContent(pid int) string {
	path := filepath.Join("proc", strconv.Itoa(pid), "cpuset")
	data, err := ioutil.ReadFile(path)
	if err != nil {
		r.logger.Info("error while reading ", path, ":", err.Error())
		return ""
	}

	return string(data)
}

func expDelay(retryNumber int) time.Duration {
	return time.Duration(math.Pow(2, float64(retryNumber-1))) * exponentialRetryPeriodBase
}
