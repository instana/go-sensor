package instana

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"
	"github.com/instana/go-sensor/gcloud"
)

const googleCloudRunMetadataURL = "http://metadata.google.internal"

type gcrSnapshot struct {
	EntityID string
	PID      int
	Zone     string
	Tags     map[string]interface{}
	Metadata gcloud.ComputeMetadata
}

func newGCRSnapshot(pid int, md gcloud.ComputeMetadata) gcrSnapshot {
	return gcrSnapshot{
		PID:      pid,
		EntityID: md.Instance.ID,
	}
}

type gcrAgent struct {
	Endpoint string
	Key      string
	PID      int
	Zone     string
	Tags     map[string]interface{}

	snapshot gcrSnapshot

	gcr    *gcloud.ComputeMetadataProvider
	client *http.Client
	logger LeveledLogger
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
		gcr:      gcloud.NewComputeMetadataProvider(mdURL, client),
		client:   client,
		logger:   logger,
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

func (a *gcrAgent) Ready() bool {
	return a.snapshot.EntityID != ""
}

func (a *gcrAgent) SendMetrics(data acceptor.Metrics) error { return nil }

func (a *gcrAgent) SendEvent(event *EventData) error { return nil }

func (a *gcrAgent) SendSpans(spans []Span) error { return nil }

func (a *gcrAgent) SendProfiles(profiles []autoprofile.Profile) error { return nil }

func (a *gcrAgent) collectSnapshot(ctx context.Context) (gcrSnapshot, bool) {
	md, err := a.gcr.ComputeMetadata(ctx)
	if err != nil {
		a.logger.Warn("failed to get service metadata: ", err)
		return gcrSnapshot{}, false
	}

	snapshot := newGCRSnapshot(a.PID, md)
	snapshot.Zone = a.Zone
	snapshot.Tags = a.Tags

	a.logger.Debug("collected snapshot")

	return snapshot, true
}
