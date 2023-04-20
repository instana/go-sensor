// (c) Copyright IBM Corp. 2022

package instana

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

// agentCommunicator is a collection of data and actions to be executed against the agent.
type agentCommunicator struct {
	// host is the agent host. It can be updated via default gateway or a new client announcement.
	host string

	// port id the agent port.
	port string

	// from is the agent information sent with each span in the "from" (span.f) section. it's format is as follows:
	// {e: "entityId", h: "hostAgentId", hl: trueIfServerlessPlatform, cp: "The cloud provider for a hostless span"}
	// Only span.f.e is mandatory.
	from *fromS

	// client is an HTTP client
	client httpClient

	// l is the Instana logger
	l LeveledLogger
}

// buildURL builds an Agent URL based on the sufix for the different Agent services.
func (a *agentCommunicator) buildURL(sufix string) string {
	url := "http://" + a.host + ":" + a.port + sufix

	if sufix[len(sufix)-1:] == "." && a.from.EntityID != "" {
		url += a.from.EntityID
	}

	return url
}

// serverHeader attempts to retrieve the "Server" header key from the Agent
func (a *agentCommunicator) serverHeader() string {
	url := a.buildURL("/")
	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		a.l.Debug("Error creating request to get the agent header: ", err.Error())
		return ""
	}

	resp, err := a.client.Do(req)

	if resp == nil {
		a.l.Debug("No response from the agent. Error: ", err.Error())
		return ""
	}

	defer func() {
		io.CopyN(ioutil.Discard, resp.Body, 256<<10)
		resp.Body.Close()
	}()

	if err == nil {
		return resp.Header.Get("Server")
	}

	a.l.Debug("Error requesting header from the agent: ", err.Error())

	return ""
}

// agentResponse attempts to retrieve the agent response containing its configuration
func (a *agentCommunicator) agentResponse(d *discoveryS) *agentResponse {
	jsonData, _ := json.Marshal(d)

	var resp agentResponse

	u := a.buildURL(agentDiscoveryURL)

	req, err := http.NewRequest(http.MethodPut, u, bytes.NewBuffer(jsonData))

	if err != nil {
		a.l.Debug("Error creating request to get the agent response:", err.Error())
		return nil
	}

	res, err := a.client.Do(req)

	if res == nil {
		a.l.Debug("No response from the agent. Error:", err.Error())
		return nil
	}

	defer func() {
		io.CopyN(ioutil.Discard, res.Body, 256<<10)
		res.Body.Close()
	}()

	badResponse := res.StatusCode < 200 || res.StatusCode >= 300

	if err != nil || badResponse {
		a.l.Debug("Error requesting response from agent", err, "Bad response:", badResponse)
		return nil
	}

	respBytes, err := ioutil.ReadAll(res.Body)

	if err != nil {
		a.l.Debug("Error reading res.Body", err.Error())
		return nil
	}

	err = json.Unmarshal(respBytes, &resp)

	if err != nil {
		a.l.Debug("Error unmarshaling body:", err.Error())
		return nil
	}

	return &resp
}

// pingAgent send a HEAD request to the agent and returns true if it receives a response from it
func (a *agentCommunicator) pingAgent() bool {
	u := a.buildURL(agentDataURL)
	req, err := http.NewRequest(http.MethodHead, u, nil)

	if err != nil {
		a.l.Debug("Error preparing request:", err.Error())
		return false
	}

	resp, err := a.client.Do(req)

	if err != nil || resp == nil {
		a.l.Debug("Error pinging the agent:", err.Error(), ", response: ", resp)
		return false
	}

	defer func() {
		io.CopyN(ioutil.Discard, resp.Body, 256<<10)
		resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		a.l.Debug("Agent ping failed, response: ", resp.StatusCode)
		return false
	}

	a.l.Debug("Agent ping ok!")

	return true
}

// sendDataToAgent makes a POST to the agent sending some data as payload. eg: spans, events or metrics
func (a *agentCommunicator) sendDataToAgent(suffix string, data interface{}) error {
	url := a.buildURL(suffix)
	ctx, cancel := context.WithTimeout(context.Background(), clientTimeout)
	defer cancel()

	var r *bytes.Buffer

	if data != nil {
		b, err := json.Marshal(data)

		if err != nil {
			a.l.Debug("Data to agent marshaling failed:", err.Error())
			return err
		}

		r = bytes.NewBuffer(b)

		if r.Len() > maxContentLength {
			return payloadTooLargeErr
		}
	}

	req, err := http.NewRequest(http.MethodPost, url, r)

	if err != nil {
		a.l.Debug("Data to agent request creation failed:", err.Error())
		return err
	}

	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)

	if resp == nil {
		a.l.Debug("sendDataToAgent: response nil for ", url)
	} else {
		if resp.StatusCode != 200 {
			a.l.Debug("sendDataToAgent: response code != 200: ", resp)
			if _, ok := os.LookupEnv("INSTANA_GO_FIX"); ok {
				return errors.New("sendDataToAgent: response code != 200: " + resp.Status)
			}
		}
	}

	if resp != nil {
		io.CopyN(ioutil.Discard, resp.Body, 256<<10)
		resp.Body.Close()
	}

	if err != nil {
		a.l.Debug("Data to agent request failed:", err.Error())
	}

	return err
}

func newAgentCommunicator(host, port string, from *fromS, logger LeveledLogger) *agentCommunicator {
	return &agentCommunicator{
		host: host,
		port: port,
		from: from,
		client: &http.Client{
			Timeout: announceTimeout,
		},
		l: logger,
	}
}
