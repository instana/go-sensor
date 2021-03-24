// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2016

package instana

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"
)

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
)

type agentResponse struct {
	Pid     uint32 `json:"pid"`
	HostID  string `json:"agentUuid"`
	Secrets struct {
		Matcher string   `json:"matcher"`
		List    []string `json:"list"`
	} `json:"secrets"`
	ExtraHTTPHeaders []string `json:"extraHeaders"`
}

type discoveryS struct {
	PID   int      `json:"pid"`
	Name  string   `json:"name"`
	Args  []string `json:"args"`
	Fd    string   `json:"fd"`
	Inode string   `json:"inode"`
}

type fromS struct {
	EntityID string `json:"e"`
	// Serverless agents fields
	Hostless      bool   `json:"hl,omitempty"`
	CloudProvider string `json:"cp,omitempty"`
	// Host agent fields
	HostID string `json:"h,omitempty"`
}

func newHostAgentFromS(pid int, hostID string) *fromS {
	return &fromS{
		EntityID: strconv.Itoa(pid),
		HostID:   hostID,
	}
}

func newServerlessAgentFromS(entityID, provider string) *fromS {
	return &fromS{
		EntityID:      entityID,
		Hostless:      true,
		CloudProvider: provider,
	}
}

type agentS struct {
	from *fromS
	host string
	port string

	fsm      *fsmS
	client   *http.Client
	snapshot *SnapshotCollector
	logger   LeveledLogger
}

//go:norace
func newAgent(serviceName, host string, port int, logger LeveledLogger) *agentS {
	if logger == nil {
		logger = defaultLogger
	}

	logger.Debug("initializing agent")

	agent := &agentS{
		from:   &fromS{},
		host:   host,
		port:   strconv.Itoa(port),
		client: &http.Client{Timeout: 5 * time.Second},
		snapshot: &SnapshotCollector{
			CollectionInterval: snapshotCollectionInterval,
			ServiceName:        serviceName,
		},
		logger: logger,
	}
	agent.fsm = newFSM(agent)

	return agent
}

// Ready returns whether the agent has finished the announcement and is ready to send data
func (agent *agentS) Ready() bool {
	return agent.fsm.fsm.Current() == "ready"
}

// SendMetrics sends collected entity data to the host agent
func (agent *agentS) SendMetrics(data acceptor.Metrics) error {
	pid, err := strconv.Atoi(agent.from.EntityID)
	if err != nil && agent.from.EntityID != "" {
		agent.logger.Debug("agent got malformed PID %q", agent.from.EntityID)
	}

	if _, err = agent.request(agent.makeURL(agentDataURL), "POST", acceptor.GoProcessData{
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
	_, err := agent.request(agent.makeURL(agentEventURL), "POST", event)
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
		spans[i].From = agent.from
	}

	_, err := agent.request(agent.makeURL(agentTracesURL), "POST", spans)
	if err != nil {
		agent.logger.Error("failed to send spans to the host agent: ", err)
		agent.reset()

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
		agentProfiles = append(agentProfiles, hostAgentProfile{p, agent.from.EntityID})
	}

	_, err := agent.request(agent.makeURL(agentProfilesURL), "POST", agentProfiles)
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

func (agent *agentS) makeURL(prefix string) string {
	return agent.makeHostURL(agent.host, prefix)
}

func (agent *agentS) makeHostURL(host string, prefix string) string {
	var buffer bytes.Buffer

	buffer.WriteString("http://")
	buffer.WriteString(host)
	buffer.WriteString(":")
	buffer.WriteString(agent.port)
	buffer.WriteString(prefix)
	if prefix[len(prefix)-1:] == "." && agent.from.EntityID != "" {
		buffer.WriteString(agent.from.EntityID)
	}

	return buffer.String()
}

func (agent *agentS) head(url string) (string, error) {
	return agent.request(url, "HEAD", nil)
}

func (agent *agentS) request(url string, method string, data interface{}) (string, error) {
	return agent.fullRequestResponse(url, method, data, nil, "")
}

func (agent *agentS) requestResponse(url string, method string, data interface{}, ret interface{}) (string, error) {
	return agent.fullRequestResponse(url, method, data, ret, "")
}

func (agent *agentS) requestHeader(url string, method string, header string) (string, error) {
	return agent.fullRequestResponse(url, method, nil, nil, header)
}

func (agent *agentS) fullRequestResponse(url string, method string, data interface{}, body interface{}, header string) (string, error) {
	var j []byte
	var ret string
	var err error
	var resp *http.Response
	var req *http.Request

	if data != nil {
		j, err = json.Marshal(data)
	}

	if err == nil {
		if j != nil {
			req, err = http.NewRequest(method, url, bytes.NewBuffer(j))
		} else {
			req, err = http.NewRequest(method, url, nil)
		}

		// Uncomment this to dump json payloads
		// log.debug(bytes.NewBuffer(j))

		if err == nil {
			req.Header.Set("Content-Type", "application/json")
			resp, err = agent.client.Do(req)
			if err == nil {
				defer resp.Body.Close()

				if resp.StatusCode < 200 || resp.StatusCode >= 300 {
					err = errors.New(resp.Status)
				} else {
					if body != nil {
						var b []byte
						b, err = ioutil.ReadAll(resp.Body)
						json.Unmarshal(b, body)
					}

					if header != "" {
						ret = resp.Header.Get(header)
					}
				}
			}
		}
	}

	if err != nil {
		// Ignore errors while in announced stated (before ready) as
		// this is the time where the entity is registering in the Instana
		// backend and it will return 404 until it's done.
		if !agent.fsm.fsm.Is("announced") {
			agent.logger.Info("failed to send a request to ", url, ": ", err)
		}
	}

	return ret, err
}

func (agent *agentS) applyHostAgentSettings(resp agentResponse) {
	agent.from = newHostAgentFromS(int(resp.Pid), resp.HostID)

	if resp.Secrets.Matcher != "" {
		m, err := NamedMatcher(resp.Secrets.Matcher, resp.Secrets.List)
		if err != nil {
			agent.logger.Warn("failed to apply secrets matcher configuration: ", err)
		} else {
			sensor.options.Tracer.Secrets = m
		}
	}

	sensor.options.Tracer.CollectableHTTPHeaders = resp.ExtraHTTPHeaders
}

func (agent *agentS) setHost(host string) {
	agent.host = host
}

func (agent *agentS) reset() {
	agent.fsm.reset()
}
