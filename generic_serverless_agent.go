// (c) Copyright IBM Corp. 2024

package instana

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"
)

const (
	flushPeriodForGenericInSec = 2
)

type genericServerlessAgent struct {
	Endpoint   string
	Key        string
	PluginName string
	PID        int

	snapshot serverlessSnapshot

	mu        sync.RWMutex
	spanQueue []Span

	client *http.Client
	logger LeveledLogger
}

func newGenericServerlessAgent(acceptorEndpoint, agentKey string, client *http.Client, logger LeveledLogger) *genericServerlessAgent {
	if logger == nil {
		logger = defaultLogger
	}

	if client == nil {
		client = http.DefaultClient
		// You can change this timeout by setting the INSTANA_TIMEOUT environment variable.
		client.Timeout = 2 * time.Second
	}

	logger.Debug("initializing generic serverless agent")

	// Creating a unique serverless host ID.
	uniqHostId := "Generic_Serverless_Agent" + uuid.New().String()

	agent := &genericServerlessAgent{
		Endpoint: acceptorEndpoint,
		Key:      agentKey,
		PID:      os.Getpid(),
		client:   client,
		logger:   logger,
		snapshot: serverlessSnapshot{
			Host:     uniqHostId,
			EntityID: uniqHostId,
		},
	}

	go func(a *genericServerlessAgent) {
		t := time.NewTicker(flushPeriodForGenericInSec * time.Second)
		defer t.Stop()

		for range t.C {
			if err := a.Flush(context.Background()); err != nil {
				a.logger.Error("failed to post collected data: ", err)
			}
		}
	}(agent)

	return agent
}

func (a *genericServerlessAgent) Ready() bool { return true }

func (a *genericServerlessAgent) SendMetrics(acceptor.Metrics) error { return nil }

func (a *genericServerlessAgent) SendEvent(*EventData) error { return nil }

func (a *genericServerlessAgent) SendSpans(spans []Span) error {
	a.enqueueSpans(spans)
	return nil
}

func (a *genericServerlessAgent) SendProfiles([]autoprofile.Profile) error { return nil }

func (a *genericServerlessAgent) Flush(ctx context.Context) error {
	a.mu.RLock()
	entityID := a.snapshot.EntityID
	a.mu.RUnlock()

	from := newServerlessAgentFromS(entityID, "generic_serverless")

	payload := struct {
		Spans []Span `json:"spans,omitempty"`
	}{}

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
			flushPeriodForGenericInSec, err.Error())
	}

	return nil
}

func (a *genericServerlessAgent) enqueueSpans(spans []Span) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.spanQueue = append(a.spanQueue, spans...)
}

func (a *genericServerlessAgent) sendRequest(req *http.Request) error {
	req.Header.Set("X-Instana-Host", a.snapshot.Host)
	req.Header.Set("X-Instana-Key", a.Key)

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to the serverless agent: %s", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			a.logger.Debug("failed to read serverless agent response: ", err.Error())
			return err
		}

		a.logger.Info("serverless agent has responded with ", resp.Status, ": ", string(respBody))
		return err
	}

	io.CopyN(io.Discard, resp.Body, 1<<20)

	return nil
}
