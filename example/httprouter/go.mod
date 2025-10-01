module github.com/instana/go-sensor/example/httprouter

go 1.23.0

require (
	github.com/instana/go-sensor v1.71.1-fedramp
	github.com/instana/go-sensor/instrumentation/instahttprouter v1.34.0-fedramp
	github.com/julienschmidt/httprouter v1.3.0
)

require (
	github.com/google/pprof v0.0.0-20250630185457-6e76a2b096b5 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/looplab/fsm v1.0.3 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/instana/go-sensor => ../../../go-sensor
	github.com/instana/go-sensor/instrumentation/instahttprouter => ../../instrumentation/instahttprouter
)
