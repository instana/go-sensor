package instana

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
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
	Zone     string
	Tags     map[string]interface{}

	snapshot serverlessSnapshot

	mu        sync.Mutex
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
		Zone:     os.Getenv("INSTANA_ZONE"),
		Tags:     parseInstanaTags(os.Getenv("INSTANA_TAGS")),
		client:   client,
		logger:   logger,
	}

	go func() {
		t := time.NewTicker(time.Second)
		defer t.Stop()

		for range t.C {
			if err := agent.Flush(context.Background()); err != nil {
				agent.logger.Error("failed to post collected data: ", err)
			}
		}
	}()

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

	a.mu.Lock()
	spans := make([]Span, len(a.spanQueue))
	copy(spans, a.spanQueue)
	a.spanQueue = a.spanQueue[:0]
	a.mu.Unlock()

	payload := struct {
		Metrics metricsPayload `json:"metrics,omitempty"`
		Spans   []agentSpan    `json:"spans,omitempty"`
	}{
		Metrics: metricsPayload{
			Plugins: []acceptor.PluginPayload{
				acceptor.NewAWSLambdaPluginPayload(snapshot.EntityID),
			},
		},
		Spans: make([]agentSpan, 0, len(spans)),
	}

	for _, sp := range spans {
		payload.Spans = append(payload.Spans, agentSpan{
			Span: sp,
			From: from,
		})
	}

	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(payload); err != nil {
		return fmt.Errorf("failed to marshal traces payload: %s", err)
	}

	req, err := http.NewRequest(http.MethodPost, a.Endpoint+"/bundle", buf)
	if err != nil {
		a.enqueueSpans(spans)
		return fmt.Errorf("failed to prepare send traces request: %s", err)
	}

	req.Header.Set("Content-Type", "application/json")

	if err := a.sendRequest(req.WithContext(ctx)); err != nil {
		a.enqueueSpans(spans)
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
	req.Header.Set("X-Instana-Host", a.snapshot.Host)
	req.Header.Set("X-Instana-Key", a.Key)
	req.Header.Set("X-Instana-Time", strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10))

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to the serverless agent: %s", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			a.logger.Debug("failed to read serverless agent response: ", err)
			return nil
		}

		a.logger.Info("serverless agent has responded with ", resp.Status, ": ", string(respBody))
		return nil
	}

	io.CopyN(ioutil.Discard, resp.Body, 1<<20)

	return nil
}

func (a *lambdaAgent) collectSnapshot(spans []Span) serverlessSnapshot {
	if a.snapshot.EntityID != "" {
		return a.snapshot
	}

	// searching for the lambda entry span in reverse order, since it's
	// more likely to be finished last
	for i := len(spans) - 1; i >= 0; i-- {
		sp, ok := spans[i].Data.(AWSLambdaSpanData)
		if !ok {
			log.Printf("span data type %t", spans[i].Data)
			continue
		}

		a.snapshot = serverlessSnapshot{
			EntityID: sp.Snapshot.ARN,
			Host:     sp.Snapshot.ARN,
			Zone:     a.Zone,
			Tags:     a.Tags,
		}
		a.logger.Debug("collected snapshot")

		break
	}

	return a.snapshot
}
