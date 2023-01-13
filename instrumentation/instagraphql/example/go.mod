module example.com

go 1.19

require github.com/graphql-go/graphql v0.8.0

require (
	github.com/instana/go-sensor v1.49.0
	github.com/looplab/fsm v0.1.0 // indirect
	github.com/opentracing/opentracing-go v1.1.0
)

replace github.com/instana/go-sensor => ../../..