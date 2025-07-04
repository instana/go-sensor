module instana-opentracing

go 1.23

require (
	github.com/instana/go-sensor v1.67.7
	github.com/opentracing/opentracing-go v1.2.0
)

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/looplab/fsm v1.0.3 // indirect
)

replace github.com/instana/go-sensor => ../../../go-sensor
