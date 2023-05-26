module github.com/instana/go-sensor/example/gin

go 1.15

require (
	github.com/gin-gonic/gin v1.9.0
	github.com/instana/go-sensor v1.55.0
	github.com/instana/go-sensor/instrumentation/instagin v1.6.0
)

replace github.com/instana/go-sensor/instrumentation/instagin => ../../instrumentation/instagin
