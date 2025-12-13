// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

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
	"sync"
	"time"

	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"
)

const awsLambdaAgentFlushPeriod = 2 * time.Second

type lambdaAgent struct {
	Endpoint string
	Key      string
	PID      int

	snapshot serverlessSnapshot

	mu        sync.RWMutex
	spanQueue []Span

	client *http.Client
	logger LeveledLogger
}

func newLambdaAgent(
	serviceName, acceptorEndpoint, agentKey string,
	client *http.Client,
	logger LeveledLogger,
) *lambdaAgent {
	if logger == nil {
		logger = defaultLogger
	}

	if client == nil {
		client = http.DefaultClient
	}

	logger.Debug("initializing aws lambda agent")

	agent := &lambdaAgent{
		Endpoint: acceptorEndpoint,
		Key:      agentKey,
		PID:      os.Getpid(),
		client:   client,
		logger:   logger,
	}

	go func(a *lambdaAgent) {
		t := time.NewTicker(awsLambdaAgentFlushPeriod)
		defer t.Stop()

		for range t.C {
			if err := a.Flush(context.Background()); err != nil {
				a.logger.Error("failed to post collected data: ", err)
			}
		}
	}(agent)

	return agent
}

func (a *lambdaAgent) Ready() bool { return true }

func (a *lambdaAgent) SendMetrics(data acceptor.Metrics) error { return nil }

func (a *lambdaAgent) SendEvent(event *EventData) error { return nil }

func (a *lambdaAgent) SendSpans(spans []Span) error {
	a.enqueueSpans(spans)
	return nil
}

func (a *lambdaAgent) SendProfiles(profiles []autoprofile.Profile) error { return nil }

func (a *lambdaAgent) Flush(ctx context.Context) error {
	snapshot := a.collectSnapshot(a.spanQueue)

	if snapshot.EntityID == "" {
		return ErrAgentNotReady
	}

	from := newServerlessAgentFromS(snapshot.EntityID, "aws")

	payload := struct {
		Metrics metricsPayload `json:"metrics,omitempty"`
		Spans   []Span         `json:"spans,omitempty"`
	}{
		Metrics: metricsPayload{
			Plugins: []acceptor.PluginPayload{
				acceptor.NewAWSLambdaPluginPayload(snapshot.EntityID),
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

func (a *lambdaAgent) enqueueSpans(spans []Span) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.spanQueue = append(a.spanQueue, spans...)
}

func (a *lambdaAgent) sendRequest(req *http.Request) error {
	a.mu.RLock()
	host := a.snapshot.Host
	a.mu.RUnlock()

	req.Header.Set("X-Instana-Host", host)
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

func (a *lambdaAgent) collectSnapshot(spans []Span) serverlessSnapshot {
	a.mu.RLock()
	entityID := a.snapshot.EntityID
	a.mu.RUnlock()

	if entityID != "" {
		return a.snapshot
	}

	// searching for the lambda entry span in reverse order, since it's
	// more likely to be finished last
	for i := len(spans) - 1; i >= 0; i-- {
		sp, ok := spans[i].Data.(AWSLambdaSpanData)
		if !ok {
			continue
		}

		a.mu.Lock()
		a.snapshot = serverlessSnapshot{
			EntityID: sp.Snapshot.ARN,
			Host:     sp.Snapshot.ARN,
		}
		a.mu.Unlock()
		a.logger.Debug("collected snapshot")

		break
	}

	return a.snapshot
}
