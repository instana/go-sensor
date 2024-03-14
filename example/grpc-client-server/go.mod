module github.com/instana/go-sensor/example/grpc-client-server

go 1.19

require (
	github.com/instana/go-sensor v1.60.0
	github.com/instana/go-sensor/instrumentation/instagrpc v1.11.0
	github.com/opentracing/opentracing-go v1.2.0
	google.golang.org/grpc v1.62.1
	google.golang.org/protobuf v1.33.0
)

require (
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/looplab/fsm v1.0.1 // indirect
	golang.org/x/net v0.20.0 // indirect
	golang.org/x/sys v0.16.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240123012728-ef4313101c80 // indirect
)

replace (
	github.com/instana/go-sensor v1.58.0 => ../../../go-sensor
	github.com/instana/go-sensor/instrumentation/instagrpc v1.11.0 => ../../instrumentation/instagrpc
)
