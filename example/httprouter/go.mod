module github.com/instana/go-sensor/example/httprouter

go 1.13

require (
	github.com/instana/go-sensor v1.59.0
	github.com/instana/go-sensor/instrumentation/instahttprouter v1.10.0
	github.com/julienschmidt/httprouter v1.3.0
)

replace (
	github.com/instana/go-sensor => ../../../go-sensor
	github.com/instana/go-sensor/instrumentation/instahttprouter => ../../instrumentation/instahttprouter
)
