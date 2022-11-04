// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2016

package instana

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net"
	"net/http"
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
	agentData                  *agentHostData
	fsm                        *f.FSM
	timer                      *time.Timer
	retriesLeft                int
	expDelayFunc               func(retryNumber int) time.Duration
	lookupAgentHostRetryPeriod time.Duration
	logger                     LeveledLogger
	agentPort                  string
}

func newHostAgentFromS(pid int, hostID string) *fromS {
	return &fromS{
		EntityID: strconv.Itoa(pid),
		HostID:   hostID,
	}
}

func newFSM(ahd *agentHostData, logger LeveledLogger, port string) *fsmS {
	logger.Warn("Stan is on the scene. Starting Instana instrumentation.")
	logger.Debug("initializing fsm")

	ret := &fsmS{
		agentData:                  ahd,
		retriesLeft:                maximumRetries,
		expDelayFunc:               expDelay,
		logger:                     logger,
		lookupAgentHostRetryPeriod: retryPeriod,
		agentPort:                  port,
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
	go r.checkHost(e, r.agentData.host)
}

func (r *fsmS) checkHost(e *f.Event, host string) {
	r.logger.Debug("checking host ", r.agentData.host)
	url := "http://" + r.agentData.host + ":" + r.agentPort + "/"

	resp, err := http.Get(url)

	var header string

	if err == nil {
		header = resp.Header.Get("Server")
	}

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

		url := "http://" + r.agentData.host + ":" + r.agentPort + "/"

		resp, err := http.Get(url)

		var header string

		if err == nil {
			header = resp.Header.Get("Server")
		}

		found := err == nil && header == agentHeader

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

	r.agentData.host = host
	r.retriesLeft = maximumRetries
	r.fsm.Event(eLookup)
}

func (r *fsmS) handleRetries(e *f.Event) {
	r.retriesLeft--
	if r.retriesLeft == 0 {
		r.logger.Error("Couldn't announce the sensor after reaching the maximum amount of attempts.")
		r.fsm.Event(eInit)
		return
	}

	r.logger.Debug("Cannot announce sensor. Scheduling retry.")

	retryNumber := maximumRetries - r.retriesLeft + 1
	r.scheduleRetryWithExponentialDelay(e, r.announceSensor, retryNumber)
}

func (r *fsmS) applyHostAgentSettings(resp agentResponse) {
	r.agentData.from = newHostAgentFromS(int(resp.Pid), resp.HostID)

	if resp.Secrets.Matcher != "" {
		m, err := NamedMatcher(resp.Secrets.Matcher, resp.Secrets.List)
		if err != nil {
			r.logger.Warn("failed to apply secrets matcher configuration: ", err)
		} else {
			sensor.options.Tracer.Secrets = m
		}
	}

	if len(sensor.options.Tracer.CollectableHTTPHeaders) == 0 {
		sensor.options.Tracer.CollectableHTTPHeaders = resp.getExtraHTTPHeaders()
	}
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

		jsonData, _ := json.Marshal(d)

		var resp agentResponse

		client := http.DefaultClient

		req, err := http.NewRequest(http.MethodPut, "http://"+r.agentData.host+":"+r.agentPort+agentDiscoveryURL, bytes.NewBuffer(jsonData))

		if err != nil {
			r.handleRetries(e)
			return
		}

		res, err := client.Do(req)

		badResponse := res != nil && (res.StatusCode < 200 || res.StatusCode >= 300)

		if err != nil || badResponse {
			r.handleRetries(e)
			return
		}

		respBytes, err := ioutil.ReadAll(res.Body)
		defer res.Body.Close()

		if err != nil {
			r.handleRetries(e)
			return
		}

		json.Unmarshal(respBytes, &resp)

		r.logger.Info("Host agent available. We're in business. Announced pid:", resp.Pid)

		r.applyHostAgentSettings(resp)

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
		if addr, err := net.ResolveTCPAddr("tcp", r.agentData.host+":42699"); err == nil {
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
		// TODO: url is missing pid at the end
		u := "http://" + r.agentData.host + ":" + r.agentPort + agentDataURL

		resp, err := http.Head(u)

		badResponse := resp != nil && (resp.StatusCode < 200 || resp.StatusCode >= 300)

		// TODO: to put this piece of code i na function
		if err != nil || badResponse {
			r.logger.Debug("Agent is not yet ready. Scheduling retry.")
			r.retriesLeft--
			if r.retriesLeft > 0 {
				retryNumber := maximumRetries - r.retriesLeft + 1
				r.scheduleRetryWithExponentialDelay(e, r.testAgent, retryNumber)
			} else {
				r.fsm.Event(eInit)
			}
		} else {
			r.retriesLeft = maximumRetries
			r.fsm.Event(eTest)
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
