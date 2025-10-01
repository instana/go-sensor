module example.com/instagraphql

go 1.23.0

require (
	github.com/google/uuid v1.6.0
	github.com/gorilla/websocket v1.5.3
	github.com/graphql-go/graphql v0.8.1
	github.com/graphql-go/handler v0.2.4
	github.com/instana/go-sensor v1.71.1-fedramp
	github.com/instana/go-sensor/instrumentation/instagraphql v1.31.0-fedramp
)

require (
	github.com/google/pprof v0.0.0-20250630185457-6e76a2b096b5 // indirect
	github.com/looplab/fsm v1.0.3 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/instana/go-sensor => ./../..
	github.com/instana/go-sensor/instrumentation/instagraphql => ./../../instrumentation/instagraphql
)
