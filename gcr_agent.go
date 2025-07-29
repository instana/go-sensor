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
	"strings"
	"sync"
	"time"

	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"
	"github.com/instana/go-sensor/gcloud"
)

const googleCloudRunMetadataURL = "http://metadata.google.internal"

type gcrMetadata struct {
	gcloud.ComputeMetadata

	Service       string
	Configuration string
	Revision      string
	Port          string
}

type gcrSnapshot struct {
	Service  serverlessSnapshot
	Metadata gcrMetadata
}

func newGCRSnapshot(pid int, md gcrMetadata) gcrSnapshot {
	return gcrSnapshot{
		Service: serverlessSnapshot{
			EntityID:  md.Instance.ID,
			Host:      "gcp:cloud-run:revision:" + md.Revision,
			PID:       pid,
			StartedAt: processStartedAt,
			Container: containerSnapshot{
				ID:   md.Instance.ID,
				Type: "gcpCloudRunInstance",
			},
		},
		Metadata: md,
	}
}

func newGCRServiceRevisionInstancePluginPayload(snapshot gcrSnapshot) acceptor.PluginPayload {
	regionName := snapshot.Metadata.Instance.Region
	if ind := strings.LastIndexByte(regionName, '/'); ind >= 0 {
		// truncate projects/<projectID>/regions/ prefix to extract the region
		// from a fully-qualified name
		regionName = regionName[ind+1:]
	}

	return acceptor.NewGCRServiceRevisionInstancePluginPayload(snapshot.Service.EntityID, acceptor.GCRServiceRevisionInstanceData{
		Runtime:          "go",
		Region:           regionName,
		Service:          snapshot.Metadata.Service,
		Configuration:    snapshot.Metadata.Configuration,
		Revision:         snapshot.Metadata.Revision,
		InstanceID:       snapshot.Metadata.Instance.ID,
		Port:             snapshot.Metadata.Port,
		NumericProjectID: snapshot.Metadata.Project.NumericProjectID,
		ProjectID:        snapshot.Metadata.Project.ProjectID,
	})
}

type gcrAgent struct {
	Endpoint string
	Key      string
	PID      int
	Zone     string
	Tags     map[string]interface{}

	snapshot         gcrSnapshot
	lastProcessStats processStats

	mu        sync.RWMutex
	spanQueue []Span

	runtimeSnapshot *SnapshotCollector
	processStats    *processStatsCollector
	gcr             *gcloud.ComputeMetadataProvider
	client          *http.Client
	logger          LeveledLogger
}

func newGCRAgent(
	serviceName, acceptorEndpoint, agentKey string,
	client *http.Client,
	logger LeveledLogger,
) *gcrAgent {
	if logger == nil {
		logger = defaultLogger
	}

	if client == nil {
		client = http.DefaultClient
	}

	logger.Debug("initializing google cloud run agent")

	// allow overriding the metadata URL endpoint for testing purposes
	mdURL, ok := os.LookupEnv("GOOGLE_CLOUD_RUN_METADATA_ENDPOINT")
	if !ok {
		mdURL = googleCloudRunMetadataURL
	}

	agent := &gcrAgent{
		Endpoint: acceptorEndpoint,
		Key:      agentKey,
		PID:      os.Getpid(),
		Zone:     os.Getenv("INSTANA_ZONE"),
		Tags:     parseInstanaTags(os.Getenv("INSTANA_TAGS")),
		runtimeSnapshot: &SnapshotCollector{
			CollectionInterval: snapshotCollectionInterval,
			ServiceName:        serviceName,
		},
		processStats: &processStatsCollector{
			logger: logger,
		},
		gcr:    gcloud.NewComputeMetadataProvider(mdURL, client),
		client: client,
		logger: logger,
	}

	go func() {
		for {
			for i := 0; i < maximumRetries; i++ {
				snapshot, ok := agent.collectSnapshot(context.Background())
				if ok {
					agent.mu.Lock()
					agent.snapshot = snapshot
					agent.mu.Unlock()
					break
				}

				time.Sleep(expDelay(i + 1))
			}
			time.Sleep(snapshotCollectionInterval)
		}
	}()
	go agent.processStats.Run(context.Background(), time.Second)

	return agent
}

func (a *gcrAgent) Ready() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.snapshot.Service.EntityID != ""
}

func (a *gcrAgent) SendMetrics(data acceptor.Metrics) (err error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	processStats := a.processStats.Collect()

	defer func() {
		if err == nil {
			a.mu.Lock()
			defer a.mu.Unlock()
			// only update the last sent stats if they were transmitted successfully
			// since they are updated on the backend incrementally using received
			// deltas
			a.lastProcessStats = processStats
		}
	}()

	payload := struct {
		Metrics metricsPayload `json:"metrics,omitempty"`
		Spans   []Span         `json:"spans,omitempty"`
	}{
		Metrics: metricsPayload{
			Plugins: []acceptor.PluginPayload{
				newGCRServiceRevisionInstancePluginPayload(a.snapshot),
				newProcessPluginPayload(a.snapshot.Service, a.lastProcessStats, processStats),
				acceptor.NewGoProcessPluginPayload(acceptor.GoProcessData{
					PID:      a.PID,
					Snapshot: a.runtimeSnapshot.Collect(),
					Metrics:  data,
				}),
			},
		},
	}

	if len(a.spanQueue) > 0 {
		payload.Spans = make([]Span, len(a.spanQueue))
		copy(payload.Spans, a.spanQueue)
		a.spanQueue = a.spanQueue[:0]
	}

	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(payload); err != nil {
		return fmt.Errorf("failed to marshal metrics payload: %s", err)
	}

	req, err := http.NewRequest(http.MethodPost, a.Endpoint+"/bundle", buf)
	if err != nil {
		return fmt.Errorf("failed to prepare send metrics request: %s", err)
	}

	req.Header.Set("Content-Type", "application/json")

	return a.sendRequest(req)
}

func (a *gcrAgent) SendEvent(event *EventData) error { return nil }

func (a *gcrAgent) SendSpans(spans []Span) error {
	from := newServerlessAgentFromS(a.snapshot.Service.EntityID, "gcp")
	for i := range spans {
		spans[i].From = from
	}

	// enqueue the spans to send them in a bundle with metrics instead of sending immediately
	a.mu.Lock()
	a.spanQueue = append(a.spanQueue, spans...)
	a.mu.Unlock()

	return nil
}

func (a *gcrAgent) SendProfiles(profiles []autoprofile.Profile) error { return nil }

func (a *gcrAgent) Flush(ctx context.Context) error {
	if len(a.spanQueue) == 0 {
		return nil
	}

	if !a.Ready() {
		return ErrAgentNotReady
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(a.spanQueue); err != nil {
		return fmt.Errorf("failed to marshal traces payload: %s", err)
	}
	a.spanQueue = a.spanQueue[:0]

	req, err := http.NewRequest(http.MethodPost, a.Endpoint+"/traces", buf)
	if err != nil {
		return fmt.Errorf("failed to prepare send traces request: %s", err)
	}

	req.Header.Set("Content-Type", "application/json")

	return a.sendRequest(req.WithContext(ctx))
}

func (a *gcrAgent) sendRequest(req *http.Request) error {
	req.Header.Set("X-Instana-Host", a.snapshot.Service.Host)
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

func (a *gcrAgent) collectSnapshot(ctx context.Context) (gcrSnapshot, bool) {
	md, err := a.gcr.ComputeMetadata(ctx)
	if err != nil {
		a.logger.Warn("failed to get service metadata: ", err)
		return gcrSnapshot{}, false
	}

	snapshot := newGCRSnapshot(a.PID, gcrMetadata{
		ComputeMetadata: md,
		Service:         os.Getenv(kService),
		Configuration:   os.Getenv(kConfiguration),
		Revision:        os.Getenv(kRevision),
		Port:            os.Getenv("PORT"),
	})
	snapshot.Service.Zone = a.Zone
	snapshot.Service.Tags = a.Tags

	a.logger.Debug("collected snapshot")

	return snapshot, true
}
