package instana

import (
	"os"
	"os/exec"
	"strconv"
	"time"

	f "github.com/looplab/fsm"
)

const (
	E_INIT     = "init"
	E_LOOKUP   = "lookup"
	E_ANNOUNCE = "announce"
	E_TEST     = "test"

	RETRY_PERIOD = 30 * 1000
)

type fsmS struct {
	agent *agentS
	fsm   *f.FSM
	timer *time.Timer
}

func (r *fsmS) init() {

	log.debug("initializing fsm")

	r.fsm = f.NewFSM(
		"none",
		f.Events{
			{Name: E_INIT, Src: []string{"none", "unannounced", "announced", "ready"}, Dst: "init"},
			{Name: E_LOOKUP, Src: []string{"init"}, Dst: "unannounced"},
			{Name: E_ANNOUNCE, Src: []string{"unannounced"}, Dst: "announced"},
			{Name: E_TEST, Src: []string{"announced"}, Dst: "ready"}},
		f.Callbacks{
			"init":              r.lookupAgentHost,
			"enter_unannounced": r.announceSensor,
			"enter_announced":   r.testAgent})
}

func (r *fsmS) scheduleRetry(e *f.Event, cb func(e *f.Event)) {
	r.timer = time.NewTimer(RETRY_PERIOD * time.Millisecond)
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
				go r.checkHost(r.getDefaultGateway(), func(b bool, host string) {
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
		go r.checkHost(AGENT_DEFAULT_HOST, cb)
	}
}

func (r *fsmS) getDefaultGateway() string {
	out, _ := exec.Command("/sbin/ip route | awk '/default/ { print $3 }'").Output()

	log.debug("checking default gateway", string(out))

	return string(out)
}

func (r *fsmS) checkHost(host string, cb func(b bool, host string)) {
	log.debug("checking host", host)

	header, err := r.agent.requestHeader(r.agent.makeHostUrl(host, "/"), "GET", "Server")

	cb(err == nil && header == AGENT_HEADER, host)
}

func (r *fsmS) lookupSuccess(host string) {
	log.debug("agent lookup success", host)

	r.agent.setHost(host)
	r.fsm.Event(E_LOOKUP)
}

func (r *fsmS) announceSensor(e *f.Event) {
	cb := func(b bool, from *FromS) {
		if b {
			r.agent.setFrom(from)
			r.fsm.Event(E_ANNOUNCE)
		} else {
			log.error("Cannot announce sensor. Scheduling retry.")
			r.scheduleRetry(e, r.announceSensor)
		}
	}

	log.debug("announcing sensor to the agent")

	go func(cb func(b bool, from *FromS)) {
		d := &Discovery{
			Pid:  os.Getpid(),
			Name: os.Args[0],
			Args: os.Args[1:]}

		ret := &agentResponse{}
		_, err := r.agent.requestResponse(r.agent.makeUrl(AGENT_DISCOVERY_URL), "PUT", d, ret)
		cb(err == nil,
			&FromS{
				Pid:    strconv.Itoa(int(ret.Pid)),
				HostId: ret.HostId})
	}(cb)
}

func (r *fsmS) testAgent(e *f.Event) {
	cb := func(b bool) {
		if b {
			r.fsm.Event(E_TEST)
		} else {
			log.error("Agent is not yet ready. Scheduling retry.")
			r.scheduleRetry(e, r.testAgent)
		}
	}

	log.debug("testing communication with the agent")

	go func(cb func(b bool)) {
		_, err := r.agent.head(r.agent.makeUrl(AGENT_DATA_URL))
		cb(err == nil)
	}(cb)
}

func (r *fsmS) reset() {
	r.fsm.Event(E_INIT)
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
