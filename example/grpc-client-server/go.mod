module github.com/instana/go-sensor/example/grpc-client-server

go 1.23.0

require (
	github.com/instana/go-sensor v1.67.1
	github.com/instana/go-sensor/instrumentation/instagrpc v1.37.2
	github.com/opentracing/opentracing-go v1.2.0
	google.golang.org/grpc v1.71.1
	google.golang.org/protobuf v1.36.4
)

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/looplab/fsm v1.0.1 // indirect
	golang.org/x/net v0.36.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.22.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250115164207-1a7da9e5054f // indirect
)

replace (
	github.com/instana/go-sensor v1.67.1 => ../../../go-sensor
	github.com/instana/go-sensor/instrumentation/instagrpc v1.37.2 => ../../instrumentation/instagrpc
)
