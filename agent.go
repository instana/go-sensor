// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2016

package instana

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"
)

var payloadTooLargeErr = errors.New(`request payload is too large`)

const (
	agentDiscoveryURL = "/com.instana.plugin.golang.discovery"
	agentTracesURL    = "/com.instana.plugin.golang/traces."
	agentDataURL      = "/com.instana.plugin.golang."
	agentEventURL     = "/com.instana.plugin.generic.event"
	agentProfilesURL  = "/com.instana.plugin.golang/profiles."
	agentDefaultHost  = "localhost"
	agentDefaultPort  = 42699
	agentHeader       = "Instana Agent"

	// SnapshotPeriod is the amount of time in seconds between snapshot reports.
	SnapshotPeriod             = 600
	snapshotCollectionInterval = SnapshotPeriod * time.Second

	announceTimeout = 15 * time.Second
	clientTimeout   = 5 * time.Second

	maxContentLength      = 1024 * 1024 * 5
	numberOfBigSpansToLog = 5
)

// Shared data between agent and FSM (aka, stuff that FSM changes in the Agent):
// - agent.from
// - agent.host
// - sensor.options.Tracer.Secrets (comes from agentResonse.Secrets.Matcher)
// - sensor.options.Tracer.CollectableHTTPHeaders (comes from agentResponse.getExtraHTTPHeaders)
// *** sensor is global, so no need to pass it through functions ***
type agentHostData struct {
	// host is the agent host. It can be updated via default gateway or a new client announcement.
	host string

	// port id the agent port.
	port string

	// from is the agent information sent with each span in the "from" (span.f) section. it's format is as follows:
	// {e: "entityId", h: "hostAgentId", hl: trueIfServerlessPlatform, cp: "The cloud provider for a hostless span"}
	// Only span.f.e is mandatory.
	from *fromS
}

type agentResponse struct {
	Pid     uint32 `json:"pid"`
	HostID  string `json:"agentUuid"`
	Secrets struct {
		Matcher string   `json:"matcher"`
		List    []string `json:"list"`
	} `json:"secrets"`
	ExtraHTTPHeaders []string `json:"extraHeaders"`
	Tracing          struct {
		ExtraHTTPHeaders []string `json:"extra-http-headers"`
	} `json:"tracing"`
}

func (a *agentResponse) getExtraHTTPHeaders() []string {
	if len(a.Tracing.ExtraHTTPHeaders) == 0 {
		return a.ExtraHTTPHeaders
	}

	return a.Tracing.ExtraHTTPHeaders
}

type discoveryS struct {
	PID               int      `json:"pid"`
	Name              string   `json:"name"`
	Args              []string `json:"args"`
	Fd                string   `json:"fd"`
	Inode             string   `json:"inode"`
	CPUSetFileContent string   `json:"cpuSetFileContent"`
}

type fromS struct {
	EntityID string `json:"e"`
	// Serverless agents fields
	Hostless      bool   `json:"hl,omitempty"`
	CloudProvider string `json:"cp,omitempty"`
	// Host agent fields
	HostID string `json:"h,omitempty"`
}

func newServerlessAgentFromS(entityID, provider string) *fromS {
	return &fromS{
		EntityID:      entityID,
		Hostless:      true,
		CloudProvider: provider,
	}
}

func makeHostURL(agentData agentHostData, prefix string) string {
	url := "http://" + agentData.host + ":" + agentData.port + prefix

	if prefix[len(prefix)-1:] == "." && agentData.from.EntityID != "" {
		url += agentData.from.EntityID
	}

	return url
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type agentS struct {
	// agentData encapsulates info about the agent host and fromS. This is a shared information between the agent and
	// the fsm layer, so we use this wrapper to prevent passing data from one side to the other in a more sophisticated
	// way.
	agentData *agentHostData
	port      string

	mu  sync.RWMutex
	fsm *fsmS

	client          httpClient
	snapshot        *SnapshotCollector
	logger          LeveledLogger
	clientTimeout   time.Duration
	announceTimeout time.Duration

	printPayloadTooLargeErrInfoOnce sync.Once
}

func newAgent(serviceName, host string, port int, logger LeveledLogger) *agentS {
	if logger == nil {
		logger = defaultLogger
	}

	logger.Debug("initializing agent")

	agent := &agentS{
		agentData: &agentHostData{
			host: host,
			port: strconv.Itoa(port),
			from: &fromS{},
		},
		port:            strconv.Itoa(port),
		clientTimeout:   clientTimeout,
		announceTimeout: announceTimeout,
		client:          &http.Client{Timeout: announceTimeout},
		snapshot: &SnapshotCollector{
			CollectionInterval: snapshotCollectionInterval,
			ServiceName:        serviceName,
		},
		logger: logger,
	}

	agent.mu.Lock()
	agent.fsm = newFSM(agent.agentData, logger)
	agent.mu.Unlock()

	return agent
}

// Ready returns whether the agent has finished the announcement and is ready to send data
func (agent *agentS) Ready() bool {
	agent.mu.RLock()
	defer agent.mu.RUnlock()

	return agent.fsm.fsm.Current() == "ready"
}

// SendMetrics sends collected entity data to the host agent
func (agent *agentS) SendMetrics(data acceptor.Metrics) error {
	pid, err := strconv.Atoi(agent.agentData.from.EntityID)
	if err != nil && agent.agentData.from.EntityID != "" {
		agent.logger.Debug("agent got malformed PID %q", agent.agentData.from.EntityID)
	}

	if err = agent.request(makeHostURL(*agent.agentData, agentDataURL), "POST", acceptor.GoProcessData{
		PID:      pid,
		Snapshot: agent.snapshot.Collect(),
		Metrics:  data,
	}); err != nil {
		agent.logger.Error("failed to send metrics to the host agent: ", err)
		agent.reset()

		return err
	}

	return nil
}

// SendEvent sends an event using Instana Events API
func (agent *agentS) SendEvent(event *EventData) error {
	err := agent.request(makeHostURL(*agent.agentData, agentEventURL), "POST", event)
	if err != nil {
		// do not reset the agent as it might be not initialized at this state yet
		agent.logger.Warn("failed to send event ", event.Title, " to the host agent: ", err)

		return err
	}

	return nil
}

// SendSpans sends collected spans to the host agent
func (agent *agentS) SendSpans(spans []Span) error {
	for i := range spans {
		spans[i].From = agent.agentData.from
	}

	err := agent.request(makeHostURL(*agent.agentData, agentTracesURL), "POST", spans)
	if err != nil {
		if err == payloadTooLargeErr {
			agent.printPayloadTooLargeErrInfoOnce.Do(
				func() {
					agent.logDetailedInformationAboutDroppedSpans(numberOfBigSpansToLog, spans, err)
				},
			)

			return nil
		} else {
			agent.logger.Error("failed to send spans to the host agent: ", err)
			agent.reset()
		}

		return err
	}

	return nil
}

// Flush is a noop for host agent
func (agent *agentS) Flush(ctx context.Context) error { return nil }

type hostAgentProfile struct {
	autoprofile.Profile
	ProcessID string `json:"pid"`
}

// SendProfiles sends profile data to the agent
func (agent *agentS) SendProfiles(profiles []autoprofile.Profile) error {
	agentProfiles := make([]hostAgentProfile, 0, len(profiles))
	for _, p := range profiles {
		agentProfiles = append(agentProfiles, hostAgentProfile{p, agent.agentData.from.EntityID})
	}

	err := agent.request(makeHostURL(*agent.agentData, agentProfilesURL), "POST", agentProfiles)
	if err != nil {
		agent.logger.Error("failed to send profile data to the host agent: ", err)
		agent.reset()

		return err
	}

	return nil
}

func (agent *agentS) setLogger(l LeveledLogger) {
	agent.logger = l
}

// All usages are now with POST method
// request will overwrite the client timeout for a single request
func (agent *agentS) request(url string, method string, data interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), agent.clientTimeout)
	defer cancel()

	var r io.Reader

	if data != nil {
		b, err := json.Marshal(data)

		if err != nil {
			return err
		}

		r = bytes.NewBuffer(b)
	}

	req, err := http.NewRequest(method, url, r)

	if err != nil {
		return err
	}

	req = req.WithContext(ctx)

	_, err = agent.client.Do(req)

	return err
}

func (agent *agentS) reset() {
	agent.mu.Lock()
	agent.fsm.reset()
	agent.mu.Unlock()
}

func (agent *agentS) logDetailedInformationAboutDroppedSpans(size int, spans []Span, err error) {
	var marshaledSpans []string
	for i := range spans {
		ms, err := json.Marshal(spans[i])
		if err == nil {
			marshaledSpans = append(marshaledSpans, string(ms))
		}
	}
	sort.Slice(marshaledSpans, func(i, j int) bool {
		// descending order
		return len(marshaledSpans[i]) > len(marshaledSpans[j])
	})

	if size > len(marshaledSpans) {
		size = len(marshaledSpans)
	}

	agent.logger.Warn(
		fmt.Sprintf("failed to send spans to the host agent: dropped %d span(s) : %s.\nThis detailed information will only be logged once.\nSpans :\n %s",
			len(spans),
			err.Error(),
			strings.Join(marshaledSpans[:size], ";"),
		),
	)
}
