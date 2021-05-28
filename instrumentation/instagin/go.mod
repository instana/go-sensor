module github.com/instana/go-sensor/instrumentation/instagin

go 1.16

require (
	github.com/gin-gonic/gin v1.7.2
	github.com/instana/go-sensor v1.29.0
	github.com/opentracing/opentracing-go v1.2.0
)

replace (
	github.com/instana/go-sensor => ../../
)