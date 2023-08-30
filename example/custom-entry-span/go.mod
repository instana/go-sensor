module custom-entry-span.com

go 1.18

require (
	github.com/instana/go-sensor v1.55.2
	github.com/opentracing/opentracing-go v1.2.0
)

require github.com/looplab/fsm v1.0.1 // indirect

replace github.com/instana/go-sensor => ../..
