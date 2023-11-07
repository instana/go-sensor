module couchbase.example

go 1.21.3

require (
	github.com/couchbase/gocb/v2 v2.6.5
	// github.com/instana/go-sensor@gocb-instrumentation
	// github.com/instana/go-sensor/instrumentation/instagocb@gocb-instrumentation
	github.com/redis/go-redis/v9 v9.3.0
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/couchbase/gocbcore/v10 v10.2.9 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/uuid v1.3.1 // indirect
	github.com/instana/go-sensor v1.58.1-0.20231107055240-4ac1225b817a // indirect
	github.com/instana/go-sensor/instrumentation/instagocb v0.0.0-20231107055240-4ac1225b817a // indirect
	github.com/looplab/fsm v1.0.1 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
)

// replace github.com/instana/go-sensor/instrumentation/instagocb => ../../instrumentation/instagocb

// replace github.com/instana/go-sensor => ../../
