module github.com/instana/go-sensor/example/gin

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/instana/go-sensor v1.60.0
	github.com/instana/go-sensor/instrumentation/instagin v1.6.0
	github.com/kr/pretty v0.3.0 // indirect
	github.com/rogpeppe/go-internal v1.8.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)

replace (
	github.com/instana/go-sensor => ../../../go-sensor
	github.com/instana/go-sensor/instrumentation/instagin => ../../instrumentation/instagin
)
