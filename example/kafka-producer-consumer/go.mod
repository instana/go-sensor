module kafka-producer-consumer

go 1.15

require (
	github.com/IBM/sarama v1.41.0
	github.com/instana/go-sensor v1.55.2
	github.com/instana/go-sensor/instrumentation/instasarama v1.12.0
	github.com/opentracing/opentracing-go v1.2.0
)

replace github.com/instana/go-sensor/instrumentation/instasarama => ../../instrumentation/instasarama
