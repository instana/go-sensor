// (c) Copyright IBM Corp. 2022

package instana

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"
)

const (
	flushPeriodInSec = 2

	azureCustomRuntime string = "custom"
)

type azureAgent struct {
	Endpoint string
	Key      string
	PID      int

	snapshot serverlessSnapshot

	mu        sync.Mutex
	spanQueue []Span

	client *http.Client
	logger LeveledLogger
}

func newAzureAgent(acceptorEndpoint, agentKey string, client *http.Client, logger LeveledLogger) *azureAgent {
	if logger == nil {
		logger = defaultLogger
	}

	if client == nil {
		client = http.DefaultClient
		client.Timeout = 500 * time.Millisecond
	}

	logger.Debug("initializing azure agent")

	agent := &azureAgent{
		Endpoint: acceptorEndpoint,
		Key:      agentKey,
		PID:      os.Getpid(),
		client:   client,
		logger:   logger,
	}

	go func() {
		t := time.NewTicker(flushPeriodInSec * time.Second)
		defer t.Stop()

		for range t.C {
			if err := agent.Flush(context.Background()); err != nil {
				agent.logger.Error("failed to post collected data: ", err)
			}
		}
	}()

	return agent
}

func (a *azureAgent) Ready() bool { return true }

func (a *azureAgent) SendMetrics(acceptor.Metrics) error { return nil }

func (a *azureAgent) SendEvent(*EventData) error { return nil }

func (a *azureAgent) SendSpans(spans []Span) error {
	a.enqueueSpans(spans)
	return nil
}

func (a *azureAgent) SendProfiles([]autoprofile.Profile) error { return nil }

func (a *azureAgent) Flush(ctx context.Context) error {
	a.collectSnapshot()

	if a.snapshot.EntityID == "" {
		return ErrAgentNotReady
	}

	from := newServerlessAgentFromS(a.snapshot.EntityID, "azure")

	payload := struct {
		Metrics metricsPayload `json:"metrics,omitempty"`
		Spans   []Span         `json:"spans,omitempty"`
	}{
		Metrics: metricsPayload{
			Plugins: []acceptor.PluginPayload{
				acceptor.NewAzurePluginPayload(a.snapshot.EntityID),
			},
		},
	}

	a.mu.Lock()
	payload.Spans = make([]Span, len(a.spanQueue))
	copy(payload.Spans, a.spanQueue)
	a.spanQueue = a.spanQueue[:0]
	a.mu.Unlock()

	for i := range payload.Spans {
		payload.Spans[i].From = from
	}

	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(payload); err != nil {
		return fmt.Errorf("failed to marshal traces payload: %s", err)
	}

	payloadSize := buf.Len()
	if payloadSize > maxContentLength {
		a.logger.Warn(fmt.Sprintf("failed to send the spans. Payload size: %d exceeded max size: %d", payloadSize, maxContentLength))
		return payloadTooLargeErr
	}

	req, err := http.NewRequest(http.MethodPost, a.Endpoint+"/bundle", buf)
	if err != nil {
		a.enqueueSpans(payload.Spans)
		return fmt.Errorf("failed to prepare send traces request: %s", err)
	}

	req.Header.Set("Content-Type", "application/json")

	if err := a.sendRequest(req.WithContext(ctx)); err != nil {
		a.enqueueSpans(payload.Spans)
		return fmt.Errorf("failed to send traces, will retry later: %dsec. Error details: %s",
			flushPeriodInSec, err.Error())
	}

	return nil
}

func (a *azureAgent) enqueueSpans(spans []Span) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.spanQueue = append(a.spanQueue, spans...)
}

func (a *azureAgent) sendRequest(req *http.Request) error {
	req.Header.Set("X-Instana-Host", a.snapshot.Host)
	req.Header.Set("X-Instana-Key", a.Key)

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to the serverless agent: %s", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			a.logger.Debug("failed to read serverless agent response: ", err.Error())
			return err
		}

		a.logger.Info("serverless agent has responded with ", resp.Status, ": ", string(respBody))
		return err
	}

	io.CopyN(ioutil.Discard, resp.Body, 1<<20)

	return nil
}

func (a *azureAgent) collectSnapshot() {
	if a.snapshot.EntityID != "" {
		return
	}

	var subscriptionID, resourceGrp, functionApp string
	var ok bool
	if websiteOwnerName, ok := os.LookupEnv("WEBSITE_OWNER_NAME"); ok {
		arr := strings.Split(websiteOwnerName, "+")
		if len(arr) > 1 {
			subscriptionID = arr[0]
		} else {
			a.logger.Warn("failed to retrieve the subscription id. This will affect the correlation metrics.")
		}
	} else {
		a.logger.Warn("failed to retrieve the subscription id. This will affect the correlation metrics.")
	}

	if resourceGrp, ok = os.LookupEnv("WEBSITE_RESOURCE_GROUP"); !ok {
		a.logger.Warn("failed to retrieve the resource group. This will affect the correlation metrics.")
	}

	if functionApp, ok = os.LookupEnv("APPSETTING_WEBSITE_SITE_NAME"); !ok {
		a.logger.Warn("failed to retrieve the function app. This will affect the correlation metrics.")
	}

	entityID := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Web/sites/%s",
		subscriptionID, resourceGrp, functionApp)

	a.snapshot = serverlessSnapshot{
		EntityID: entityID,
		PID:      a.PID,
	}
	a.logger.Debug("collected snapshot")
}
