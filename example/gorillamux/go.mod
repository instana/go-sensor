module github.com/instana/go-sensor/example/gorillamux

go 1.16

require (
	github.com/gorilla/mux v1.8.0
	github.com/instana/go-sensor v1.31.0
	github.com/instana/go-sensor/instrumentation/instagorillamux v0.0.0-00010101000000-000000000000
)

replace (
	github.com/instana/go-sensor => ../../.
	github.com/instana/go-sensor/instrumentation/instagorillamux => ../../instrumentation/instagorillamux
)
