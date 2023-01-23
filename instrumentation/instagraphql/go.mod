module github.com/instana/go-sensor/instrumentation/instagraphql

go 1.19

require (
	github.com/graphql-go/graphql v0.8.0
	github.com/instana/go-sensor v1.49.0
	github.com/stretchr/testify v1.8.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require (
	github.com/looplab/fsm v0.1.0 // indirect
	github.com/opentracing/opentracing-go v1.1.0
)

replace github.com/instana/go-sensor => ../../
