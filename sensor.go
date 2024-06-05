// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2016

package instana

import (
	"context"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"
	"github.com/instana/go-sensor/aws"
	"github.com/instana/go-sensor/logger"
)

const (
	// DefaultMaxBufferedSpans is the default span buffer size
	DefaultMaxBufferedSpans = 1000
	// DefaultForceSpanSendAt is the default max number of spans to buffer before force sending them to the agent
	DefaultForceSpanSendAt = 500

	// TODO: defaultServerlessTimeout is increased from 500 millisecond to 2 second
	// as serverless API latency is high. This should be reduced once latency is minimized.
	defaultServerlessTimeout = 2 * time.Second
)

// aws constants
const (
	awsExecutionEnv         = "AWS_EXECUTION_ENV"
	awsECSFargate           = "AWS_ECS_FARGATE"
	ecsContainerMetadataURI = "ECS_CONTAINER_METADATA_URI"
	awsLambdaPrefix         = "AWS_Lambda_"
)

// knative constants
const (
	kService       = "K_SERVICE"
	kConfiguration = "K_CONFIGURATION"
	kRevision      = "K_REVISION"
)

// azure constants
const (
	containerAppHostName  = "CONTAINER_APP_HOSTNAME"
	azureFunctionsRuntime = "FUNCTIONS_WORKER_RUNTIME"
)

type AgentClient interface {
	Ready() bool
	SendMetrics(data acceptor.Metrics) error
	SendEvent(event *EventData) error
	SendSpans(spans []Span) error
	SendProfiles(profiles []autoprofile.Profile) error
	Flush(context.Context) error
}

// zero value for sensorS.Agent()
type noopAgent struct{}

func (noopAgent) Ready() bool                                       { return false }
func (noopAgent) SendMetrics(data acceptor.Metrics) error           { return nil }
func (noopAgent) SendEvent(event *EventData) error                  { return nil }
func (noopAgent) SendSpans(spans []Span) error                      { return nil }
func (noopAgent) SendProfiles(profiles []autoprofile.Profile) error { return nil }
func (noopAgent) Flush(context.Context) error                       { return nil }

type sensorS struct {
	meter       *meterS
	logger      LeveledLogger
	options     *Options
	serviceName string
	binaryName  string

	mu    sync.RWMutex
	agent AgentClient
}

var (
	sensor           *sensorS
	muSensor         sync.Mutex
	binaryName       = filepath.Base(os.Args[0])
	processStartedAt = time.Now()
	C                TracerLogger
)

func init() {
	C = newNoopCollector()
}

func newSensor(options *Options) *sensorS {
	options.setDefaults()

	s := &sensorS{
		options:     options,
		serviceName: options.Service,
		binaryName:  binaryName,
	}

	s.setLogger(defaultLogger)

	// override service name with an env value if set
	if name, ok := os.LookupEnv("INSTANA_SERVICE_NAME"); ok && strings.TrimSpace(name) != "" {
		s.serviceName = name
	}

	// handle the legacy (instana.Options).LogLevel value if we use logger.Logger to log
	if l, ok := s.logger.(*logger.Logger); ok {

		_, isInstanaLogLevelSet := os.LookupEnv("INSTANA_LOG_LEVEL")

		if !isInstanaLogLevelSet {
			setLogLevel(l, options.LogLevel)
		}
	}

	var agent AgentClient

	if options.AgentClient != nil {
		agent = options.AgentClient
	}

	if agentEndpoint := os.Getenv("INSTANA_ENDPOINT_URL"); agentEndpoint != "" && agent == nil {
		s.logger.Debug("INSTANA_ENDPOINT_URL= is set, switching to the serverless mode")

		timeout, err := parseInstanaTimeout(os.Getenv("INSTANA_TIMEOUT"))
		if err != nil {
			s.logger.Warn("malformed INSTANA_TIMEOUT value, falling back to the default one: ", err)
			timeout = defaultServerlessTimeout
		}

		client, err := acceptor.NewHTTPClient(timeout)
		if err != nil {
			if err == acceptor.ErrMalformedProxyURL {
				s.logger.Warn(err)
			} else {
				s.logger.Error("failed to initialize acceptor HTTP client, falling back to the default one: ", err)
				client = http.DefaultClient
			}
		}

		agent = newServerlessAgent(s.serviceOrBinaryName(), agentEndpoint, os.Getenv("INSTANA_AGENT_KEY"), client, s.logger)
	}

	if agent == nil {
		agent = newAgent(s.serviceOrBinaryName(), s.options.AgentHost, s.options.AgentPort, s.logger)
	}

	s.setAgent(agent)
	s.meter = newMeter(s.logger)

	return s
}

func (r *sensorS) setLogger(l LeveledLogger) {
	r.logger = l

	if agent, ok := r.Agent().(*agentS); ok && agent != nil {
		agent.setLogger(r.logger)
	}
}

func (r *sensorS) setAgent(agent AgentClient) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.agent = agent
}

// Agent returns the agent client used by the global sensor. It will return a noopAgent that is never ready
// until both the global sensor and its agent are initialized
func (r *sensorS) Agent() AgentClient {
	if r == nil {
		return noopAgent{}
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.agent == nil {
		return noopAgent{}
	}

	return r.agent
}

func (r *sensorS) serviceOrBinaryName() string {
	if r == nil {
		return ""
	}
	if r.serviceName != "" {
		return r.serviceName
	}
	return r.binaryName
}

// InitSensor initializes the sensor (without tracing) to begin collecting
// and reporting metrics.
//
// Deprecated: Use [StartMetrics] instead.
func InitSensor(options *Options) {
	if sensor != nil {
		return
	}

	if options == nil {
		options = DefaultOptions()
	}

	muSensor.Lock()
	sensor = newSensor(options)
	muSensor.Unlock()

	// configure auto-profiling
	autoprofile.SetLogger(sensor.logger)
	autoprofile.SetOptions(autoprofile.Options{
		IncludeProfilerFrames: options.IncludeProfilerFrames,
		MaxBufferedProfiles:   options.MaxBufferedProfiles,
	})

	autoprofile.SetSendProfilesFunc(func(profiles []autoprofile.Profile) error {
		if !sensor.Agent().Ready() {
			return errors.New("sender not ready")
		}

		sensor.logger.Debug("sending profiles to agent")

		return sensor.Agent().SendProfiles(profiles)
	})

	if _, ok := os.LookupEnv("INSTANA_AUTO_PROFILE"); ok || options.EnableAutoProfile {
		if !options.EnableAutoProfile {
			sensor.logger.Info("INSTANA_AUTO_PROFILE is set, activating AutoProfile™")
		}

		autoprofile.Enable()
	}

	// start collecting metrics
	go sensor.meter.Run(1 * time.Second)

	sensor.logger.Debug("initialized Instana sensor v", Version)
}

// StartMetrics initializes the communication with the agent. Then it starts collecting and reporting metrics to the agent.
// Calling StartMetrics multiple times has no effect and the function will simply return, and provided options will not
// be reapplied.
func StartMetrics(options *Options) {
	InitSensor(options)
}

// Ready returns whether the Instana collector is ready to collect and send data to the agent
func Ready() bool {
	if sensor == nil {
		return false
	}

	return sensor.Agent().Ready()
}

// Flush forces Instana collector to send all buffered data to the agent. This method is intended to implement
// graceful service shutdown and not recommended for intermittent use. Once Flush() is called, it's not guaranteed
// that collector remains in operational state.
func Flush(ctx context.Context) error {
	if sensor == nil {
		return nil
	}

	return sensor.Agent().Flush(ctx)
}

// ShutdownSensor cleans up the internal global sensor reference. The next time that instana.InitSensor is called,
// directly or indirectly, the internal sensor will be reinitialized.
func ShutdownSensor() {
	muSensor.Lock()
	if sensor != nil {
		sensor = nil
	}
	muSensor.Unlock()
}

func newServerlessAgent(serviceName, agentEndpoint, agentKey string,
	client *http.Client, logger LeveledLogger) AgentClient {

	switch {
	// AWS Fargate
	case os.Getenv(awsExecutionEnv) == awsECSFargate &&
		os.Getenv(ecsContainerMetadataURI) != "":
		return newFargateAgent(
			serviceName,
			agentEndpoint,
			agentKey,
			client,
			aws.NewECSMetadataProvider(os.Getenv(ecsContainerMetadataURI), client),
			logger,
		)

	// AWS Lambda
	case strings.HasPrefix(os.Getenv(awsExecutionEnv), awsLambdaPrefix):
		return newLambdaAgent(serviceName, agentEndpoint, agentKey, client, logger)

	// Knative, e.g. Google Cloud Run
	case os.Getenv(kService) != "" && os.Getenv(kConfiguration) != "" &&
		os.Getenv(kRevision) != "":
		return newGCRAgent(serviceName, agentEndpoint, agentKey, client, logger)

	// azure functions or container apps
	case os.Getenv(azureFunctionsRuntime) == azureCustomRuntime ||
		os.Getenv(containerAppHostName) != "":
		return newAzureAgent(agentEndpoint, agentKey, client, logger)
	default:
		return nil
	}
}
