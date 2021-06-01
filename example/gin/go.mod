module github.com/instana/go-sensor/example/gin

go 1.11

require (
	github.com/gin-gonic/gin v1.7.2

	github.com/instana/go-sensor v1.29.0
	github.com/instana/go-sensor/instrumentation/instagin v1.0.0
)

// todo: remove before release
replace github.com/instana/go-sensor/instrumentation/instagin => ../../instrumentation/instagin
replace github.com/instana/go-sensor => ../../
