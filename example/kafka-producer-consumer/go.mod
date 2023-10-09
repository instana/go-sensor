module kafka-producer-consumer

go 1.15

require (
	github.com/Shopify/sarama v1.27.2
	github.com/instana/go-sensor v1.57.0
	github.com/instana/go-sensor/instrumentation/instasarama v1.10.0
	github.com/opentracing/opentracing-go v1.2.0
)

replace github.com/instana/go-sensor/instrumentation/instasarama => ../../instrumentation/instasarama
