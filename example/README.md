Examples
========

This folder contains examples of instrumenting the common use-cases with `github.com/instana/go-sensor`

* [Greeter](./http-database-greeter) - an instrumented HTTP server that queries a database
* [Doubler](./kafka-producer-consumer) - an instrumented Kafka processor, that counsumes and produces messages
* [event/](./example/event/) - Demonstrates usage of the Events API
* [autoprofile/](./example/autoprofile/) - Demonstrates usage of the AutoProfile™

For more up-to-date instrumentation code examples please consult the respective package documentation page:

* [`github.com/instana/go-sensor`](https://pkg.go.dev/github.com/instana/go-sensor?tab=doc#pkg-overview) - HTTP client and server instrumentation
* [`github.com/instana/go-sensor/instrumentation/instagrpc`](https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instagrpc?tab=doc#pkg-overview) - GRPC server and client instrumentation
* [`github.com/instana/go-sensor/instrumentation/instasarama`](https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instasarama?tab=doc#pkg-overview) - Kafka producer and consumer instrumentation for [`github.com/Shopify/sarama`](https://github.com/Shopify/sarama)
* [`github.com/instana/go-sensor/instrumentation/cloud.google.com/go/pubsub`](https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/cloud.google.com/go/pubsub?tab=doc#pkg-overview) - Google Cloud Pub/Sub producer and consumer instrumentation for [`cloud.google.com/go/pubsub`](https://cloud.google.com/go/pubsub)
* [`github.com/instana/go-sensor/instrumentation/cloud.google.com/go/storage`](https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/cloud.google.com/go/storage?tab=doc#pkg-overview) - Google Cloud Storage client instrumentation for [`cloud.google.com/go/storage`](https://cloud.google.com/go/storage)
