package instana

import (
	"bytes"
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
	Pid    uint32 `json:"pid"`
	HostID string `json:"agentUuid"`
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

type agentS struct {
	from *fromS
	host string
	port string

	fsm      *fsmS
	client   *http.Client
	snapshot *SnapshotCollector
	logger   LeveledLogger
}

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

type hostAgentSpan struct {
	Span
	From *fromS `json:"f"` // override the `f` fields with agent-specific type
}

// SendSpans sends collected spans to the host agent
func (agent *agentS) SendSpans(spans []Span) error {
	agentSpans := make([]hostAgentSpan, 0, len(spans))
	for _, sp := range spans {
		agentSpans = append(agentSpans, hostAgentSpan{sp, agent.from})
	}

	_, err := agent.request(agent.makeURL(agentTracesURL), "POST", agentSpans)
	if err != nil {
		agent.logger.Error("failed to send spans to the host agent: ", err)
		agent.reset()

		return err
	}

	return nil
}

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

func (r *agentS) setLogger(l LeveledLogger) {
	r.logger = l
}

func (r *agentS) makeURL(prefix string) string {
	return r.makeHostURL(r.host, prefix)
}

func (r *agentS) makeHostURL(host string, prefix string) string {
	var buffer bytes.Buffer

	buffer.WriteString("http://")
	buffer.WriteString(host)
	buffer.WriteString(":")
	buffer.WriteString(r.port)
	buffer.WriteString(prefix)
	if prefix[len(prefix)-1:] == "." && r.from.EntityID != "" {
		buffer.WriteString(r.from.EntityID)
	}

	return buffer.String()
}

func (r *agentS) head(url string) (string, error) {
	return r.request(url, "HEAD", nil)
}

func (r *agentS) request(url string, method string, data interface{}) (string, error) {
	return r.fullRequestResponse(url, method, data, nil, "")
}

func (r *agentS) requestResponse(url string, method string, data interface{}, ret interface{}) (string, error) {
	return r.fullRequestResponse(url, method, data, ret, "")
}

func (r *agentS) requestHeader(url string, method string, header string) (string, error) {
	return r.fullRequestResponse(url, method, nil, nil, header)
}

func (r *agentS) fullRequestResponse(url string, method string, data interface{}, body interface{}, header string) (string, error) {
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
			resp, err = r.client.Do(req)
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
		if !r.fsm.fsm.Is("announced") {
			r.logger.Info(err, url)
		}
	}

	return ret, err
}

func (r *agentS) setFrom(from *fromS) {
	r.from = from
}

func (r *agentS) setHost(host string) {
	r.host = host
}

func (r *agentS) reset() {
	r.fsm.reset()
}
