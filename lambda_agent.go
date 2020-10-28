package instana

import (
	"net/http"
	"os"

	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"
)

type lambdaAgent struct {
	Endpoint string
	Key      string
	PID      int
	Zone     string
	Tags     map[string]interface{}

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

	return &lambdaAgent{
		Endpoint: acceptorEndpoint,
		Key:      agentKey,
		PID:      os.Getpid(),
		Zone:     os.Getenv("INSTANA_ZONE"),
		Tags:     parseInstanaTags(os.Getenv("INSTANA_TAGS")),
		client:   client,
		logger:   logger,
	}
}

func (a *lambdaAgent) Ready() bool { return false }

func (a *lambdaAgent) SendMetrics(data acceptor.Metrics) error { return nil }

func (a *lambdaAgent) SendEvent(event *EventData) error { return nil }

func (a *lambdaAgent) SendSpans(spans []Span) error { return nil }

func (a *lambdaAgent) SendProfiles(profiles []autoprofile.Profile) error { return nil }
