module http-client

go 1.24.6

require (
	github.com/instana/go-sensor v1.72.0
	github.com/instana/go-sensor/instrumentation/instalogrus v1.38.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/sirupsen/logrus v1.9.3
)

require (
	github.com/google/pprof v0.0.0-20250630185457-6e76a2b096b5 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/looplab/fsm v1.0.3 // indirect
	golang.org/x/sys v0.34.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/instana/go-sensor => ../../
	github.com/instana/go-sensor/instrumentation/instalogrus => ../../instrumentation/instalogrus
)
