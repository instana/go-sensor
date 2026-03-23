module github.com/instana/go-sensor/example/grpc-client-server

go 1.24.0

require (
	github.com/instana/go-sensor v1.68.0
	github.com/instana/go-sensor/instrumentation/instagrpc v1.46.0
	github.com/opentracing/opentracing-go v1.2.0
	google.golang.org/grpc v1.79.3
	google.golang.org/protobuf v1.36.10
)

require (
	github.com/google/pprof v0.0.0-20250630185457-6e76a2b096b5 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/looplab/fsm v1.0.3 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
)

replace (
	github.com/instana/go-sensor v1.67.1 => ../../../go-sensor
	github.com/instana/go-sensor/instrumentation/instagrpc v1.37.2 => ../../instrumentation/instagrpc
)
