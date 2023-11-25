module github.com/instana/go-sensor/instrumentation/instagocb

go 1.13

require (
	github.com/couchbase/gocb/v2 v2.7.0
	github.com/couchbase/gocbcore/v10 v10.3.0
	github.com/instana/go-sensor v1.58.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/stretchr/testify v1.8.4
)

replace github.com/instana/go-sensor => ../../
