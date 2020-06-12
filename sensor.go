package instana

import (
	"errors"
	"os"

	"github.com/instana/go-sensor/autoprofile"
	"github.com/instana/go-sensor/logger"
)

const (
	DefaultMaxBufferedSpans = 1000
	DefaultForceSpanSendAt  = 500
)

type sensorS struct {
	meter       *meterS
	agent       *agentS
	logger      LeveledLogger
	options     *Options
	serviceName string
}

var sensor *sensorS

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
	s.setLogger(defaultLogger)

	// override service name with an env value if set
	if name, ok := os.LookupEnv("INSTANA_SERVICE_NAME"); ok {
		s.serviceName = name
	}

	// handle the legacy (instana.Options).LogLevel value if we use logger.Logger to log
	if l, ok := s.logger.(*logger.Logger); ok {
		setLogLevel(l, options.LogLevel)
	}

	s.agent = newAgent(s.serviceName, s.options.AgentHost, s.options.AgentPort, s.logger)
	s.meter = newMeter(s.agent, s.logger)

	return s
}

func (r *sensorS) setLogger(l LeveledLogger) {
	r.logger = l

	if r.agent != nil {
		r.agent.setLogger(r.logger)
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

	// enable auto-profiling
	if options.EnableAutoProfile {
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

		autoprofile.Enable()
	}

	sensor.logger.Debug("initialized sensor")
}
