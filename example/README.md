Examples
========

This folder contains examples of instrumenting the common use-cases with `github.com/instana/go-sensor`

* [autoprofile](./autoprofile) - Demonstrates the usage of autoprofiling using `go-sensor` SDK.
* [greeter](./http-database-greeter) - an instrumented application using `go-sensor` SDK with basic HTTP and database queries.
* [beego](./beego) - an example of instrumenting application using `instabeego`.
* [couchbase](./couchbase) - an example of instrumenting application using `instagocb`.
* [event](./event) - Demonstrates usage of the Events API in `go-sensor` SDK.
* [gin](./gin) - an example of instrumenting application using `instagin`.
* [gorm-sqlite](./gorm-sqlite) - an example of instrumenting a SQLite application using `instagorm`.
* [gorm-postgres](./gorm-postgres) - an example of instrumenting a Postgres application using `instagorm`.
* [graphQL](./graphql) - an example of instrumenting application using `instagraphql`.
* [gRPC](./grpc-client-server) - an example of usage of Instana tracer in an app instrumented with `instagrpc`.
* [httprouter](./httprouter) - an example of instrumenting application using `instahttprouter`.
* [openTracing](./opentracing) - an example of usage of Instana tracer in an app instrumented with OpenTracing.
* [sarama](./sarama) - an example of instrumenting application using `instasarama`.
* [sql-mysql](./sql-mysql) - an example of instrumenting a SQL application using `go-sensor` SDK.
* [sql-mysql-gin](./sql-mysql-gin) - an example of instrumenting a Gin web server that uses MySQL for database operations with the `go-sensor` SDK..
* [sql-redis](./sql-redis) - an example of instrumenting a Redis application using `go-sensor` SDK.
* [pgxv5](./pgxv5) - an example of instrumenting pgx v5 library using `go-sensor` SDK.
* [sql-opendb](./sql-opendb) - an example of instrumenting `sql.OpenDB` using `go-sensor` SDK.
* [instaredis-v2-example](./instaredis-v2-example) - an example of instrumenting a Redis application using `instaredis/v2` package.
* [fasthttp-example](./basic_usage_fasthttp) - an example of instrumenting `fasthttp` using `go-sensor` SDK.
* [mongo-driver-v2](./mongo-driver-v2) - an example of instrumenting `mongo-driver v2 library` using `go-sensor` SDK.
* [http_secret_matcher](./http_secret_matcher) - an example showing how to set up secret matchers using the Instana Go tracer(`go-sensor`) SDK.
* [docker-env](./docker-env) - an example showing how to set up secret matchers using the Instana Go tracer(`go-sensor`) SDK.
* [k8s-env](./k8s-env) - Demonstrates running a Go application with MySQL database in Kubernetes, monitored by Instana Go tracer(`go-sensor`) SDK.

For more up-to-date instrumentation code examples please consult the respective package documentation page:

* [`github.com/instana/go-sensor`](https://pkg.go.dev/github.com/instana/go-sensor?tab=doc#pkg-overview) - HTTP client and server instrumentation
* [`github.com/instana/go-sensor/instrumentation/instagrpc`](https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instagrpc?tab=doc#pkg-overview) - GRPC server and client instrumentation
* [`github.com/instana/go-sensor/instrumentation/instasarama`](https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instasarama?tab=doc#pkg-overview) - Kafka producer and consumer instrumentation for [`github.com/IBM/sarama`](https://github.com/IBM/sarama)
* [`github.com/instana/go-sensor/instrumentation/cloud.google.com/go/pubsub`](https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/cloud.google.com/go/pubsub?tab=doc#pkg-overview) - Google Cloud Pub/Sub producer and consumer instrumentation for [`cloud.google.com/go/pubsub`](https://cloud.google.com/go/pubsub)
* [`github.com/instana/go-sensor/instrumentation/cloud.google.com/go/storage`](https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/cloud.google.com/go/storage?tab=doc#pkg-overview) - Google Cloud Storage client instrumentation for [`cloud.google.com/go/storage`](https://cloud.google.com/go/storage)
* [`github.com/instana/go-sensor/instrumentation/instagin`](https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instagin?tab=doc#pkg-overview) - Instrumentation module for HTTP servers written using [`github.com/gin-gonic/gin`](https://github.com/gin-gonic/gin) framework
* [`github.com/instana/go-sensor/instrumentation/instamux`](https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instamux?tab=doc#pkg-overview) - Instrumentation module for HTTP servers written with [`github.com/gorilla/mux`](https://github.com/gorilla/mux) router
* [`github.com/instana/go-sensor/instrumentation/instahttprouter`](https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instahttprouter?tab=doc#pkg-overview) - Instrumentation module for HTTP servers written with [`github.com/julienschmidt/httprouter`](https://github.com/julienschmidt/httprouter) router
