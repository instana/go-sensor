module github.com/instana/go-sensor/example/httprouter

go 1.23

require (
	github.com/instana/go-sensor v1.67.7
	github.com/instana/go-sensor/instrumentation/instahttprouter v1.29.0
	github.com/julienschmidt/httprouter v1.3.0
)

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/looplab/fsm v1.0.3 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
)

replace (
	github.com/instana/go-sensor => ../../../go-sensor
	github.com/instana/go-sensor/instrumentation/instahttprouter => ../../instrumentation/instahttprouter
)
