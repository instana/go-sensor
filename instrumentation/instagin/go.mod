module github.com/instana/go-sensor/instrumentation/instagin

go 1.11

require (
	github.com/gin-gonic/gin v1.7.2
	github.com/instana/go-sensor v1.44.0
	github.com/instana/testify v1.6.2-0.20200721153833-94b1851f4d65
	github.com/opentracing/opentracing-go v1.2.0
	golang.org/x/sys v0.0.0-20220503163025-988cb79eb6c6 // indirect
)

replace github.com/instana/go-sensor => ../../