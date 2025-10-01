module instagoredisv2-example

go 1.23.0

require (
	github.com/bonede/go-redis-driver v0.1.0
	github.com/instana/go-sensor v1.71.1-fedramp
	github.com/instana/go-sensor/instrumentation/instaredis/v2 v2.38.0-fedramp
	github.com/redis/go-redis/v9 v9.14.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/google/pprof v0.0.0-20250630185457-6e76a2b096b5 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/looplab/fsm v1.0.3 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/instana/go-sensor => ../../

replace github.com/instana/go-sensor/instrumentation/instaredis/v2 => ../../instrumentation/instaredis/v2
