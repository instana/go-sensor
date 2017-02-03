package instana

import (
	"os"
	"path/filepath"
)

type sensorS struct {
	meter       *meterS
	agent       *agentS
	options     *Options
	serviceName string
}

var sensor *sensorS

func (r *sensorS) init(options *Options) {
	//sensor can be initialized explicit or implicit through OpenTracing global init
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
}

func (r *sensorS) getOptions() *Options {
	return r.options
}

func (r *sensorS) configureServiceName() {
	if r.options != nil {
		r.serviceName = r.options.Service
	}

	if r.serviceName == "" {
		r.serviceName = filepath.Base(os.Args[0])
	}
}

func InitSensor(options *Options) {
	sensor = new(sensorS)
	sensor.init(options)

	Logger.Println("initialized sensor")
}
