// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2016

package instana

import (
	"context"
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
	agentComm                  *agentCommunicator
	fsm                        *f.FSM
	timer                      *time.Timer
	retriesLeft                int
	expDelayFunc               func(retryNumber int) time.Duration
	lookupAgentHostRetryPeriod time.Duration
	logger                     LeveledLogger
}

func newHostAgentFromS(pid int, hostID string) *fromS {
	return &fromS{
		EntityID: strconv.Itoa(pid),
		HostID:   hostID,
	}
}

func newFSM(ahd *agentCommunicator, logger LeveledLogger) *fsmS {
	logger.Warn("Stan is on the scene. Starting Instana instrumentation.")
	logger.Debug("initializing fsm")

	ret := &fsmS{
		agentComm:                  ahd,
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
			"ready":             ret.ready,
		})
	ret.fsm.Event(context.Background(), eInit)

	return ret
}

func (r *fsmS) scheduleRetry(e *f.Event, cb func(_ context.Context, e *f.Event)) {
	r.timer = time.NewTimer(r.lookupAgentHostRetryPeriod)
	go func() {
		<-r.timer.C
		cb(context.Background(), e)
	}()
}

func (r *fsmS) scheduleRetryWithExponentialDelay(e *f.Event, cb func(_ context.Context, e *f.Event), retryNumber int) {
	time.Sleep(r.expDelayFunc(retryNumber))
	cb(context.Background(), e)
}

func (r *fsmS) lookupAgentHost(_ context.Context, e *f.Event) {
	go r.checkHost(e)
}

// checkHost verifies and set the agent host address
func (r *fsmS) checkHost(e *f.Event) {

	// Look for a successful ping from the configured host
	host := r.agentComm.host
	r.logger.Debug("checking host ", r.agentComm.host)

	found := r.agentComm.checkForSuccessResponse()

	if found {
		r.lookupSuccess(host)
		r.logger.Debug("Agent host found: '", host, "' when attempting to read the string 'Instana Agent' from the response header.")
		return
	}

	// Check whether agent host is configured in env variable and look for a successful ping from the configured host
	r.logger.Debug("Attempting to retrieve host from the INSTANA_AGENT_HOST environment variable")
	hostFromEnv, ok := os.LookupEnv("INSTANA_AGENT_HOST")

	if !ok {
		r.logger.Debug("No INSTANA_AGENT_HOST environment variable present")
	} else {
		r.logger.Debug("Attempting to reach the agent with host found from the INSTANA_AGENT_HOST environment variable: ", hostFromEnv)
		originalHost := r.agentComm.host
		r.agentComm.host = hostFromEnv
		found = r.agentComm.checkForSuccessResponse()

		if found {
			r.logger.Debug("Lookup successful with host from the INSTANA_AGENT_HOST environment variable: ", hostFromEnv)
			r.lookupSuccess(hostFromEnv)
			return
		}

		r.logger.Debug("Lookup failed with host from the INSTANA_AGENT_HOST environment variable: ", hostFromEnv, ". Updating host back to the original: ", originalHost)

		r.agentComm.host = originalHost
	}

	// Look for a successful ping for the configured default gateway
	routeFilename := "/proc/net/route"
	r.logger.Debug("Lookup failed for expected host: ", r.agentComm.host, ". Will attempt to read host from ", routeFilename)
	if _, fileNotFoundErr := os.Stat(routeFilename); fileNotFoundErr == nil {
		gateway, err := getDefaultGateway(routeFilename)
		r.logger.Debug("Identified the gateway: ", gateway)
		if err != nil {
			// This will be always the "failed to open /proc/net/route: no such file or directory" error.
			// As this info is not relevant to the customer, we can remove it from the message.
			r.logger.Error("Couldn't open the ", routeFilename, " file in order to retrieve the default gateway. Scheduling retry.")
			r.scheduleRetry(e, r.lookupAgentHost)

			return
		}

		if gateway == "" {
			r.logger.Error("Couldn't parse the default gateway address from ", routeFilename, ". Scheduling retry.")
			r.scheduleRetry(e, r.lookupAgentHost)

			return
		}

		originalHost := r.agentComm.host
		r.agentComm.host = gateway
		found := r.agentComm.checkForSuccessResponse()

		if found {
			r.logger.Debug("Lookup successful with host from ", routeFilename, ": ", gateway)
			r.lookupSuccess(gateway)
			return
		}

		r.logger.Debug("Lookup failed with host from ", routeFilename, ": ", gateway, ". Updating host back to the original: ", originalHost)

		r.agentComm.host = originalHost

		r.logger.Error("Cannot connect to the agent through default gateway. Scheduling retry.")
		r.scheduleRetry(e, r.lookupAgentHost)
	} else {
		r.logger.Error("Cannot connect to the agent. Scheduling retry.")
		r.logger.Debug("Connecting through the default gateway has not been attempted because ", routeFilename, " does not exist.")
		r.scheduleRetry(e, r.lookupAgentHost)
	}
}

func (r *fsmS) lookupSuccess(host string) {
	r.logger.Debug("agent lookup success ", host)

	r.agentComm.host = host
	r.retriesLeft = maximumRetries
	r.fsm.Event(context.Background(), eLookup)
}

func (r *fsmS) handleRetries(e *f.Event, cb func(_ context.Context, e *f.Event), retryFailMsg, retryMsg string) {
	r.retriesLeft--
	if r.retriesLeft == 0 {
		r.logger.Error(retryFailMsg)
		r.fsm.Event(context.Background(), eInit)
		return
	}

	r.logger.Debug(retryMsg)
	retryNumber := maximumRetries - r.retriesLeft + 1
	r.scheduleRetryWithExponentialDelay(e, cb, retryNumber)
}

func (r *fsmS) applyHostAgentSettings(resp agentResponse) {
	r.agentComm.from = newHostAgentFromS(int(resp.Pid), resp.HostID)

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

func (r *fsmS) announceSensor(_ context.Context, e *f.Event) {
	r.logger.Debug("announcing sensor to the agent")

	go func() {
		defer func() {
			if err := recover(); err != nil {
				r.logger.Debug("Announce recovered:", err)
			}
		}()

		retryFailedMsg := "announceSensor: Couldn't announce the sensor after reaching the maximum amount of attempts."
		retryMsg := "Cannot announce sensor. Scheduling retry."

		d := r.getDiscoveryS()

		resp := r.agentComm.agentResponse(d)

		if resp == nil {
			r.handleRetries(e, r.announceSensor, retryFailedMsg, retryMsg)
			return
		}

		r.logger.Info("Host agent available. We're in business. Announced pid:", resp.Pid)

		r.applyHostAgentSettings(*resp)

		r.retriesLeft = maximumRetries
		r.fsm.Event(context.Background(), eAnnounce)
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
		if addr, err := net.ResolveTCPAddr("tcp", r.agentComm.host+":42699"); err == nil {
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

func (r *fsmS) testAgent(_ context.Context, e *f.Event) {
	r.logger.Debug("testing communication with the agent")
	go func() {
		if !r.agentComm.pingAgent() {
			r.handleRetries(e, r.testAgent, "testAgent: Couldn't announce the sensor after reaching the maximum amount of attempts.", "Agent is not yet ready. Scheduling retry.")
			return
		}

		r.retriesLeft = maximumRetries
		r.fsm.Event(context.Background(), eTest)
	}()
}

func (r *fsmS) reset() {
	r.retriesLeft = maximumRetries
	r.fsm.Event(context.Background(), eInit)
}

func (r *fsmS) ready(_ context.Context, e *f.Event) {
	go delayed.flush()
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
