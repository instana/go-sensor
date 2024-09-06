module http-database-greeter

go 1.22

require (
	github.com/instana/go-sensor v1.59.0
	github.com/lib/pq v1.10.9
	github.com/opentracing/opentracing-go v1.2.0
)

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/looplab/fsm v1.0.1 // indirect
)

replace github.com/instana/go-sensor => ../../
