module couchbase.example

go 1.19

require github.com/couchbase/gocb/v2 v2.7.2

require (
	github.com/instana/go-sensor v1.60.0
	github.com/instana/go-sensor/instrumentation/instagocb v0.0.0-20231107055240-4ac1225b817a
)

require (
	github.com/couchbase/gocbcore/v10 v10.3.2 // indirect
	github.com/couchbase/gocbcoreps v0.1.2 // indirect
	github.com/couchbase/goprotostellar v1.0.2 // indirect
	github.com/couchbaselabs/gocbconnstr/v2 v2.0.0-20230515165046-68b522a21131 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0 // indirect
	github.com/looplab/fsm v1.0.1 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.26.0 // indirect
	golang.org/x/net v0.20.0 // indirect
	golang.org/x/sys v0.16.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240213162025-012b6fc9bca9 // indirect
	google.golang.org/grpc v1.61.1 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
)

replace github.com/instana/go-sensor/instrumentation/instagocb => ../../instrumentation/instagocb
