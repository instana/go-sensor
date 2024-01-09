module couchbase.example

go 1.13

require github.com/couchbase/gocb/v2 v2.6.5

require (
	github.com/instana/go-sensor v1.59.0
	github.com/instana/go-sensor/instrumentation/instagocb v0.0.0-20231107055240-4ac1225b817a
)

replace github.com/instana/go-sensor/instrumentation/instagocb => ../../instrumentation/instagocb
