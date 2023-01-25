module github.com/instana/go-sensor/instrumentation/instagraphql

go 1.13

require (
	github.com/graphql-go/graphql v0.8.0
	github.com/instana/go-sensor v1.49.0
	github.com/stretchr/testify v1.8.1
)

require github.com/opentracing/opentracing-go v1.1.0

replace github.com/instana/go-sensor => ../../
