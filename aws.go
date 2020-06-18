package instana

import (
	"net/http"

	"github.com/instana/go-sensor/autoprofile"
)

type fargateAgent struct {
	Endpoint string
	Key      string

	client *http.Client
	logger LeveledLogger
}

func newFargateAgent(acceptorEndpoint, agentKey string, c *http.Client, logger LeveledLogger) *fargateAgent {
	if c == nil {
		c = http.DefaultClient
	}

	if logger == nil {
		logger = defaultLogger
	}

	logger.Debug("initializing aws fargate agent")

	return &fargateAgent{
		Endpoint: acceptorEndpoint,
		Key:      agentKey,
		client:   c,
		logger:   logger,
	}
}

func (a *fargateAgent) Ready() bool { return true }

func (a *fargateAgent) SendMetrics(data *MetricsS) error { return nil }

func (a *fargateAgent) SendEvent(event *EventData) error                  { return nil }
func (a *fargateAgent) SendSpans(spans []Span) error                      { return nil }
func (a *fargateAgent) SendProfiles(profiles []autoprofile.Profile) error { return nil }
