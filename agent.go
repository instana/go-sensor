package instana

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"
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
	PID    string `json:"e"`
	HostID string `json:"h"`
}

type agentS struct {
	ServiceName string
	from        *fromS
	host        string
	port        string

	snapshotMu                 sync.RWMutex
	lastSnapshotCollectionTime time.Time

	fsm    *fsmS
	client *http.Client
	logger LeveledLogger
}

func newAgent(serviceName, host string, port int, logger LeveledLogger) *agentS {
	if logger == nil {
		logger = defaultLogger
	}

	logger.Debug("initializing agent")

	agent := &agentS{
		ServiceName: serviceName,
		from:        &fromS{},
		host:        host,
		port:        strconv.Itoa(port),
		client:      &http.Client{Timeout: 5 * time.Second},
		logger:      logger,
	}
	agent.fsm = newFSM(agent)

	return agent
}

// Ready returns whether the agent has finished the announcement and is ready to send data
func (agent *agentS) Ready() bool {
	return agent.fsm.fsm.Current() == "ready"
}

// SendMetrics sends collected entity data to the host agent
func (agent *agentS) SendMetrics(data *MetricsS) {
	pid, err := strconv.Atoi(agent.from.PID)
	if err != nil && agent.from.PID != "" {
		agent.logger.Debug("agent got malformed PID %q", agent.from.PID)
	}

	if _, err = agent.request(agent.makeURL(agentDataURL), "POST", &EntityData{
		PID:      pid,
		Snapshot: agent.collectSnapshot(),
		Metrics:  data,
	}); err != nil {
		agent.reset()
	}
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
	if prefix[len(prefix)-1:] == "." && r.from.PID != "" {
		buffer.WriteString(r.from.PID)
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

func (agent *agentS) collectSnapshot() *SnapshotS {
	agent.snapshotMu.RLock()
	lastSnapshotCollectionTime := agent.lastSnapshotCollectionTime
	agent.snapshotMu.RUnlock()

	if time.Since(lastSnapshotCollectionTime) < snapshotCollectionInterval {
		return nil
	}

	agent.snapshotMu.Lock()
	defer agent.snapshotMu.Unlock()

	agent.lastSnapshotCollectionTime = time.Now()
	agent.logger.Debug("collected snapshot")

	return &SnapshotS{
		Name:     agent.ServiceName,
		Version:  runtime.Version(),
		Root:     runtime.GOROOT(),
		MaxProcs: runtime.GOMAXPROCS(0),
		Compiler: runtime.Compiler,
		NumCPU:   runtime.NumCPU(),
	}
}
