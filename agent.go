package instana

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

const (
	AGENT_DISCOVERY_URL = "/com.instana.plugin.golang.discovery"
	AGENT_TRACES_URL    = "/com.instana.plugin.golang/traces."
	AGENT_DATA_URL      = "/com.instana.plugin.golang."
	AGENT_DEFAULT_HOST  = "localhost"
	AGENT_DEFAULT_PORT  = 42699
	AGENT_HEADER        = "Instana Agent"
)

type agentResponse struct {
	Pid    uint32 `json:"pid"`
	HostId string `json:"agentUuid"`
}

type Discovery struct {
	Pid  int      `json:"pid"`
	Name string   `json:"name"`
	Args []string `json:"args"`
}

type FromS struct {
	Pid    string `json:"e"`
	HostId string `json:"h"`
}

type agentI interface {
	makeUrl(prefix string) string
	makeHostUrl(host string, prefix string) string
	makeFullUrl(host string, port int, prefix string) string
	head(url string) (string, error)
	request(url string, method string, data interface{}) (string, error)
	requestResponse(url string, method string, data interface{}, ret interface{}) (string, error)
	requestHeader(url string, method string, header string) (string, error)
	fullRequestResponse(url string, method string, data interface{}, body interface{}, header string) (string, error)
	reset()
	setFrom(from *FromS)
	setHost(host string)
	initFsm() *fsmS
	canSend() bool
}

type agentS struct {
	sensor *sensorS
	fsm    *fsmS
	from   *FromS
	host   string
}

func (r *agentS) init() {
	r.fsm = r.initFsm()
}

func (r *agentS) makeUrl(prefix string) string {
	return r.makeHostUrl(r.host, prefix)
}

func (r *agentS) makeHostUrl(host string, prefix string) string {
	var port int
	if r.sensor.options.AgentPort == 0 {
		port = AGENT_DEFAULT_PORT
	} else {
		port = r.sensor.options.AgentPort
	}

	return r.makeFullUrl(host, port, prefix)
}

func (r *agentS) makeFullUrl(host string, port int, prefix string) string {
	var buffer bytes.Buffer

	buffer.WriteString("http://")
	buffer.WriteString(host)
	buffer.WriteString(":")
	buffer.WriteString(strconv.Itoa(port))
	buffer.WriteString(prefix)
	if r.from.Pid != "" {
		buffer.WriteString(r.from.Pid)
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
	j, err = json.Marshal(data)
	if err == nil {
		req, err = http.NewRequest(method, url, bytes.NewBuffer(j))
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err = client.Do(req)
		if err == nil {
			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				err = errors.New(resp.Status)
				log.error(err)
				if r.canSend() {
					r.reset()
				}
			} else {
				defer resp.Body.Close()

				log.debug("agent response:", url, resp.Status)

				if body != nil {
					var b []byte
					b, err = ioutil.ReadAll(resp.Body)
					json.Unmarshal(b, body)
				}

				if header != "" {
					ret = resp.Header.Get(header)
				}
			}
		} else {
			log.error(err)

			if resp == nil {
				r.reset()
			}
		}
	} else {
		log.error(err)
	}

	return ret, err
}

func (r *agentS) reset() {
	r.setFrom(&FromS{})
	r.fsm.reset()
}

func (r *agentS) setFrom(from *FromS) {
	r.from = from
}

func (r *agentS) setHost(host string) {
	r.host = host
}

func (r *sensorS) initAgent() *agentS {

	log.debug("initializing agent")

	ret := new(agentS)
	ret.sensor = r
	ret.init()
	ret.reset()

	return ret
}
