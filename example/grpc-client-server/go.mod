module github.com/instana/go-sensor/example/grpc-client-server

go 1.23.0

require (
	github.com/instana/go-sensor v1.67.3
	github.com/instana/go-sensor/instrumentation/instagrpc v1.37.2
	github.com/opentracing/opentracing-go v1.2.0
	google.golang.org/grpc v1.72.0
	google.golang.org/protobuf v1.36.6
)

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/looplab/fsm v1.0.2 // indirect
	golang.org/x/net v0.39.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/text v0.24.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250421163800-61c742ae3ef0 // indirect
)

replace (
	github.com/instana/go-sensor v1.67.1 => ../../../go-sensor
	github.com/instana/go-sensor/instrumentation/instagrpc v1.37.2 => ../../instrumentation/instagrpc
)
