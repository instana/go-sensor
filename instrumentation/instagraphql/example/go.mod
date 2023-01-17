module example.com

go 1.19

require (
	github.com/graphql-go/graphql v0.8.0
	github.com/instana/go-sensor/instrumentation/instagraphql v0.0.0-00010101000000-000000000000
)

require (
	github.com/instana/go-sensor v1.49.0
	github.com/looplab/fsm v0.1.0 // indirect
	github.com/opentracing/opentracing-go v1.1.0 // indirect
)

replace github.com/instana/go-sensor => ../../..

replace github.com/instana/go-sensor/instrumentation/instagraphql => ../
