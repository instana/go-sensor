package instana

import (
	"os"
	"os/exec"
	"strconv"
	"time"

	f "github.com/looplab/fsm"
)

const (
	EInit     = "init"
	ELookup   = "lookup"
	EAnnounce = "announce"
	ETest     = "test"

	RetryPeriod    = 30 * 1000
	MaximumRetries = 2
)

type fsmS struct {
	agent   *agentS
	fsm     *f.FSM
	timer   *time.Timer
	retries int
}

func (r *fsmS) init() {

	log.debug("initializing fsm")

	r.fsm = f.NewFSM(
		"none",
		f.Events{
			{Name: EInit, Src: []string{"none", "unannounced", "announced", "ready"}, Dst: "init"},
			{Name: ELookup, Src: []string{"init"}, Dst: "unannounced"},
			{Name: EAnnounce, Src: []string{"unannounced"}, Dst: "announced"},
			{Name: ETest, Src: []string{"announced"}, Dst: "ready"}},
		f.Callbacks{
			"init":              r.lookupAgentHost,
			"enter_unannounced": r.announceSensor,
			"enter_announced":   r.testAgent})

	r.retries = MaximumRetries
	r.fsm.Event(EInit)
}

func (r *fsmS) scheduleRetry(e *f.Event, cb func(e *f.Event)) {
	r.timer = time.NewTimer(RetryPeriod * time.Millisecond)
	go func() {
		<-r.timer.C
		cb(e)
	}()
}

func (r *fsmS) lookupAgentHost(e *f.Event) {
	cb := func(b bool, host string) {
		if b {
			r.lookupSuccess(host)
		} else {
			gateway := r.getDefaultGateway()
			if gateway != "" {
				go r.checkHost(gateway, func(b bool, host string) {
					if b {
						r.lookupSuccess(host)
					} else {
						log.error("Cannot connect to the agent through localhost or default gateway. Scheduling retry.")
						r.scheduleRetry(e, r.lookupAgentHost)
					}
				})
			} else {
				log.error("Default gateway not available. Scheduling retry")
				r.scheduleRetry(e, r.lookupAgentHost)
			}
		}
	}

	if r.agent.sensor.options.AgentHost != "" {
		go r.checkHost(r.agent.sensor.options.AgentHost, cb)
	} else {
		go r.checkHost(AgentDefaultHost, cb)
	}
}

func (r *fsmS) getDefaultGateway() string {
	out, _ := exec.Command("/bin/sh", "-c", "/sbin/ip route | awk '/default/' | cut -d ' ' -f 3 | tr -d '\n'").Output()

	log.debug("checking default gateway", string(out[:]))

	return string(out[:])
}

func (r *fsmS) checkHost(host string, cb func(b bool, host string)) {
	log.debug("checking host", host)

	header, err := r.agent.requestHeader(r.agent.makeHostURL(host, "/"), "GET", "Server")

	cb(err == nil && header == AgentHeader, host)
}

func (r *fsmS) lookupSuccess(host string) {
	log.debug("agent lookup success", host)

	r.agent.setHost(host)
	r.retries = MaximumRetries
	r.fsm.Event(ELookup)
}

func (r *fsmS) announceSensor(e *f.Event) {
	cb := func(b bool, from *FromS) {
		if b {
			r.agent.setFrom(from)
			r.retries = MaximumRetries
			r.fsm.Event(EAnnounce)
		} else {
			log.error("Cannot announce sensor. Scheduling retry.")
			r.retries--
			if r.retries > 0 {
				r.scheduleRetry(e, r.announceSensor)
			} else {
				r.fsm.Event(EInit)
			}
		}
	}

	log.debug("announcing sensor to the agent")

	go func(cb func(b bool, from *FromS)) {
		d := &Discovery{
			PID:  os.Getpid(),
			Name: os.Args[0],
			Args: os.Args[1:]}

		ret := &agentResponse{}
		_, err := r.agent.requestResponse(r.agent.makeURL(AgentDiscoveryURL), "PUT", d, ret)
		cb(err == nil,
			&FromS{
				PID:    strconv.Itoa(int(ret.Pid)),
				HostID: ret.HostID})
	}(cb)
}

func (r *fsmS) testAgent(e *f.Event) {
	cb := func(b bool) {
		if b {
			r.retries = MaximumRetries
			r.fsm.Event(ETest)
		} else {
			log.error("Agent is not yet ready. Scheduling retry.")
			r.retries--
			if r.retries > 0 {
				r.scheduleRetry(e, r.testAgent)
			} else {
				r.fsm.Event(EInit)
			}
		}
	}

	log.debug("testing communication with the agent")

	go func(cb func(b bool)) {
		_, err := r.agent.head(r.agent.makeURL(AgentDataURL))
		cb(err == nil)
	}(cb)
}

func (r *fsmS) reset() {
	r.retries = MaximumRetries
	r.fsm.Event(EInit)
}

func (r *agentS) initFsm() *fsmS {
	ret := new(fsmS)
	ret.agent = r
	ret.init()

	return ret
}

func (r *agentS) canSend() bool {
	return r.fsm.fsm.Current() == "ready"
}
