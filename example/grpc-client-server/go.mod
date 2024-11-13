module github.com/instana/go-sensor/example/grpc-client-server

go 1.22.7

require (
	github.com/instana/go-sensor v1.65.0
	github.com/instana/go-sensor/instrumentation/instagrpc v1.11.0
	github.com/opentracing/opentracing-go v1.2.0
	google.golang.org/grpc v1.68.0
	google.golang.org/protobuf v1.34.2
)

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/looplab/fsm v1.0.1 // indirect
	golang.org/x/net v0.29.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	golang.org/x/text v0.18.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240903143218-8af14fe29dc1 // indirect
)

replace (
	github.com/instana/go-sensor v1.58.0 => ../../../go-sensor
	github.com/instana/go-sensor/instrumentation/instagrpc v1.11.0 => ../../instrumentation/instagrpc
)
