Instana instrumentation for Google Cloud Pub/Sub
================================================

This module contains instrumentation code for [Google Cloud Pub/Sub][pubsub] producers and consumers that use `cloud.google.com/go/pubsub` library starting from `v1.3.1` and above.

[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)][godoc]

Installation
------------

To add the module to your `go.mod` file run the following command in your project directory:

```bash
$ go get github.com/instana/go-sensor/instrumentation/cloud.google.com/go/pubsub
```

Usage
-----

This module is a drop-in replacement for `cloud.google.com/go/pubsub` in common publisher/subscriber use cases. It provides
type aliases for `pubsub.Message`, `pubsub.PublishResult` and `pubsub.Subscription`, however if your code performs any
configuration calls to the Pub/Sub API, such as [updating the topic configuration](https://pkg.go.dev/cloud.google.com/go/pubsub#example-Topic.Update) or [creating subscriptions](https://pkg.go.dev/cloud.google.com/go/pubsub#example-Client.CreateSubscription), you might need add the named import for `cloud.google.com/go/pubsub` as well.

The instrumentation is implemented as a thin wrapper around service object methods and does not change their behavior. Thus,
any limitations/usage patterns/recommendations for the original method also apply to the wrapped one.

In most cases it is enough to change the import path from `cloud.google.com/go/pubsub` to `github.com/instana/go-sensor/instrumentation/cloud.google.com/go/pubsub` and add an instance of [`instana.Collector`][Collector] to the list of [`pubsub.NewClient()`][pubsub.NewClient] arguments to start tracing your communication over Google Cloud Pub/Sub with Instana.

### Instrumenting push delivery handlers

In case your service uses [push method](https://cloud.google.com/pubsub/docs/push) to receive Google Cloud Pub/Sub messages,
this instrumentation package provides an HTTP middleware [`pubsub.TracingHandlerFunc()`][pubsub.TracingHandlerFunc] that extracts
Instana trace context and ensures trace continuation:

```go
http.Handle("/", pubsub.TracingHandlerFunc(sensor, func (w http.ResponseWriter, req *http.Request) {
	// deserialize and process the message from req.Body
})
```

The [`pubsub.TracingHandlerFunc()`][pubsub.TracingHandlerFunc] is a complete replacement for a more generic HTTP instrumentation
done with [`instana.TracingHandlerFunc()`][instana.TracingHandlerFunc], and will use it for any non-Pub/Sub request, so there is
no need to use both middleware wrappers on one handler.

[godoc]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/cloud.google.com/go/pubsub
[pubsub]: https://cloud.google.com/pubsub
[instana.Sensor]: https://pkg.go.dev/github.com/instana/go-sensor#Sensor
[instana.TracingHandlerFunc]: https://pkg.go.dev/github.com/instana/go-sensor#TracingHandlerFunc
[pubsub.TracingHandlerFunc]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/cloud.google.com/go/pubsub#TracingHandlerFunc
[pubsub.NewClient]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/cloud.google.com/go/pubsub#NewClient
[Collector]: https://pkg.go.dev/github.com/instana/go-sensor#Collector
[InitCollector]: https://pkg.go.dev/github.com/instana/go-sensor#InitCollector
