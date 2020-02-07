package instana

import (
	"os"
	"path/filepath"
)

const (
	DefaultMaxBufferedSpans = 1000
	DefaultForceSpanSendAt  = 500
)

type sensorS struct {
	meter       *meterS
	agent       *agentS
	options     *Options
	serviceName string
}

var sensor *sensorS

func (r *sensorS) init(options *Options) {
	// sensor can be initialized explicitly or implicitly through OpenTracing global init
	if r.meter == nil {
		r.setOptions(options)
		r.configureServiceName()
		r.agent = r.initAgent()
		r.meter = r.initMeter()
	}
}

func (r *sensorS) setOptions(options *Options) {
	r.options = options
	if r.options == nil {
		r.options = &Options{}
	}

	if r.options.MaxBufferedSpans == 0 {
		r.options.MaxBufferedSpans = DefaultMaxBufferedSpans
	}

	if r.options.ForceTransmissionStartingAt == 0 {
		r.options.ForceTransmissionStartingAt = DefaultForceSpanSendAt
	}
}

func (r *sensorS) configureServiceName() {
	if name, ok := os.LookupEnv("INSTANA_SERVICE_NAME"); ok {
		r.serviceName = name
		return
	}

	if r.options != nil {
		r.serviceName = r.options.Service
	}

	if r.serviceName == "" {
		r.serviceName = filepath.Base(os.Args[0])
	}
}

// InitSensor intializes the sensor (without tracing) to begin collecting
// and reporting metrics.
func InitSensor(options *Options) {
	if sensor != nil {
		return
	}

	sensor = &sensorS{}

	// If this environment variable is set, then override log level
	_, ok := os.LookupEnv("INSTANA_DEBUG")
	if ok {
		options.LogLevel = Debug
	}

	sensor.initLog()
	sensor.init(options)
	log.debug("initialized sensor")
}
