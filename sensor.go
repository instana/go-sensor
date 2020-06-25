package instana

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"
	"github.com/instana/go-sensor/aws"
	"github.com/instana/go-sensor/logger"
)

const (
	DefaultMaxBufferedSpans = 1000
	DefaultForceSpanSendAt  = 500
)

type agentClient interface {
	Ready() bool
	SendMetrics(data acceptor.Metrics) error
	SendEvent(event *EventData) error
	SendSpans(spans []Span) error
	SendProfiles(profiles []autoprofile.Profile) error
}

type sensorS struct {
	meter       *meterS
	agent       agentClient
	logger      LeveledLogger
	options     *Options
	serviceName string
}

var sensor *sensorS
var binaryName = filepath.Base(os.Args[0])

func newSensor(options *Options) *sensorS {
	if options == nil {
		options = DefaultOptions()
	} else {
		options.setDefaults()
	}

	s := &sensorS{
		options:     options,
		serviceName: options.Service,
	}
	if s.serviceName == "" {
		s.serviceName = binaryName
	}

	s.setLogger(defaultLogger)

	// override service name with an env value if set
	if name, ok := os.LookupEnv("INSTANA_SERVICE_NAME"); ok {
		s.serviceName = name
	}

	// handle the legacy (instana.Options).LogLevel value if we use logger.Logger to log
	if l, ok := s.logger.(*logger.Logger); ok {
		setLogLevel(l, options.LogLevel)
	}

	switch {
	case os.Getenv("AWS_EXECUTION_ENV") == "AWS_ECS_FARGATE":
		// seems like the app is running on AWS Fargate, but we still need to check
		// whether ECS_CONTAINER_METADATA_URI is set
		if mdURI := os.Getenv("ECS_CONTAINER_METADATA_URI"); mdURI != "" {
			s.agent = newFargateAgent(
				s.serviceName,
				os.Getenv("INSTANA_ENDPOINT_URL"),
				os.Getenv("INSTANA_AGENT_KEY"),
				aws.NewECSMetadataProvider(mdURI, nil),
				s.logger,
			)

			break
		}

		s.logger.Error("the task metadata URI is not set in ECS_CONTAINER_METADATA_URI env var, falling back to host agent mode")

		fallthrough
	default:
		s.agent = newAgent(s.serviceName, s.options.AgentHost, s.options.AgentPort, s.logger)
	}

	s.meter = newMeter(s.agent, s.logger)

	return s
}

func (r *sensorS) setLogger(l LeveledLogger) {
	r.logger = l

	if agent, ok := r.agent.(*agentS); ok && agent != nil {
		agent.setLogger(r.logger)
	}

	if r.meter != nil {
		r.meter.setLogger(r.logger)
	}
}

// InitSensor intializes the sensor (without tracing) to begin collecting
// and reporting metrics.
func InitSensor(options *Options) {
	if sensor != nil {
		return
	}

	sensor = newSensor(options)

	// configure auto-profiling
	autoprofile.SetLogger(sensor.logger)
	autoprofile.SetOptions(autoprofile.Options{
		IncludeProfilerFrames: options.IncludeProfilerFrames,
		MaxBufferedProfiles:   options.MaxBufferedProfiles,
	})

	autoprofile.SetSendProfilesFunc(func(profiles []autoprofile.Profile) error {
		if !sensor.agent.Ready() {
			return errors.New("sender not ready")
		}

		sensor.logger.Debug("sending profiles to agent")

		return sensor.agent.SendProfiles(profiles)
	})

	if _, ok := os.LookupEnv("INSTANA_AUTO_PROFILE"); ok || options.EnableAutoProfile {
		if !options.EnableAutoProfile {
			sensor.logger.Info("INSTANA_AUTO_PROFILE is set, activating AutoProfile™")
		}

		autoprofile.Enable()
	}

	sensor.logger.Debug("initialized sensor")
}
