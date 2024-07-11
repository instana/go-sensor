module example.com/instagraphql

go 1.21

require (
	github.com/google/uuid v1.4.0
	github.com/gorilla/websocket v1.5.0
	github.com/graphql-go/graphql v0.8.1
	github.com/graphql-go/handler v0.2.3
	github.com/instana/go-sensor v1.63.1
	github.com/instana/go-sensor/instrumentation/instagraphql v1.6.0
)

require (
	github.com/looplab/fsm v1.0.1 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
)

replace (
	github.com/instana/go-sensor => ./../..
	github.com/instana/go-sensor/instrumentation/instagraphql => ./../../instrumentation/instagraphql
)
