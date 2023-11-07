module github.com/instana/go-sensor/instrumentation/instagocb

go 1.19

require (
	github.com/couchbase/gocb/v2 v2.6.5
	github.com/instana/go-sensor v1.58.0
	github.com/opentracing/opentracing-go v1.2.0
)

require (
	github.com/couchbase/gocbcore/v10 v10.2.9 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/uuid v1.3.1 // indirect
	github.com/looplab/fsm v1.0.1 // indirect
)

replace github.com/instana/go-sensor => ../../
