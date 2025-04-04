Instana instrumentation for MongoDB driver
==========================================

This module contains instrumentation code for MongoDB clients written with [`go.mongodb.org/mongo-driver/v2`](https://pkg.go.dev/go.mongodb.org/mongo-driver/v2).

[![GoDoc](https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instamongo/v2)][godoc]

Installation
------------

To add the module to your `go.mod` file run the following command in your project directory:

```bash
$ go get github.com/instana/go-sensor/instrumentation/instamongo/v2
```

Usage
-----

`instamongo/v2` offers function wrappers for [`mongo.Connect()`][instamongo.Connect] that initialize and instrument an instance of `mongo.Client` by adding a command monitor for instrumentation to its existing configuration. This monitor then use the provided [`instana.Collector`][Collector] to trace the MongoDB queries made with this client instance.

```go
client, err := instamongo.Connect(
	collector, // an instance of instana.Collector used to instrument your application
	options.Client().ApplyURI("mongodb://localhost:27017"),
)
```

Any existing `options.CommandMonitor` provided with client options will be preserved and called for each command event.

See the [`instamongo` package documentation][godoc] for detailed examples.



[godoc]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instamongo/v2
[instamongo.Connect]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instamongo/v2#Connect
[Collector]: https://pkg.go.dev/github.com/instana/go-sensor#Collector

