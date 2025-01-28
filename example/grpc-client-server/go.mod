module github.com/instana/go-sensor/example/grpc-client-server

go 1.22.7

require (
	github.com/instana/go-sensor v1.66.1
	github.com/instana/go-sensor/instrumentation/instagrpc v1.11.0
	github.com/opentracing/opentracing-go v1.2.0
	google.golang.org/grpc v1.70.0
	google.golang.org/protobuf v1.35.2
)

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/looplab/fsm v1.0.1 // indirect
	golang.org/x/net v0.33.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241202173237-19429a94021a // indirect
)

replace (
	github.com/instana/go-sensor v1.58.0 => ../../../go-sensor
	github.com/instana/go-sensor/instrumentation/instagrpc v1.11.0 => ../../instrumentation/instagrpc
)
