module github.com/instana/go-sensor/instrumentation/instasarama/example

go 1.13

require (
	github.com/Shopify/sarama v1.19.0
	github.com/instana/go-sensor v1.49.0
	github.com/instana/go-sensor/instrumentation/instasarama v1.1.0
	github.com/opentracing/opentracing-go v1.1.0
)

replace github.com/instana/go-sensor/instrumentation/instasarama => ../
