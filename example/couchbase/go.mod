module couchbase.example

go 1.21

toolchain go1.21.0

require github.com/couchbase/gocb/v2 v2.8.1

require (
	github.com/instana/go-sensor v1.62.0
	github.com/instana/go-sensor/instrumentation/instagocb v1.4.1
)

require (
	github.com/couchbase/gocbcore/v10 v10.4.1 // indirect
	github.com/couchbase/gocbcoreps v0.1.2 // indirect
	github.com/couchbase/goprotostellar v1.0.2 // indirect
	github.com/couchbaselabs/gocbconnstr/v2 v2.0.0-20230515165046-68b522a21131 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0 // indirect
	github.com/looplab/fsm v1.0.1 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/net v0.24.0 // indirect
	golang.org/x/sys v0.19.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240401170217-c3f982113cda // indirect
	google.golang.org/grpc v1.63.2 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
)

replace github.com/instana/go-sensor/instrumentation/instagocb => ../../instrumentation/instagocb
