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
	"strconv"
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
}

type gcrSnapshot struct {
	Service  serverlessSnapshot
	Metadata gcrMetadata
}

func newGCRSnapshot(pid int, md gcrMetadata) gcrSnapshot {
	return gcrSnapshot{
		Service: serverlessSnapshot{
			EntityID: md.Instance.ID,
			Host:     "gcp:cloud-run:revision:" + md.Revision,
			PID:      pid,
			Container: containerSnapshot{
				ID:   md.Instance.ID,
				Type: "gcpCloudRunInstance",
			},
		},
		Metadata: md,
	}
}

type gcrAgent struct {
	Endpoint string
	Key      string
	PID      int
	Zone     string
	Tags     map[string]interface{}

	snapshot gcrSnapshot

	runtimeSnapshot *SnapshotCollector
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
		gcr:    gcloud.NewComputeMetadataProvider(mdURL, client),
		client: client,
		logger: logger,
	}

	go func() {
		for {
			for i := 0; i < maximumRetries; i++ {
				snapshot, ok := agent.collectSnapshot(context.Background())
				if ok {
					agent.snapshot = snapshot
					break
				}

				time.Sleep(retryPeriod)
			}
			time.Sleep(snapshotCollectionInterval)
		}
	}()

	return agent
}

func (a *gcrAgent) Ready() bool { return a.snapshot.Service.EntityID != "" }

func (a *gcrAgent) SendMetrics(data acceptor.Metrics) error {
	payload := struct {
		Metrics metricsPayload `json:"metrics,omitempty"`
		Spans   []agentSpan    `json:"spans,omitempty"`
	}{
		Metrics: metricsPayload{
			Plugins: []acceptor.PluginPayload{
				acceptor.NewGoProcessPluginPayload(acceptor.GoProcessData{
					PID:      a.PID,
					Snapshot: a.runtimeSnapshot.Collect(),
					Metrics:  data,
				}),
			},
		},
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

func (a *gcrAgent) SendSpans(spans []Span) error { return nil }

func (a *gcrAgent) SendProfiles(profiles []autoprofile.Profile) error { return nil }

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

func (a *gcrAgent) collectSnapshot(ctx context.Context) (gcrSnapshot, bool) {
	md, err := a.gcr.ComputeMetadata(ctx)
	if err != nil {
		a.logger.Warn("failed to get service metadata: ", err)
		return gcrSnapshot{}, false
	}

	snapshot := newGCRSnapshot(a.PID, gcrMetadata{
		ComputeMetadata: md,
		Service:         os.Getenv("K_SERVICE"),
		Configuration:   os.Getenv("K_CONFIGURATION"),
		Revision:        os.Getenv("K_REVISION"),
	})
	snapshot.Service.Zone = a.Zone
	snapshot.Service.Tags = a.Tags

	a.logger.Debug("collected snapshot")

	return snapshot, true
}
