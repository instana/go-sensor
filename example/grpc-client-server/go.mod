module github.com/instana/go-sensor/example/grpc-client-server

go 1.23.0

require (
	github.com/instana/go-sensor v1.71.1-fedramp
	github.com/instana/go-sensor/instrumentation/instagrpc v1.52.0-fedramp
	github.com/opentracing/opentracing-go v1.2.0
	google.golang.org/grpc v1.75.1
	google.golang.org/protobuf v1.36.6
)

require (
	github.com/google/pprof v0.0.0-20250630185457-6e76a2b096b5 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/looplab/fsm v1.0.3 // indirect
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250707201910-8d1bb00bc6a7 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/instana/go-sensor v1.67.1 => ../../../go-sensor
	github.com/instana/go-sensor/instrumentation/instagrpc v1.37.2 => ../../instrumentation/instagrpc
)
