module github.com/instana/go-sensor/instrumentation/instagocb

go 1.13

require (
	github.com/couchbase/gocb/v2 v2.6.5
	github.com/couchbase/gocbcore/v10 v10.2.9
	github.com/instana/go-sensor v1.59.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/stretchr/testify v1.8.2
)

replace github.com/instana/go-sensor => ../../
