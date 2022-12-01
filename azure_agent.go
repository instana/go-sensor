// (c) Copyright IBM Corp. 2022
// (c) Copyright Instana Inc. 2022

package instana

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"
)

const (
	azureAgentFlushPeriod = 2 * time.Second

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

func newAzureAgent(serviceName, acceptorEndpoint, agentKey string, client *http.Client, logger LeveledLogger) *azureAgent {
	if logger == nil {
		logger = defaultLogger
	}

	if client == nil {
		client = http.DefaultClient
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
		t := time.NewTicker(azureAgentFlushPeriod)
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

func (a *azureAgent) SendMetrics(data acceptor.Metrics) error { return nil }

func (a *azureAgent) SendEvent(event *EventData) error { return nil }

func (a *azureAgent) SendSpans(spans []Span) error {
	a.enqueueSpans(spans)
	return nil
}

func (a *azureAgent) SendProfiles(profiles []autoprofile.Profile) error { return nil }

func (a *azureAgent) Flush(ctx context.Context) error {
	snapshot := a.collectSnapshot(a.spanQueue)

	if snapshot.EntityID == "" {
		return ErrAgentNotReady
	}

	from := newServerlessAgentFromS(snapshot.EntityID, "azure")

	payload := struct {
		Metrics metricsPayload `json:"metrics,omitempty"`
		Spans   []Span         `json:"spans,omitempty"`
	}{
		Metrics: metricsPayload{
			Plugins: []acceptor.PluginPayload{
				acceptor.NewAzurePluginPayload(snapshot.EntityID),
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

	req, err := http.NewRequest(http.MethodPost, a.Endpoint+"/bundle", buf)
	if err != nil {
		a.enqueueSpans(payload.Spans)
		return fmt.Errorf("failed to prepare send traces request: %s", err)
	}

	req.Header.Set("Content-Type", "application/json")

	if err := a.sendRequest(req.WithContext(ctx)); err != nil {
		a.enqueueSpans(payload.Spans)
		return fmt.Errorf("failed to send traces, will retry later: %s", err)
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
	req.Header.Set("X-Instana-Time", strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10))

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to the serverless agent: %s", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			a.logger.Debug("failed to read serverless agent response: ", err)
			return nil
		}

		a.logger.Info("serverless agent has responded with ", resp.Status, ": ", string(respBody))
		return nil
	}

	io.CopyN(io.Discard, resp.Body, 1<<20)

	return nil
}

func (a *azureAgent) collectSnapshot(spans []Span) serverlessSnapshot {
	if a.snapshot.EntityID != "" {
		return a.snapshot
	}

	var subscriptionID, resourceGrp, functionApp string
	if val, ok := os.LookupEnv("WEBSITE_OWNER_NAME"); ok {
		arr := strings.Split(val, "+")
		if len(arr) > 1 {
			subscriptionID = arr[0]
		}
	}

	if val, ok := os.LookupEnv("WEBSITE_RESOURCE_GROUP"); ok {
		resourceGrp = val
	}

	if val, ok := os.LookupEnv("APPSETTING_WEBSITE_SITE_NAME"); ok {
		functionApp = val
	}

	entityID := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Web/sites/%s",
		subscriptionID, resourceGrp, functionApp)

	a.snapshot = serverlessSnapshot{
		EntityID: entityID,
	}
	a.logger.Debug("collected snapshot")

	return a.snapshot
}
