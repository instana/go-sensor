Examples
========

This folder contains examples of instrumenting the common use-cases with `github.com/instana/go-sensor`

* [Greeter](./http-database-greeter) - an instrumented HTTP server that queries a database
* [Doubler](./kafka-producer-consumer) - an instrumented Kafka processor, that consumes and produces messages
* [Event](./event) - Demonstrates usage of the Events API
* [Autoprofile](./autoprofile) - Demonstrates usage of the AutoProfile™
* [OpenTracing](./opentracing) - an example of usage of Instana tracer in an app instrumented with OpenTracing
* [gRPC](./grpc-client-server) - an example of usage of Instana tracer in an app instrumented with gRPC
* [Gin](./gin) - an example of usage of Instana tracer instrumenting a [`Gin`](github.com/gin-gonic/gin) application
* [github.com/julienschmidt/httprouter](./httprouter) - an example of usage of Instana tracer to instrument a [`github.com/julienschmidt/httprouter`](https://github.com/julienschmidt/httprouter) router

For more up-to-date instrumentation code examples please consult the respective package documentation page:

* [`github.com/instana/go-sensor`](https://pkg.go.dev/github.com/instana/go-sensor?tab=doc#pkg-overview) - HTTP client and server instrumentation
* [`github.com/instana/go-sensor/instrumentation/instagrpc`](https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instagrpc?tab=doc#pkg-overview) - GRPC server and client instrumentation
* [`github.com/instana/go-sensor/instrumentation/instasarama`](https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instasarama?tab=doc#pkg-overview) - Kafka producer and consumer instrumentation for [`github.com/IBM/sarama`](https://github.com/IBM/sarama)
* [`github.com/instana/go-sensor/instrumentation/cloud.google.com/go/pubsub`](https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/cloud.google.com/go/pubsub?tab=doc#pkg-overview) - Google Cloud Pub/Sub producer and consumer instrumentation for [`cloud.google.com/go/pubsub`](https://cloud.google.com/go/pubsub)
* [`github.com/instana/go-sensor/instrumentation/cloud.google.com/go/storage`](https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/cloud.google.com/go/storage?tab=doc#pkg-overview) - Google Cloud Storage client instrumentation for [`cloud.google.com/go/storage`](https://cloud.google.com/go/storage)
* [`github.com/instana/go-sensor/instrumentation/instagin`](https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instagin?tab=doc#pkg-overview) - Instrumentation module for HTTP servers written using [`github.com/gin-gonic/gin`](https://github.com/gin-gonic/gin) framework
* [`github.com/instana/go-sensor/instrumentation/instamux`](https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instamux?tab=doc#pkg-overview) - Instrumentation module for HTTP servers written with [`github.com/gorilla/mux`](https://github.com/gorilla/mux) router
* [`github.com/instana/go-sensor/instrumentation/instahttprouter`](https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instahttprouter?tab=doc#pkg-overview) - Instrumentation module for HTTP servers written with [`github.com/julienschmidt/httprouter`](https://github.com/julienschmidt/httprouter) router
