Instana instrumentation for MongoDB driver
==========================================

This module contains instrumentation code for MongoDB clients written with [`go.mongodb.org/mongo-driver`](https://go.mongodb.org/mongo-driver).

[![GoDoc](https://pkg.go.dev/badge/github.com/instana/go-sensor/instrumentation/instamongo)][godoc]

Installation
------------

To add the module to your `go.mod` file run the following command in your project directory:

```bash
$ go get github.com/instana/go-sensor/instrumentation/instamongo
```

Usage
-----

`instamongo` offers function wrappers for [`mongo.Connect()`][instamongo.Connect] and [`mongo.NewClient()`][instamongo.NewClient]
that initialize and instrument an instance of `mongo.Client` by adding a command monitor to its configuration. This monitor then
use provided `instana.Sensor` to trace MongoDB queries made with this client instance:

```go
client, err := instamongo.Connect(
	context.Background(),
	sensor, // an instance of instana.Sensor used to instrument your application
	options.Client().ApplyURI("mongodb://localhost:27017"),
)
```

Any existing `options.CommandMonitor` provided with client options will be preserved and called for each command event.

See the [`instamongo` package documentation][godoc] for detailed examples.



[godoc]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instamongo
[instamongo.Connect]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instamongo#Connect
[instamongo.NewClient]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instamongo#NewClient

<!---
Mandatory comment section for CI/CD !!
target-pkg-url: go.mongodb.org/mongo-driver
current-version: v1.7.2
--->
