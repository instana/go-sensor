module example.com

go 1.13

require (
	github.com/google/uuid v1.4.0
	github.com/gorilla/websocket v1.5.0
	github.com/graphql-go/graphql v0.8.1
	github.com/graphql-go/handler v0.2.3
	github.com/instana/go-sensor v1.58.0
	github.com/instana/go-sensor/instrumentation/instagraphql v1.6.0
)

replace (
	github.com/instana/go-sensor => ./../..
	github.com/instana/go-sensor/instrumentation/instagraphql => ./../../instrumentation/instagraphql
)
