module instagoredisv2-example

go 1.23

require (
	github.com/bonede/go-redis-driver v0.1.0
	github.com/instana/go-sensor v1.67.5
	github.com/instana/go-sensor/instrumentation/instaredis/v2 v2.24.0
	github.com/redis/go-redis/v9 v9.11.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/looplab/fsm v1.0.3 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
)

replace github.com/instana/go-sensor => ../../

replace github.com/instana/go-sensor/instrumentation/instaredis/v2 => ../../instrumentation/instaredis/v2
