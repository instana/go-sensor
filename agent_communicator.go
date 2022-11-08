// (c) Copyright IBM Corp. 2022

package instana

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
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
}

// urlBuilder builds an Agent URL based on the sufix for the different Agent services.
func (a agentCommunicator) urlBuilder(sufix string) string {
	url := "http://" + a.host + ":" + a.port + sufix

	if sufix[len(sufix)-1:] == "." && a.from.EntityID != "" {
		url += a.from.EntityID
	}

	return url
}

// getServerHeader attempts to retrieve the "Server" header key from the Agent
func (a *agentCommunicator) getServerHeader() string {
	url := a.urlBuilder("/")
	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		return ""
	}

	resp, err := a.client.Do(req)

	if resp == nil {
		return ""
	}

	defer func() {
		io.CopyN(ioutil.Discard, resp.Body, 256<<10)
		resp.Body.Close()
	}()

	if err == nil {
		return resp.Header.Get("Server")
	}

	return ""
}

// getAgentResponse attempts to retrieve the agent response containing its configuration
func (a *agentCommunicator) getAgentResponse(d *discoveryS) *agentResponse {
	jsonData, _ := json.Marshal(d)

	var resp agentResponse

	u := a.urlBuilder(agentDiscoveryURL)

	req, err := http.NewRequest(http.MethodPut, u, bytes.NewBuffer(jsonData))

	if err != nil {
		return nil
	}

	res, err := a.client.Do(req)

	if res == nil {
		return nil
	}

	defer func() {
		io.CopyN(ioutil.Discard, res.Body, 256<<10)
		res.Body.Close()
	}()

	badResponse := res.StatusCode < 200 || res.StatusCode >= 300

	if err != nil || badResponse {
		return nil
	}

	respBytes, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return nil
	}

	err = json.Unmarshal(respBytes, &resp)

	if err != nil {
		return nil
	}

	return &resp
}

// pingAgent send a HEAD request to the agent and returns true if it receives a response from it
func (a *agentCommunicator) pingAgent() bool {
	u := a.urlBuilder(agentDataURL)
	req, err := http.NewRequest(http.MethodHead, u, nil)

	if err != nil {
		return false
	}

	resp, err := a.client.Do(req)

	if err != nil || resp == nil {
		return false
	}

	defer func() {
		io.CopyN(ioutil.Discard, resp.Body, 256<<10)
		resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return false
	}

	return true
}

// sendDataToAgent makes a POST to the agent sending some data as payload. eg: spans, events or metrics
func (a *agentCommunicator) sendDataToAgent(suffix string, data interface{}) error {
	url := a.urlBuilder(suffix)
	ctx, cancel := context.WithTimeout(context.Background(), clientTimeout)
	defer cancel()

	var r io.Reader

	if data != nil {
		b, err := json.Marshal(data)

		if err != nil {
			return err
		}

		r = bytes.NewBuffer(b)

		// if r > maxContentLength {
		// 	// agent.logger.Warn(`A batch of spans has been rejected because it is too large to be sent to the agent.`)
		// 	return payloadTooLargeErr
		// }
	}

	req, err := http.NewRequest(http.MethodPost, url, r)

	if err != nil {
		return err
	}

	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)

	if resp != nil {
		io.CopyN(ioutil.Discard, resp.Body, 256<<10)
		resp.Body.Close()
	}

	return err
}

func newAgentCommunicator(host, port string, from *fromS) *agentCommunicator {
	return &agentCommunicator{
		host: host,
		port: port,
		from: from,
		client: &http.Client{
			Timeout: announceTimeout,
		},
	}
}
