module http-database-greeter

go 1.23.0

require (
	github.com/instana/go-sensor v1.68.0
	github.com/lib/pq v1.10.9
	github.com/opentracing/opentracing-go v1.2.0
)

require (
	github.com/google/pprof v0.0.0-20250630185457-6e76a2b096b5 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/looplab/fsm v1.0.3 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/instana/go-sensor => ../../
