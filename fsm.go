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

type fsmS struct {
	agent       *agentS
	fsm         *f.FSM
	timer       *time.Timer
	retriesLeft int
}

func newFSM(agent *agentS) *fsmS {
	agent.logger.Warn("Stan is on the scene. Starting Instana instrumentation.")
	agent.logger.Debug("initializing fsm")

	ret := &fsmS{
		agent:       agent,
		retriesLeft: maximumRetries,
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
	r.timer = time.NewTimer(retryPeriod)
	go func() {
		<-r.timer.C
		cb(e)
	}()
}

func (r *fsmS) scheduleRetryWithExponentialDelay(e *f.Event, cb func(e *f.Event), retryNumber int) {
	time.Sleep(getRetryPeriodForIteration(retryNumber))
	cb(e)
}

func (r *fsmS) lookupAgentHost(e *f.Event) {
	cb := func(found bool, host string) {
		if found {
			r.lookupSuccess(host)
			return
		}

		gateway, err := getDefaultGateway("/proc/net/route")
		if err != nil {
			r.agent.logger.Error("failed to fetch the default gateway, scheduling retry: ", err)
			r.scheduleRetry(e, r.lookupAgentHost)

			return
		}

		if gateway == "" {
			r.agent.logger.Error("default gateway not available, scheduling retry")
			r.scheduleRetry(e, r.lookupAgentHost)

			return
		}

		go r.checkHost(gateway, func(found bool, host string) {
			if found {
				r.lookupSuccess(host)
				return
			}

			r.agent.logger.Error("cannot connect to the agent through localhost or default gateway, scheduling retry")
			r.scheduleRetry(e, r.lookupAgentHost)
		})

	}

	go r.checkHost(r.agent.host, cb)
}

func (r *fsmS) checkHost(host string, cb func(found bool, host string)) {
	r.agent.logger.Debug("checking host ", host)

	header, err := r.agent.requestHeader(r.agent.makeHostURL(host, "/"), "GET", "Server")

	cb(err == nil && header == agentHeader, host)
}

func (r *fsmS) lookupSuccess(host string) {
	r.agent.logger.Debug("agent lookup success ", host)

	r.agent.setHost(host)
	r.retriesLeft = maximumRetries
	r.fsm.Event(eLookup)
}

func (r *fsmS) announceSensor(e *f.Event) {
	cb := func(success bool, resp agentResponse) {
		if !success {
			r.agent.logger.Error("Cannot announce sensor. Scheduling retry.")
			r.retriesLeft--
			if r.retriesLeft == 0 {
				r.fsm.Event(eInit)
				return
			}

			retryNumber := maximumRetries - r.retriesLeft + 1
			go r.scheduleRetryWithExponentialDelay(e, r.announceSensor, retryNumber)

			return
		}

		r.agent.logger.Info("Host agent available. We're in business. Announced pid:", resp.Pid)
		r.agent.applyHostAgentSettings(resp)

		r.retriesLeft = maximumRetries
		r.fsm.Event(eAnnounce)
	}

	r.agent.logger.Debug("announcing sensor to the agent")

	go func(cb func(success bool, resp agentResponse)) {
		defer func() {
			if err := recover(); err != nil {
				r.agent.logger.Debug("Announce recovered:", err)
			}
		}()

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
			r.agent.logger.Debug("got cmdline from /proc: ", name, args)
			d.Name, d.Args = name, args
		} else {
			r.agent.logger.Debug("no /proc, using OS reported cmdline")
		}

		if _, err := os.Stat("/proc"); err == nil {
			if addr, err := net.ResolveTCPAddr("tcp", r.agent.host+":42699"); err == nil {
				if tcpConn, err := net.DialTCP("tcp", nil, addr); err == nil {
					defer tcpConn.Close()

					file, err := tcpConn.File()

					if err != nil {
						r.agent.logger.Error(err)
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

		var resp agentResponse
		_, err := r.agent.announceRequest(r.agent.makeURL(agentDiscoveryURL), "PUT", d, &resp)
		cb(err == nil, resp)
	}(cb)
}

func (r *fsmS) testAgent(e *f.Event) {
	cb := func(b bool) {
		if b {
			r.retriesLeft = maximumRetries
			r.fsm.Event(eTest)
		} else {
			r.agent.logger.Debug("Agent is not yet ready. Scheduling retry.")
			r.retriesLeft--
			if r.retriesLeft > 0 {
				retryNumber := maximumRetries - r.retriesLeft + 1
				go r.scheduleRetryWithExponentialDelay(e, r.testAgent, retryNumber)
			} else {
				r.fsm.Event(eInit)
			}
		}
	}

	r.agent.logger.Debug("testing communication with the agent")

	go func(cb func(b bool)) {
		_, err := r.agent.head(r.agent.makeURL(agentDataURL))
		cb(err == nil)
	}(cb)
}

func (r *fsmS) reset() {
	r.retriesLeft = maximumRetries
	r.fsm.Event(eInit)
}

func (r *fsmS) cpuSetFileContent(pid int) string {
	path := filepath.Join("proc", strconv.Itoa(pid), "cpuset")
	data, err := ioutil.ReadFile(path)
	if err != nil {
		r.agent.logger.Info("error while reading ", path, ":", err.Error())
		return ""
	}

	return string(data)
}

func getRetryPeriodForIteration(retryNumber int) time.Duration {
	return time.Duration(math.Pow(2, float64(retryNumber-1))) * exponentialRetryPeriodBase
}
